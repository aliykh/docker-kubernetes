package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aliykh/docker-kubernetes/pkg/runtime"
	"github.com/thomaspoignant/go-feature-flag/exporter/fileexporter"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aliykh/docker-kubernetes/pkg/helpers"
	"github.com/aliykh/docker-kubernetes/pkg/redis"
	"github.com/aliykh/docker-kubernetes/pkg/server"
	ffclient "github.com/thomaspoignant/go-feature-flag"
)

type AppConfig struct {
	srv   *server.Config
	redis *redis.Config
}

var appCfg AppConfig

func init() {
	appCfg = AppConfig{
		srv: &server.Config{
			Name: "main-server",
			Port: 80,
		},
		redis: &redis.Config{
			Host:     "redis-server:6379",
			Password: "",
			DB:       0,
		},
	}
}

func main() {

	err := ffclient.Init(ffclient.Config{
		PollingInterval: 10 * time.Second,
		Logger:          log.New(os.Stdout, "", log.LstdFlags),
		Context:         context.Background(),
		Environment:     "production",
		Retrievers: []retriever.Retriever{
			&fileretriever.Retriever{
				Path: "features/feature.yaml",
			},
			//	&githubretriever.Retriever{
			//	RepositorySlug: "aliykh/docker-kubernetes",
			//	Branch:         "project/features",
			//	FilePath:       "features/remote_feature.yaml",
			//	GithubToken:    "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",
			//	Timeout:        5 * time.Second,
			//	},
		},
		DataExporter: ffclient.DataExporter{
			FlushInterval:    10 * time.Second,
			MaxEventInMemory: 1000,
			Exporter: &fileexporter.Exporter{
				OutputDir:   "tmp/",
				Format:      "csv",
				Filename:    "flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}",
				CsvTemplate: "{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};{{ .Value}};{{ .Default}};{{ .Source}}\n",
			},
		},
	})
	runtime.Require(err, "ffclient")

	defer ffclient.Close()

	redisC := redis.NewClient(appCfg.redis)
	err = helpers.Retry(func(attempt int, lastRetryCause string) error {
		return redisC.Ping()
	}, 2, time.Second*3, redis.RetryRedis)
	if err != nil {
		log.Printf("failed to connect to redis: %s", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /some-feature", customHandler("local-feature-flag"))
	mux.HandleFunc("POST /remote-some-feature", customHandler("remote-feature-flag"))

	mux.HandleFunc("GET /redis-ping", func(w http.ResponseWriter, r *http.Request) {
		err = redisC.Ping()
		if err != nil {
			log.Println(err)
			WriteJsonResponse(w, map[string]string{
				"message": "failed to ping redis",
			})
			return
		}

		WriteJsonResponse(w, map[string]string{
			"message": "redis is up and running!",
		})
	})

	srv := server.NewServer(appCfg.srv, mux)

	if err := srv.Start(); err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("Server [%s] started on port: [%d]\n", appCfg.srv.Name, appCfg.srv.Port)

	healthChecker := server.NewServer(&server.Config{
		Name: "health-checker",
		Port: 8080,
	}, http.HandlerFunc(HealthCheckHandler))

	if err := healthChecker.Start(); err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("Server [%s] started on port: [%d]\n", "Health check", 8080)

	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	go func() {
		sig := <-sigCh
		log.Printf("received signal: %s", sig)
		cancel()
	}()

	<-ctx.Done()
	if err := srv.Stop(); err != nil {
		log.Printf("Error on shutdown http: %s\n", err.Error())
		os.Exit(1)
	}

	if err := healthChecker.Stop(); err != nil {
		log.Printf("Error on shutdown health checker: %s\n", err.Error())
		os.Exit(1)
	}

	log.Println("shutdown successfully")
	os.Exit(0)
}

func WriteJsonResponse(w http.ResponseWriter, out any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(out); err != nil {
		log.Println(err)
	}
}

func customHandler(feature string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user1 := ffcontext.NewEvaluationContext(fmt.Sprintf("%x", time.Now().Nanosecond()))
		testFlag, err := ffclient.BoolVariation(feature, user1, false)
		runtime.Require(err, "Variation failed. please call init on ffclient")
		var msg string
		if testFlag {
			msg = "new feature"
		} else {
			msg = "old feature"
		}

		WriteJsonResponse(w, map[string]string{
			"message": msg,
		})
	}
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(w, "{ \"success\": true }"); err != nil {
		log.Printf("io.WriteString failed: %s", err)
	}
}

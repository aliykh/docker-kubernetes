package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aliykh/docker-kubernetes/pkg/helpers"
	"github.com/aliykh/docker-kubernetes/pkg/redis"
	"github.com/aliykh/docker-kubernetes/pkg/server"
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
			Port: 5001,
		},
		redis: &redis.Config{
			Host:     "redis-server:6379",
			Password: "",
			DB:       0,
		},
	}
}

func main() {
	redisC := redis.NewClient(appCfg.redis)
	err := helpers.Retry(func(attempt int, lastRetryCause string) error {
		return redisC.Ping()
	}, 5, time.Second*3, redis.RetryRedis)
	if err != nil {
		log.Printf("failed to connect to redis: %s", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health-check", HealthCheck)
	mux.HandleFunc("GET /hello-world", HelloWorldHandler)
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
		log.Fatal(err)
	}

	log.Println("shutdown successfully")
}

func WriteJsonResponse(w http.ResponseWriter, out any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(out); err != nil {
		log.Println(err)
	}
}

func HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	WriteJsonResponse(w, map[string]string{
		"message": "Hello World!",
	})
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type Config struct {
	Name string
	Port int
}

func (c Config) getAddr() string {
	return fmt.Sprintf(":%d", c.Port)
}

type Server struct {
	server *http.Server
	name   string
}

func NewServer(cfg *Config, handler http.Handler) *Server {
	return &Server{
		server: &http.Server{
			Handler: handler,
			Addr:    cfg.getAddr(),
			// TODO: add additional parameters like timouts later
		},
		name: cfg.Name,
	}
}

func (sv *Server) Start() error {
	listener, err := net.Listen("tcp", sv.server.Addr)
	if err != nil {
		return errors.Wrap(err, "Server.Start")
	}

	go func() {
		if err := sv.server.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Serve: %v", err)
		}
	}()

	return nil
}

func (sv *Server) Stop() error {
	if sv.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if err := sv.server.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

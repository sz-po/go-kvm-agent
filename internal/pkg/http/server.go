package http

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ServerOpt func(router chi.Router)

type ServerConfig struct {
	ListenHost string `json:"ListenHost" default:"0.0.0.0"`
	ListenPort int    `json:"listenPort" default:"8080"`
}

type ServerHandler func(router chi.Router)

type ServerMiddleware interface {
	Handle(http.Handler) http.Handler
}

func WithHandler(handler ServerHandler) ServerOpt {
	return func(router chi.Router) {
		handler(router)
		slog.Debug("Registered HTTP server handler.")
	}
}

func WithMiddleware(middleware ServerMiddleware) ServerOpt {
	return func(router chi.Router) {
		router.Use(middleware.Handle)
	}
}

func Listen(ctx context.Context, config ServerConfig, opts ...ServerOpt) error {
	listenPath := fmt.Sprintf("%s:%d", config.ListenHost, config.ListenPort)

	listener, err := net.Listen("tcp", listenPath)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", listenPath, err)
	}

	router := chi.NewRouter()

	for _, opt := range opts {
		opt(router)
	}

	go http.Serve(listener, router)

	slog.Info("API HTTP server listening.",
		slog.String("listenHost", config.ListenHost),
		slog.Int("listenPort", config.ListenPort),
	)

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	return nil
}

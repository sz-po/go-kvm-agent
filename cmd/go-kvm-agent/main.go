package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/alecthomas/kong"
	go_kvm_agent "github.com/szymonpodeszwa/go-kvm-agent/internal/app/go-kvm-agent"
)

func main() {
	var config go_kvm_agent.Config
	kong.Parse(&config)

	// Setup slog
	logger, err := setupLogger(config.Log.Level, config.Log.Format)
	if err != nil {
		slog.Error("Failed to setup logger.", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.SetDefault(logger)

	// Context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		slog.Info("Starting application.")
		if err := go_kvm_agent.Start(config, &wg, ctx); err != nil {
			slog.Error("Application start error.", slog.String("error", err.Error()))
		}
		slog.Info("Application started.")
	}()

	<-sigChan
	wg.Done()
	slog.Info("Shutdown signal received, stopping application.")
	cancel()

	wg.Wait()
	slog.Info("Application stopped.")
}

func setupLogger(level, format string) (*slog.Logger, error) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	default:
		return nil, fmt.Errorf("invalid log level: %s", level)
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: logLevel}

	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		return nil, fmt.Errorf("invalid log format: %s", format)
	}

	return slog.New(handler), nil
}

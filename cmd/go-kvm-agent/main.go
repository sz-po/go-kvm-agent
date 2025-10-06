package main

import (
	"context"
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
	logger := setupLogger(config.LogLevel, config.LogFormat)
	slog.SetDefault(logger)

	// Context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// WaitGroup
	var wg sync.WaitGroup

	// Start application
	wg.Add(1)
	go func() {
		if err := go_kvm_agent.Start(config, &wg, ctx); err != nil {
			slog.Error("application error", "error", err)
		}
	}()

	// Wait for signal
	<-sigChan
	slog.Info("shutdown signal received, stopping...")
	cancel()

	// Wait for graceful shutdown
	wg.Wait()
	slog.Info("application stopped")
}

func setupLogger(level, format string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	default:
		logLevel = slog.LevelInfo
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: logLevel}

	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

package main

import (
	"fmt"
	"log/slog"
	"os"
)

func createLogger(parameters LogParameters) (*slog.Logger, error) {
	var logLevel slog.Level
	switch parameters.Level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	default:
		return nil, fmt.Errorf("invalid log level: %s", parameters.Level)
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: logLevel}

	switch parameters.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		return nil, fmt.Errorf("invalid log format: %s", parameters.Format)
	}

	logger := slog.New(handler)

	logger.Info("Logger created.",
		slog.String("logLevel", parameters.Level),
		slog.String("logFormat", parameters.Format),
	)

	return logger, nil
}

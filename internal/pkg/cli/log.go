package cli

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

type LogConfig struct {
	Level  string `help:"Log level." enum:"debug,info" default:"info"`
	Format string `help:"Log format." enum:"text,json" default:"text"`
	Quiet  bool   `help:"Suppress all log output." default:"false"`
}

type SupportLogConfig interface {
	GetLogConfig() LogConfig
}

type LogConfigHelper struct {
	Log LogConfig `kong:"embed,prefix='log-',group='Logging configuration'"`
}

func (helper LogConfigHelper) GetLogConfig() LogConfig {
	return helper.Log
}

func CreateLogger(config LogConfig) (*slog.Logger, error) {
	var logLevel slog.Level
	switch config.Level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	default:
		return nil, fmt.Errorf("invalid log level: %s", config.Level)
	}

	output := io.Writer(os.Stdout)
	if config.Quiet {
		output = io.Discard
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: logLevel}

	switch config.Format {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	case "text":
		handler = slog.NewTextHandler(output, opts)
	default:
		return nil, fmt.Errorf("invalid log format: %s", config.Format)
	}

	logger := slog.New(handler)

	if !config.Quiet {
		logger.Info("Logger created.",
			slog.String("logLevel", config.Level),
			slog.String("logFormat", config.Format),
		)
	}

	return logger, nil
}

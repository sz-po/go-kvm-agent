package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"

	"github.com/alecthomas/kong"
	go_kvm_agent "github.com/szymonpodeszwa/go-kvm-agent/internal/app/go-kvm-agent"
)

const (
	ExitOk                 = 0
	ExitErrParseParameters = 1
	ExitErrorLoggerCreate  = 2
	ExitErrorConfigLoad    = 3
)

func main() {
	var parameters Parameters
	if ctx := kong.Parse(&parameters); ctx.Error != nil {
		fmt.Println(ctx.Error)
		os.Exit(ExitErrParseParameters)
	}

	logger, err := createLogger(parameters.Log)
	if err != nil {
		slog.Error("Failed to create logger.", slog.String("error", err.Error()))
		os.Exit(ExitErrorLoggerCreate)
	}
	slog.SetDefault(logger)

	config, err := loadConfigFromPath(parameters.ConfigPath)
	if err != nil {
		slog.Error("Failed to load config from path.", slog.String("configPath", parameters.ConfigPath))
		os.Exit(ExitErrorConfigLoad)
	}

	if parameters.Machine.ConfigPath != nil {
		machinesConfig, err := loadMachineConfigFromPath(*parameters.Machine.ConfigPath)
		if err != nil {
			slog.Error("Failed to load machine config from path.",
				slog.String("configPath", *parameters.Machine.ConfigPath),
				slog.String("error", err.Error()),
			)
		}
		for _, machineConfig := range machinesConfig {
			config.Machines = append(config.Machines, machineConfig)
		}
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var wg sync.WaitGroup

	slog.Info("Starting application.")
	if err := go_kvm_agent.Start(ctx, &wg, *config); err != nil {
		slog.Error("Application start error.", slog.String("error", err.Error()))
		return
	}

	slog.Info("Application started.")

	wg.Wait()
	slog.Info("Application stopped.")

	os.Exit(ExitOk)
}

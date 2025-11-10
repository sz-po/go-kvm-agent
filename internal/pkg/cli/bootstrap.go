package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"

	"github.com/alecthomas/kong"
)

type BootstrapOptions struct {
	kongOptions []kong.Option
}

type BootstrapOpt func(options *BootstrapOptions)

func WithApplicationName(name string) BootstrapOpt {
	return func(options *BootstrapOptions) {
		options.kongOptions = append(options.kongOptions, kong.Name(name))
	}
}

func WithKongOptions(opts ...kong.Option) BootstrapOpt {
	return func(options *BootstrapOptions) {
		options.kongOptions = append(options.kongOptions, opts...)
	}
}

const (
	ExitOk                 = 0
	ExitErrParseParameters = 1
	ExitErrorLoggerCreate  = 2
	ExitErrorStart         = 3
)

type DaemonStartHandler[CONFIG Config] func(ctx context.Context, wg *sync.WaitGroup, config CONFIG) error

func BootstrapDaemon[CONFIG Config](startHandler DaemonStartHandler[CONFIG], opts ...BootstrapOpt) {
	options := &BootstrapOptions{}
	for _, opt := range opts {
		opt(options)
	}

	kongOptions := append([]kong.Option{
		kong.UsageOnError(),
	}, options.kongOptions...)

	var config CONFIG
	if ctx := kong.Parse(&config, kongOptions...); ctx.Error != nil {
		fmt.Println(ctx.Error)
		os.Exit(ExitErrParseParameters)
	}

	ctx, cancelSignal := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelSignal()

	logger, err := CreateLogger(config.GetLogConfig())
	if err != nil {
		fmt.Println(fmt.Errorf("create logger: %w", err))
		os.Exit(ExitErrorLoggerCreate)
	}

	slog.SetDefault(logger)

	wg := &sync.WaitGroup{}

	if err := startHandler(ctx, wg, config); err != nil {
		logger.Error("Application start error.", slog.String("error", err.Error()))
		os.Exit(ExitErrorStart)
	}

	slog.Info("Application started.")

	wg.Wait()

	slog.Info("Application finished.")

	os.Exit(ExitOk)
}

func BootstrapCommands[CONFIG Config, COMMANDS any](opts ...BootstrapOpt) {
	wg := &sync.WaitGroup{}

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	ctx, cancelSignal := signal.NotifyContext(ctx, os.Interrupt)
	defer cancelSignal()

	options := &BootstrapOptions{}
	for _, opt := range opts {
		opt(options)
	}

	kongOptions := append([]kong.Option{
		kong.UsageOnError(),
		kong.Bind(wg),
		kong.BindTo(ctx, (*context.Context)(nil)),
		kong.Help(FilteredHelpPrinter),
	}, options.kongOptions...)

	var cli struct {
		Config   CONFIG   `embed:"true"`
		Commands COMMANDS `embed:"true"`
	}

	runtime := kong.Parse(&cli, kongOptions...)
	if runtime.Error != nil {
		fmt.Println(runtime.Error)
		os.Exit(ExitErrParseParameters)
	}
	runtime.Bind(cli.Config.GetTransportConfig())

	logger, err := CreateLogger(cli.Config.GetLogConfig())
	if err != nil {
		fmt.Println(fmt.Errorf("create logger: %w", err))
		os.Exit(ExitErrorLoggerCreate)
	}
	runtime.Bind(logger)

	slog.SetDefault(logger)

	err = runtime.Run()
	if err != nil {
		logger.Error("Application start error.", slog.String("error", err.Error()))
		os.Exit(ExitErrorStart)
	}

	cancelFn()

	wg.Wait()

	os.Exit(ExitOk)
}

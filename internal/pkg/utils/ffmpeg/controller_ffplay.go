package ffmpeg

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/process"
)

type ffplayControllerOptions struct {
	executablePath string
	logger         *slog.Logger
}

func defaultFFplayControllerOptions() ffplayControllerOptions {
	return ffplayControllerOptions{
		executablePath: "/usr/local/bin/ffplay",
		logger:         slog.New(slog.DiscardHandler),
	}
}

type FFplayControlerOpt func(*ffplayControllerOptions) error

type FFplayStatus struct {
	FrameDrops    int64 `json:"frameDrops,omitempty"`
	VideoQueue    int64 `json:"videoQueue,omitempty"`
	AudioQueue    int64 `json:"audioQueue,omitempty"`
	SubtitleQueue int64 `json:"subtitleQueue,omitempty"`
}

const ffplayStableDuration = 5 * time.Second

type FFplayController struct {
	options *ffplayControllerOptions
	process process.Supervisor

	currentInput         Input
	currentConfiguration Configuration

	currentStatus     *FFplayStatus
	currentStatusLock sync.RWMutex

	specMutex            sync.Mutex
	reloadStableDuration time.Duration
	logger               *slog.Logger
}

func WithFFplayLogger(logger *slog.Logger) FFplayControlerOpt {
	return func(options *ffplayControllerOptions) error {
		options.logger = logger
		return nil
	}
}

func NewFFplayController(input Input, configuration Configuration, opts ...FFplayControlerOpt) (*FFplayController, error) {
	if input == nil {
		return nil, fmt.Errorf("create ffplay controller: %w", ErrFFplayNilInput)
	}

	if configuration == nil {
		return nil, fmt.Errorf("create ffplay controller: %w", ErrFFplayNilConfiguration)
	}

	options := defaultFFplayControllerOptions()
	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return nil, err
		}
	}

	controller := &FFplayController{
		options:              &options,
		currentInput:         input,
		currentConfiguration: configuration,
		currentStatus:        &FFplayStatus{},
		currentStatusLock:    sync.RWMutex{},
		specMutex:            sync.Mutex{},
		reloadStableDuration: ffplayStableDuration,
		logger:               options.logger,
	}

	specification := controller.buildSpecification(controller.currentInput, controller.currentConfiguration)

	controller.process = process.SuperviseLocal(specification, process.RestartPolicy{
		Enabled:      true,
		MaxAttempts:  10,
		Strategy:     process.StrategyExponential,
		InitialDelay: time.Second,
		MaxDelay:     time.Second * 5,
		ResetWindow:  time.Second * 2,
	})

	controller.logger.Debug("FFplay prepared for start.",
		slog.String("ffplayArguments", fmt.Sprintf("%v", specification.Arguments)),
	)

	return controller, nil
}

func (controller *FFplayController) Start(ctx context.Context) error {
	if err := controller.process.Start(ctx, controller.reloadStableDuration); err != nil {
		return fmt.Errorf("error starting ffplay: %w", err)
	}

	go func() {
		scanner := bufio.NewScanner(controller.process.Stderr())
		for scanner.Scan() {
			controller.parseStatusLine(scanner.Text())
		}
	}()

	controller.logger.Debug("FFplay started.")

	return nil
}

func (controller *FFplayController) Stop(ctx context.Context) error {
	err := controller.process.Stop(ctx)

	controller.logger.Debug("FFplay stopped.")

	if err == nil {
		return nil
	}

	if errors.Is(err, process.KilledError{}) {
		return nil
	}

	return fmt.Errorf("error stopping ffplay: %w", err)
}

func (controller *FFplayController) SetInputWithConfiguration(ctx context.Context, input Input, configuration Configuration) error {
	if input == nil {
		return fmt.Errorf("set input: %w", ErrFFplayNilInput)
	}

	controller.specMutex.Lock()
	controller.currentInput = input
	controller.currentConfiguration = configuration
	controller.specMutex.Unlock()

	specification := controller.buildSpecification(input, configuration)

	controller.logger.Debug("Reloading ffplay with updated input.",
		slog.String("ffplayArguments", fmt.Sprintf("%v", specification.Arguments)),
	)

	if err := controller.process.ReloadWithSpecification(ctx, specification, controller.reloadStableDuration); err != nil {
		return fmt.Errorf("reload ffplay with updated input: %w", err)
	}

	return nil
}

func (controller *FFplayController) GetStatus() FFplayStatus {
	controller.currentStatusLock.RLock()
	defer controller.currentStatusLock.RUnlock()

	return *controller.currentStatus
}

func (controller *FFplayController) GetStdin() io.Writer {
	return controller.process.Stdin()
}

func (controller *FFplayController) GetStdout() io.Reader {
	return controller.process.Stdout()
}

func (controller *FFplayController) buildSpecification(input Input, configuration Configuration) process.Specification {
	inputParameters := []string{}
	if input != nil {
		inputParameters = slices.Clone(input.Parameters())
	}

	configurationParameters := []string{}
	if configuration != nil {
		configurationParameters = slices.Clone(configuration.Parameters())
	}

	arguments := slices.Concat(
		[]string{
			"-hide_banner",
		},
		inputParameters,
		[]string{
			"-stats",
		},
		configurationParameters,
	)

	return process.Specification{
		ExecutablePath: controller.options.executablePath,
		Arguments:      arguments,
	}
}

var ffplayStatsRegex = regexp.MustCompile(`fd=\s*(\d+)|aq=\s*(\d+)KB|vq=\s*(\d+)KB|sq=\s*(\d+)B`)

func (controller *FFplayController) parseStatusLine(statusLine string) {
	matches := ffplayStatsRegex.FindAllStringSubmatch(statusLine, -1)
	if len(matches) == 0 {
		return
	}

	controller.currentStatusLock.Lock()
	defer controller.currentStatusLock.Unlock()

	for _, match := range matches {
		if match[1] != "" {
			frameDrops, err := strconv.ParseInt(match[1], 10, 64)
			if err != nil {
				continue
			}
			controller.currentStatus.FrameDrops = frameDrops
		}

		if match[2] != "" {
			audioQueue, err := strconv.ParseInt(match[2], 10, 64)
			if err != nil {
				continue
			}
			controller.currentStatus.AudioQueue = audioQueue
		}

		if match[3] != "" {
			videoQueue, err := strconv.ParseInt(match[3], 10, 64)
			if err != nil {
				continue
			}
			controller.currentStatus.VideoQueue = videoQueue
		}

		if match[4] != "" {
			subtitleQueue, err := strconv.ParseInt(match[4], 10, 64)
			if err != nil {
				continue
			}
			controller.currentStatus.SubtitleQueue = subtitleQueue
		}
	}
}

var (
	ErrFFplayNilInput         = errors.New("ffplay input cannot be nil")
	ErrFFplayNilConfiguration = errors.New("ffplay configuration cannot be nil")
)

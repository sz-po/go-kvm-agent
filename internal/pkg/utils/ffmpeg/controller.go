package ffmpeg

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/process"
)

type controllerOptions struct {
	executablePath string
	logger         *slog.Logger
}

func defaultControllerOptions() controllerOptions {
	return controllerOptions{
		executablePath: "/usr/local/bin/ffmpeg",
		logger:         slog.New(slog.DiscardHandler),
	}
}

type ControlerOpt func(*controllerOptions) error

type Input interface {
	Parameters() []string
}

type Output interface {
	Parameters() []string
}

type Configuration interface {
	Parameters() []string
}

type Status struct {
	DroppedFrames int64   `json:"droppedFrames,omitempty"`
	Speed         float64 `json:"speed,omitempty"`
	FrameRate     float64 `json:"frameRate,omitempty"`
	TotalSize     int64   `json:"totalSize,omitempty"`
}

type Controller struct {
	options *controllerOptions
	process process.Supervisor

	currentStatus     *Status
	currentStatusLock sync.RWMutex

	logger *slog.Logger
}

func WithLogger(logger *slog.Logger) ControlerOpt {
	return func(options *controllerOptions) error {
		options.logger = logger
		return nil
	}
}

func NewController(input Input, output Output, configuration Configuration, opts ...ControlerOpt) (*Controller, error) {
	options := defaultControllerOptions()
	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return nil, err
		}
	}

	arguments := slices.Concat(
		input.Parameters(),
		[]string{
			"-progress",
			"pipe:2",
		},
		configuration.Parameters(),
		output.Parameters(),
	)

	controller := &Controller{
		options: &options,
		process: process.SuperviseLocal(process.Specification{
			ExecutablePath: options.executablePath,
			Arguments:      arguments,
		}, process.RestartPolicy{
			Enabled:      true,
			MaxAttempts:  10,
			Strategy:     process.StrategyExponential,
			InitialDelay: time.Second,
			MaxDelay:     time.Second * 10,
			ResetWindow:  time.Second * 5,
		}),
		currentStatus:     &Status{},
		currentStatusLock: sync.RWMutex{},
		logger:            options.logger,
	}

	controller.logger.Debug("FFmpeg prepared for start.",
		slog.String("ffmpegArguments", strings.Join(arguments, " ")),
	)

	return controller, nil
}

func (controller *Controller) Start(ctx context.Context) error {
	err := controller.process.Start(ctx, time.Second*5)
	if err != nil {
		return fmt.Errorf("error starting ffmpeg: %w", err)
	}

	go func() {
		scanner := bufio.NewScanner(controller.process.Stderr())
		for scanner.Scan() {
			controller.parseStatusLine(scanner.Text())
		}
	}()

	controller.logger.Debug("FFmpeg started.")

	return nil
}

func (controller *Controller) Stop(ctx context.Context) error {
	err := controller.process.Stop(ctx)

	controller.logger.Debug("FFmpeg stopped.")

	if errors.Is(err, process.KilledError{}) {
		return nil
	} else {
		return fmt.Errorf("error stopping ffmpeg: %w", err)
	}
}

func (controller *Controller) GetStatus() Status {
	controller.currentStatusLock.RLock()
	defer controller.currentStatusLock.RUnlock()

	return *controller.currentStatus
}

func (controller *Controller) parseStatusLine(statusLine string) {
	statusLineParts := strings.Split(statusLine, "=")
	if len(statusLineParts) != 2 {
		return
	}

	key, value := statusLineParts[0], statusLineParts[1]

	controller.currentStatusLock.Lock()
	defer controller.currentStatusLock.Unlock()

	switch key {
	case "drop_frames":
		droppedFrames, err := strconv.ParseInt(statusLineParts[1], 10, 64)
		if err != nil {
			return
		}
		controller.currentStatus.DroppedFrames = droppedFrames
	case "speed":
		speed, err := strconv.ParseFloat(strings.TrimRight(value, "x"), 64)
		if err != nil {
			return
		}
		controller.currentStatus.Speed = speed
	case "fps":
		fps, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return
		}
		controller.currentStatus.FrameRate = fps
	case "total_size":
		totalSize, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return
		}
		controller.currentStatus.TotalSize = totalSize
	default:
		return
	}
}

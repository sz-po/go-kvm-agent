package mpv

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/go-cmd/cmd"
	"golang.org/x/sys/unix"
)

const RestartDelay = 1 * time.Second

type ControllerOpt func(*Controller)

// Controller controls MPV instance. It watches config parameters that may change in runtime and restart instance if
// necessary. It also watches process and ensure it is running.
type Controller struct {
	mpvPath            string
	parameters         map[ParameterKey]string
	requiredParameters []ParameterKey

	mpvCmd       *cmd.Cmd
	cancel       context.CancelCauseFunc
	reloadSignal chan struct{}
	wg           *sync.WaitGroup
	logger       *slog.Logger

	tmpDir        string
	videoPipePath string
	videoPipeFile *os.File
	pipeMutex     sync.Mutex
}

func WithStaticParameters(parameters ...RenderedParameter) ControllerOpt {
	return func(controller *Controller) {
		for _, parameter := range parameters {
			controller.parameters[parameter.key] = parameter.rendered
		}
	}
}

func WithRequiredParameters(parameters ...Parameter) ControllerOpt {
	return func(controller *Controller) {
		for _, parameter := range parameters {
			controller.requiredParameters = append(controller.requiredParameters, parameter.GetKey())
		}
	}
}

func NewController(opts ...ControllerOpt) (*Controller, error) {
	controller := &Controller{
		mpvPath:            "/usr/local/bin/mpv",
		parameters:         map[ParameterKey]string{},
		requiredParameters: []ParameterKey{},
		reloadSignal:       make(chan struct{}, 1),
		wg:                 &sync.WaitGroup{},
		logger: slog.Default().With(
			slog.String("component", "mpv-controller"),
		),
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller, nil
}

func (controller *Controller) SetParameters(parameters ...RenderedParameter) {
	for _, parameter := range parameters {
		controller.parameters[parameter.key] = parameter.rendered
	}

	controller.Reload()
}

func (controller *Controller) RenderFrame(frame []byte) error {
	controller.pipeMutex.Lock()
	defer controller.pipeMutex.Unlock()

	if controller.videoPipeFile == nil {
		return fmt.Errorf("video pipe not open")
	}

	if _, err := controller.videoPipeFile.Write(frame); err != nil {
		return fmt.Errorf("write frame to video pipe: %w", err)
	}

	return nil
}

func (controller *Controller) Reload() {
	if controller.cancel == nil {
		controller.logger.Warn("Cannot reload: controller not started.")
		return
	}

	controller.logger.Info("Triggering MPV reload.")

	select {
	case controller.reloadSignal <- struct{}{}:
	default:
		controller.logger.Warn("Reload already pending, skipping.")
	}
}

func (controller *Controller) Start(ctx context.Context) error {
	if controller.cancel != nil {
		return ErrControllerAlreadyStarted
	}

	controller.logger.Debug("Starting MPV controller.")

	tmpDir, err := os.MkdirTemp("/tmp", "mpv-controller-")
	if err != nil {
		return fmt.Errorf("create temp directory: %w", err)
	}
	controller.tmpDir = tmpDir

	controller.videoPipePath = filepath.Join(tmpDir, "video.pipe")

	if err := unix.Mkfifo(controller.videoPipePath, 0600); err != nil {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("create video pipe: %w", err)
	}

	controller.logger.Debug("Created named pipe.", slog.String("videoPipePath", controller.videoPipePath))

	ctx, cancel := context.WithCancelCause(ctx)
	controller.cancel = cancel

	controller.wg.Add(1)
	go controller.controlLoop(ctx)

	return nil
}

func (controller *Controller) Stop() error {
	if controller.cancel == nil {
		return ErrControllerNotStarted
	}

	controller.logger.Debug("Stopping MPV controller.")

	// Cancel context to stop controlLoop
	controller.cancel(nil)
	controller.cancel = nil

	// Wait for controlLoop to finish (it will stop mpvCmd and close pipe)
	controller.wg.Wait()

	// Remove temp directory
	if controller.tmpDir != "" {
		if err := os.RemoveAll(controller.tmpDir); err != nil {
			controller.logger.Warn("Failed to remove temp directory.",
				slog.String("path", controller.tmpDir),
				slog.String("error", err.Error()),
			)
		}
	}

	controller.logger.Debug("MPV controller stopped successfully.")

	return nil
}

func (controller *Controller) controlLoop(ctx context.Context) {
	defer controller.wg.Done()

	controller.logger.Debug("Starting MPV supervision loop.")

	for {
		// Check if we can start MPV (all required parameters set)
		args, err := controller.renderArguments()
		if err != nil {
			controller.logger.Debug("Cannot start MPV: missing required parameters.",
				slog.String("error", err.Error()),
			)

			// Wait for reload signal or context cancellation
			select {
			case <-ctx.Done():
				controller.logger.Debug("MPV supervision loop stopped.")
				return
			case <-controller.reloadSignal:
				controller.logger.Debug("Reload signal received, checking parameters again.")
				continue
			}
		}

		// Start MPV process
		controller.logger.Debug("Starting MPV process.",
			slog.String("executable", controller.mpvPath),
			slog.Any("args", args),
		)

		mpvCmd := cmd.NewCmdOptions(cmd.Options{
			Buffered:  false,
			Streaming: true,
		}, controller.mpvPath, args...)
		statusChan := mpvCmd.Start()
		controller.mpvCmd = mpvCmd

		// Open video pipe for writing
		videoPipeFile, err := os.OpenFile(controller.videoPipePath, os.O_WRONLY, 0600)
		if err != nil {
			controller.logger.Error("Failed to open video pipe.",
				slog.String("error", err.Error()),
			)
			mpvCmd.Stop()
			// Wait before retry
			select {
			case <-ctx.Done():
				controller.logger.Debug("MPV supervision loop stopped.")
				return
			case <-time.After(RestartDelay):
				continue
			}
		}

		controller.pipeMutex.Lock()
		controller.videoPipeFile = videoPipeFile
		controller.pipeMutex.Unlock()

		controller.logger.Debug("Opened video pipe for writing.")

		// Monitor process
		select {
		case <-ctx.Done():
			// Context cancelled - stop MPV and exit
			controller.logger.Debug("Stopping MPV process.")
			controller.closePipe()
			mpvCmd.Stop()
			controller.logger.Debug("MPV supervision loop stopped.")
			return

		case <-controller.reloadSignal:
			// Reload requested - stop current MPV and restart with new config
			controller.logger.Debug("Reload signal received, restarting MPV.")
			controller.closePipe()
			mpvCmd.Stop()
			continue

		case status := <-statusChan:
			// Process exited - close pipe first
			controller.closePipe()

			// Process exited
			if len(status.Stdout) > 0 {
				for _, line := range status.Stdout {
					controller.logger.Debug("MPV stdout.", slog.String("line", line))
				}
			}
			if len(status.Stderr) > 0 {
				for _, line := range status.Stderr {
					controller.logger.Debug("MPV stderr.", slog.String("line", line))
				}
			}

			controller.logger.Warn("MPV process exited.",
				slog.Int("exitCode", status.Exit),
				slog.Duration("restartDelay", RestartDelay),
			)

			// Wait before restart
			select {
			case <-ctx.Done():
				controller.logger.Debug("MPV supervision loop stopped during restart delay.")
				return
			case <-controller.reloadSignal:
				controller.logger.Debug("Reload signal received during restart delay.")
				continue
			case <-time.After(RestartDelay):
				controller.logger.Debug("Restarting MPV process.")
				continue
			}
		}
	}
}

func (controller *Controller) closePipe() {
	controller.pipeMutex.Lock()
	defer controller.pipeMutex.Unlock()

	if controller.videoPipeFile != nil {
		if err := controller.videoPipeFile.Close(); err != nil {
			controller.logger.Warn("Failed to close video pipe.",
				slog.String("error", err.Error()),
			)
		}
		controller.videoPipeFile = nil
	}
}

func (controller *Controller) renderArguments() ([]string, error) {
	for _, requiredParameter := range controller.requiredParameters {
		if _, ok := controller.parameters[requiredParameter]; !ok {
			return nil, fmt.Errorf("%w: %s", ErrMissingRequiredParameter, requiredParameter)
		}
	}

	args := slices.Collect(maps.Values(controller.parameters))

	// Add video pipe path as input file (last argument)
	args = append(args, controller.videoPipePath)

	return args, nil
}

var ErrMissingRequiredParameter = errors.New("missing required parameter")
var ErrControllerNotStarted = errors.New("controller not started")
var ErrControllerAlreadyStarted = errors.New("controller already started")

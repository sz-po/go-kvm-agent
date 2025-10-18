package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/ffmpeg"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/formats/ppm"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

// TODO: refine coments in this file
// TODO: change ffmpeg output read to new (not implemented yet) output_reader

// DisplaySourceDriver is the peripheral driver identifier for FFMPEG-based display sources.
const DisplaySourceDriver = peripheralSDK.PeripheralDriver("ffmpeg/display-source")

type DisplaySourceInputConfig struct {
	MessageBoard *DisplaySourceMessageBoardInputConfig `json:"messageBoard"`
}

type DisplaySourceInput interface {
	ffmpeg.Input
	GetPixelFormat() peripheralSDK.DisplayPixelFormat
	GetDisplayMode() (peripheralSDK.DisplayMode, error)
}

// DisplaySourceConfig holds configuration for creating an FFMPEG display source.
type DisplaySourceConfig struct {
	Executable struct {
		Path *string `json:"path"`
	} `json:"executable"`
	Input DisplaySourceInputConfig `json:"input"`
}

type DisplaySourceOptions struct {
	logger *slog.Logger
}

func defaultDisplaySourceOptions() *DisplaySourceOptions {
	return &DisplaySourceOptions{
		logger: slog.New(slog.DiscardHandler),
	}
}

type DisplaySourceOpt func(options *DisplaySourceOptions)

// DisplaySource is a mpv implementation of a display source using FFMPEG.
type DisplaySource struct {
	id   peripheralSDK.PeripheralId
	name peripheralSDK.PeripheralName

	lifecycleCtx    context.Context
	lifecycleCancel context.CancelFunc

	ffmpegController *ffmpeg.FFmpegController

	frameBuffer     *peripheralSDK.DisplayFrameBuffer
	frameBufferLock sync.RWMutex

	displayMode peripheralSDK.DisplayMode
	pixelFormat peripheralSDK.DisplayPixelFormat

	input DisplaySourceInput

	metrics     *peripheralSDK.DisplaySourceMetrics
	metricsLock sync.RWMutex

	logger *slog.Logger
}

var _ peripheralSDK.DisplaySource = (*DisplaySource)(nil)

func WithDisplaySourceLogger(logger *slog.Logger) DisplaySourceOpt {
	return func(options *DisplaySourceOptions) {
		options.logger = logger
	}
}

// NewDisplaySource creates a new FFMPEG display source from the provided configuration.
func NewDisplaySource(ctx context.Context, config DisplaySourceConfig, name peripheralSDK.PeripheralName, opts ...DisplaySourceOpt) (*DisplaySource, error) {
	options := defaultDisplaySourceOptions()
	for _, opt := range opts {
		opt(options)
	}

	id := peripheralSDK.CreatePeripheralRandomId("ffmpeg-display-source")

	logger := options.logger.With(slog.String("peripheralId", id.String()))

	var ffmpegInput DisplaySourceInput

	if config.Input.MessageBoard != nil {
		ffmpegInput = NewDisplaySourceMessageBoardInput(*config.Input.MessageBoard)
	} else {
		return nil, ErrDisplaySourceMissingInput
	}

	displayMode, err := ffmpegInput.GetDisplayMode()
	if err != nil {
		return nil, fmt.Errorf("error getting display mode from input: %w", err)
	}

	pixelFormat := ffmpegInput.GetPixelFormat()

	ffmpegOutput := ffmpeg.NewOutputStdout()

	ffmpegConfig := ffmpeg.RawConfiguration{
		"-f",
		"image2pipe",
		"-vcodec",
		"ppm",
	}

	ffmpegControllerOpts := []ffmpeg.FFmpegControlerOpt{
		ffmpeg.WithFFmpegLogger(logger),
	}

	if config.Executable.Path != nil {
		ffmpegControllerOpts = append(ffmpegControllerOpts, ffmpeg.WithFFmpegExecutablePath(*config.Executable.Path))
	}

	ffmpegController, err := ffmpeg.NewFFmpegController(ffmpegInput, ffmpegOutput, ffmpegConfig, ffmpegControllerOpts...)
	if err != nil {
		return nil, fmt.Errorf("error creating ffmpeg controller: %w", err)
	}

	lifecycleCtx, lifecycleCancel := context.WithCancel(ctx)

	source := &DisplaySource{
		id:   id,
		name: name,

		ffmpegController: ffmpegController,

		lifecycleCancel: lifecycleCancel,
		lifecycleCtx:    lifecycleCtx,

		displayMode: displayMode,
		pixelFormat: pixelFormat,

		input: ffmpegInput,

		metrics:     &peripheralSDK.DisplaySourceMetrics{},
		metricsLock: sync.RWMutex{},

		logger: logger,
	}

	err = ffmpegController.Start(ctx)
	if err != nil {
		lifecycleCancel()
		return nil, fmt.Errorf("start ffmpeg controller: %w", err)
	}

	go source.ffmpegDataReadLoop(lifecycleCtx)

	source.logger.Debug("FFmpeg display source created.")

	return source, nil
}

// Capabilities returns the list of peripheral capabilities supported by this display source.
func (source *DisplaySource) GetCapabilities() []peripheralSDK.PeripheralCapability {
	return []peripheralSDK.PeripheralCapability{
		peripheralSDK.DisplaySourceCapability,
	}
}

// Id returns the unique identifier of this peripheral.
func (source *DisplaySource) GetId() peripheralSDK.PeripheralId {
	return source.id
}

func (source *DisplaySource) GetName() peripheralSDK.PeripheralName {
	return source.name
}

// Terminate shuts down the display source.
func (source *DisplaySource) Terminate(ctx context.Context) error {
	source.lifecycleCancel()

	err := source.ffmpegController.Stop(ctx)
	if err != nil {
		return fmt.Errorf("error stopping ffmpeg controller: %w", err)
	}

	return nil
}

// GetDisplayMode returns the current display mode configuration.
func (source *DisplaySource) GetDisplayMode(ctx context.Context) (*peripheralSDK.DisplayMode, error) {
	return &source.displayMode, nil
}

func (source *DisplaySource) GetDisplayPixelFormat(ctx context.Context) (*peripheralSDK.DisplayPixelFormat, error) {
	return &source.pixelFormat, nil
}

func (source *DisplaySource) GetDisplayFrameBuffer(ctx context.Context) (*peripheralSDK.DisplayFrameBuffer, error) {
	source.frameBufferLock.RLock()
	defer source.frameBufferLock.RUnlock()

	if source.frameBuffer == nil {
		return nil, peripheralSDK.ErrDisplayFrameBufferNotReady
	}

	err := source.frameBuffer.Retain()
	if err != nil {
		return nil, fmt.Errorf("retainin frame buffer: %w", err)
	}

	return source.frameBuffer, nil
}

func (source *DisplaySource) GetDisplaySourceMetrics() peripheralSDK.DisplaySourceMetrics {
	//TODO implement me
	panic("implement me")
}

func (source *DisplaySource) updateMetrics(updateFn func(metrics *peripheralSDK.DisplaySourceMetrics)) {
	source.metricsLock.Lock()
	defer source.metricsLock.Unlock()

	updateFn(source.metrics)
}

func (source *DisplaySource) ffmpegDataReadLoop(ctx context.Context) {
	err := ppm.ParseStream(ctx, source.ffmpegController.GetStdout(), source.frameBufferHandler)
	if err != nil {
		source.logger.Warn("Error parsing ffmpeg pipe data stream.")
	}
}

func (source *DisplaySource) frameBufferHandler(frameBuffer *peripheralSDK.DisplayFrameBuffer) error {
	source.frameBufferLock.Lock()
	defer source.frameBufferLock.Unlock()

	if source.frameBuffer != nil {
		err := source.frameBuffer.Release()
		if err != nil {
			return fmt.Errorf("error releasing old frame buffer to pool: %w", err)
		}
	}

	source.frameBuffer = frameBuffer

	return nil
}

var ErrDisplaySourceMissingInput = errors.New("display source missing input")

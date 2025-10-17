package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/ffmpeg"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

const DisplaySinkDriver = peripheralSDK.PeripheralDriver("ffmpeg/display-sink")

const defaultInputChannelBufferSize = 4

type DisplaySinkConfig struct {
	Title                 *string                       `json:"title"`
	SupportedDisplayModes peripheralSDK.DisplayModeList `json:"supportedDisplayModes"`
}

type DisplaySinkOptions struct {
	logger *slog.Logger
}

type DisplaySinkOpt func(*DisplaySinkOptions)

func defaultDisplaySinkOptions() DisplaySinkOptions {
	return DisplaySinkOptions{
		logger: slog.New(slog.DiscardHandler),
	}
}

func WithDisplaySinkLogger(logger *slog.Logger) DisplaySinkOpt {
	return func(options *DisplaySinkOptions) {
		options.logger = logger
	}
}

type DisplaySink struct {
	id    peripheralSDK.PeripheralId
	name  peripheralSDK.PeripheralName
	title string

	lifecycleCtx    context.Context
	lifecycleCancel context.CancelFunc

	framePumpTicker         *time.Ticker
	frameBufferProvider     peripheralSDK.DisplayFrameBufferProvider
	frameBufferProviderLock sync.RWMutex

	supportedDisplayModes  peripheralSDK.DisplayModeList
	currentDisplayMode     peripheralSDK.DisplayMode
	currentDisplayModeLock sync.RWMutex

	controller *ffmpeg.FFplayController

	logger *slog.Logger
}

var _ peripheralSDK.DisplaySink = (*DisplaySink)(nil)

func NewDisplaySink(ctx context.Context, config DisplaySinkConfig, name peripheralSDK.PeripheralName, opts ...DisplaySinkOpt) (*DisplaySink, error) {
	if len(config.SupportedDisplayModes) == 0 {
		return nil, ErrMissingSupportedDisplayMode
	}

	for _, displayMode := range config.SupportedDisplayModes {
		if err := displayMode.Valid(); err != nil {
			return nil, fmt.Errorf("invalid display mode: %w", err)
		}
	}

	options := defaultDisplaySinkOptions()
	for _, opt := range opts {
		opt(&options)
	}

	id := peripheralSDK.CreatePeripheralRandomId("ffmpeg-display-sink")
	title := utils.DefaultNil(config.Title, "ffplay-window")

	defaultDisplayMode := config.SupportedDisplayModes[0]

	logger := options.logger.With(slog.String("peripheralId", string(id)))

	controller, err := ffmpeg.NewFFplayController(ffmpeg.NewInputStdin(), ffmpeg.RawConfiguration{
		"-nodisp",
	},
		ffmpeg.WithFFplayLogger(logger),
	)
	if err != nil {
		return nil, fmt.Errorf("create ffplay controller: %w", err)
	}

	if err := controller.Start(ctx); err != nil {
		return nil, fmt.Errorf("start ffplay controller: %w", err)
	}

	lifecycleCtx, lifecycleCancel := context.WithCancel(ctx)

	displaySink := &DisplaySink{
		id:    id,
		name:  name,
		title: title,

		lifecycleCtx:    lifecycleCtx,
		lifecycleCancel: lifecycleCancel,

		framePumpTicker:         time.NewTicker(time.Second),
		frameBufferProviderLock: sync.RWMutex{},

		supportedDisplayModes:  config.SupportedDisplayModes,
		currentDisplayMode:     defaultDisplayMode,
		currentDisplayModeLock: sync.RWMutex{},

		controller: controller,

		logger: logger,
	}

	go displaySink.framePump(lifecycleCtx)

	err = displaySink.setControllerMissingInput(ctx)
	if err != nil {
		displaySink.lifecycleCancel()
		displaySink.framePumpTicker.Stop()
		displaySink.controller.Stop(ctx)
		return nil, fmt.Errorf("set controller missing input: %w", err)
	}

	displaySink.logger.Debug("The ffplay display sink created.")

	return displaySink, nil
}

func (sink *DisplaySink) GetCapabilities() []peripheralSDK.PeripheralCapability {
	return []peripheralSDK.PeripheralCapability{
		peripheralSDK.DisplaySinkCapability,
	}
}

func (sink *DisplaySink) GetName() peripheralSDK.PeripheralName {
	return sink.name
}

func (sink *DisplaySink) GetId() peripheralSDK.PeripheralId {
	return sink.id
}

func (sink *DisplaySink) GetDisplayInfo() (peripheralSDK.DisplayInfo, error) {
	sink.currentDisplayModeLock.RLock()
	defer sink.currentDisplayModeLock.RUnlock()

	return peripheralSDK.DisplayInfo{
		Manufacturer:   "FFmpeg",
		Model:          "FFplay Window",
		SerialNumber:   sink.id.String(),
		SupportedModes: sink.supportedDisplayModes,
		CurrentMode:    sink.currentDisplayMode,
	}, nil
}

func (sink *DisplaySink) SetDisplayFrameBufferProvider(provider peripheralSDK.DisplayFrameBufferProvider) error {
	providerDisplayMode, err := provider.GetDisplayMode(sink.lifecycleCtx)
	if err != nil {
		return fmt.Errorf("get display mode from provider: %w", err)
	}

	err = providerDisplayMode.Valid()
	if err != nil {
		return fmt.Errorf("invalid display mode: %w", err)
	}

	if !sink.supportedDisplayModes.Supports(*providerDisplayMode) {
		return ErrDisplayUnsupportedDisplayMode
	}

	if provider.GetDisplayPixelFormat(sink.lifecycleCtx) != peripheralSDK.DisplayPixelFormatRGB24 {
		return ErrDisplayPixelFormatUnsupported
	}

	sink.frameBufferProviderLock.Lock()
	sink.frameBufferProvider = provider
	sink.frameBufferProviderLock.Unlock()

	sink.currentDisplayModeLock.Lock()
	sink.currentDisplayMode = *providerDisplayMode
	sink.currentDisplayModeLock.Unlock()

	err = sink.setControllerValidInput(sink.lifecycleCtx)
	if err != nil {
		sink.frameBufferProviderLock.Lock()
		sink.frameBufferProvider = nil
		sink.frameBufferProviderLock.Unlock()
	}

	return err
}

func (sink *DisplaySink) ClearDisplayFrameBufferProvider() error {
	sink.frameBufferProviderLock.Lock()
	sink.frameBufferProvider = nil
	sink.frameBufferProviderLock.Unlock()

	return sink.setControllerMissingInput(sink.lifecycleCtx)
}

func (sink *DisplaySink) Terminate(ctx context.Context) error {
	sink.lifecycleCancel()

	sink.framePumpTicker.Stop()

	err := sink.controller.Stop(ctx)
	if err != nil {
		return fmt.Errorf("stop ffplay controller: %w", err)
	}

	return nil
}

func (sink *DisplaySink) framePump(ctx context.Context) {
	ticker := sink.framePumpTicker.C
	done := ctx.Done()

	for {
		select {
		case <-done:
			return
		case <-ticker:
			err := sink.writeFrameFromProvider()
			if err != nil {
				sink.logger.Warn("Failed to write frame from provider.", slog.String("error", err.Error()))
			}
		}
	}
}

func (sink *DisplaySink) writeFrameFromProvider() error {
	sink.frameBufferProviderLock.RLock()

	if sink.frameBufferProvider == nil {
		sink.frameBufferProviderLock.RUnlock()
		return nil
	}

	frameBuffer, err := sink.frameBufferProvider.GetDisplayFrameBuffer(sink.lifecycleCtx)
	sink.frameBufferProviderLock.RUnlock()

	if err != nil {
		return fmt.Errorf("get frame buffer from provider: %w", err)
	}

	defer func() {
		err = frameBuffer.Release()
		if err != nil {
			sink.logger.Warn("Failed to release frame buffer.", slog.String("error", err.Error()))
		}
	}()

	_, err = frameBuffer.WriteTo(sink.controller.GetStdin())
	if err != nil {
		return fmt.Errorf("write frame to stdin: %w", err)
	}

	return nil
}

func (sink *DisplaySink) setControllerMissingInput(ctx context.Context) error {
	sink.currentDisplayModeLock.Lock()
	displayMode := sink.currentDisplayMode
	sink.currentDisplayModeLock.Unlock()

	windowTitle := sink.getWindowTitle(displayMode, true)

	return sink.controller.SetInputWithConfiguration(ctx,
		ffmpeg.NewInputMessageBoard("[NO INPUT]", "ffmpeg-display-sink", displayMode),
		ffmpeg.RawConfiguration{
			"-window_title",
			windowTitle,
		},
	)
}

func (sink *DisplaySink) setControllerValidInput(ctx context.Context) error {
	sink.currentDisplayModeLock.Lock()
	displayMode := sink.currentDisplayMode
	sink.currentDisplayModeLock.Unlock()

	windowTitle := sink.getWindowTitle(displayMode, false)

	sink.framePumpTicker.Reset(time.Second / time.Duration(displayMode.RefreshRate))

	return sink.controller.SetInputWithConfiguration(ctx,
		ffmpeg.NewInputStdin(),
		ffmpeg.RawConfiguration{
			"-f",
			"rawvideo",
			"-pixel_format",
			"rgb24",
			"-video_size",
			fmt.Sprintf("%dx%d", displayMode.Width, displayMode.Height),
			"-framerate",
			fmt.Sprintf("%d", displayMode.RefreshRate),
			"-window_title",
			windowTitle,
		})
}

func (sink *DisplaySink) getWindowTitle(displayMode peripheralSDK.DisplayMode, noInput bool) string {
	if noInput {
		return fmt.Sprintf("%s [NO INPUT] [%s]", sink.title, displayMode.String())
	} else {
		return fmt.Sprintf("%s [%s]", sink.title, displayMode.String())
	}
}

var (
	ErrMissingSupportedDisplayMode   = errors.New("missing supported display modes")
	ErrDisplayUnsupportedDisplayMode = errors.New("display mode is not supported")
	ErrDisplayPixelFormatUnsupported = errors.New("display pixel format unsupported")
)

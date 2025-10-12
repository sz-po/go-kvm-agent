package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/ffmpeg"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/stream"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

// TODO: refine coments in this file

// DisplaySourceDriver is the peripheral driver identifier for FFMPEG-based display sources.
const DisplaySourceDriver = peripheralSDK.PeripheralDriver("ffmpeg/display-source")

type DisplaySourceInputConfig struct {
	TestPattern *DisplaySourceTestPatternInputConfig `json:"testPattern"`
}

type DisplaySourceInput interface {
	ffmpeg.Input
	GetCurrentPixelFormat() peripheralSDK.DisplayPixelFormat
	GetCurrentDisplayMode() peripheralSDK.DisplayMode
}

// DisplaySourceConfig holds configuration for creating an FFMPEG display source.
type DisplaySourceConfig struct {
	Input DisplaySourceInputConfig `json:"input"`
}

type DisplaySourceOptions struct {
	chunkSize int
	logger    *slog.Logger
}

func defaultDisplaySourceOptions(chunkSize int) *DisplaySourceOptions {
	return &DisplaySourceOptions{
		chunkSize: chunkSize,
		logger:    slog.New(slog.DiscardHandler),
	}
}

type DisplaySourceOpt func(options *DisplaySourceOptions)

// DisplaySource is a mpv implementation of a display source using FFMPEG.
type DisplaySource struct {
	id               peripheralSDK.PeripheralId
	ffmpegController *ffmpeg.Controller
	frameParser      *stream.PPMFrameParser

	pipeStop context.CancelFunc
	pipeCtx  context.Context

	input DisplaySourceInput

	dataEventQueue    *utils.EventEmitter[peripheralSDK.DisplayDataEvent]
	controlEventQueue *utils.EventEmitter[peripheralSDK.DisplayControlEvent]

	chunkSize int

	metrics     *peripheralSDK.DisplaySourceMetrics
	metricsLock sync.RWMutex

	logger *slog.Logger
}

var _ peripheralSDK.DisplaySource = (*DisplaySource)(nil)

func WithDisplaySourceChunkSize(chunkSize int) DisplaySourceOpt {
	return func(options *DisplaySourceOptions) {
		options.chunkSize = chunkSize
	}
}

func WithDisplaySourceLogger(logger *slog.Logger) DisplaySourceOpt {
	return func(options *DisplaySourceOptions) {
		options.logger = logger
	}
}

// NewDisplaySource creates a new FFMPEG display source from the provided configuration.
func NewDisplaySource(ctx context.Context, config DisplaySourceConfig, opts ...DisplaySourceOpt) (*DisplaySource, error) {
	id := peripheralSDK.CreatePeripheralRandomId("ffmpeg-display-source")

	var input DisplaySourceInput

	if config.Input.TestPattern != nil {
		input = NewDisplaySourceTestPatternInput(*config.Input.TestPattern)
	} else {
		return nil, ErrDisplaySourceMissingInput
	}

	lineSize := input.GetCurrentPixelFormat().BytesPerPixel() * int(input.GetCurrentDisplayMode().Width)

	options := defaultDisplaySourceOptions(lineSize * 16)
	for _, opt := range opts {
		opt(options)
	}

	if options.chunkSize <= 0 {
		return nil, fmt.Errorf("%w: chunk size must be greater than zero", ErrDisplaySourceInvalidChunkSize)
	}

	if options.chunkSize%lineSize != 0 {
		return nil, fmt.Errorf("%w: chunk size must be a multiple of line size", ErrDisplaySourceInvalidChunkSize)
	}

	frameParser, err := stream.NewPPMFrameParser(options.chunkSize)
	if err != nil {
		return nil, fmt.Errorf("error creating PPM frame parser: %w", err)
	}

	logger := options.logger.With(slog.String("peripheralId", id.String()))

	pipeCtx, pipeStop := context.WithCancel(context.Background())

	ffmpegOutput, err := ffmpeg.NewGolangChannelOutput(pipeCtx, ffmpeg.GolangChannelOutputModeUnixSocket)
	if err != nil {
		pipeStop()
		return nil, fmt.Errorf("create output pipe: %w", err)
	}

	ffmpegConfig := ffmpeg.RawConfiguration{
		"-f",
		"image2pipe",
		"-vcodec",
		"ppm",
	}

	ffmpegController, err := ffmpeg.NewController(input, ffmpegOutput, ffmpegConfig, ffmpeg.WithLogger(logger))
	if err != nil {
		pipeStop()
		return nil, fmt.Errorf("error creating ffmpeg controller: %w", err)
	}

	source := &DisplaySource{
		id: id,

		ffmpegController: ffmpegController,
		frameParser:      frameParser,

		pipeStop: pipeStop,
		pipeCtx:  pipeCtx,

		input: input,

		dataEventQueue:    utils.NewEventEmitter[peripheralSDK.DisplayDataEvent](),
		controlEventQueue: utils.NewEventEmitter[peripheralSDK.DisplayControlEvent](),

		chunkSize: options.chunkSize,

		metrics:     &peripheralSDK.DisplaySourceMetrics{},
		metricsLock: sync.RWMutex{},

		logger: logger,
	}

	err = ffmpegController.Start(ctx)
	if err != nil {
		pipeStop()
		return nil, fmt.Errorf("start ffmpeg controller: %w", err)
	}

	go source.ffmpegDataReadLoop(ffmpegOutput.Channel(ctx))

	source.logger.Debug("FFmpeg display source created.",
		slog.Int("chunkSize", options.chunkSize),
	)

	return source, nil
}

// Capabilities returns the list of peripheral capabilities supported by this display source.
func (source *DisplaySource) Capabilities() []peripheralSDK.PeripheralCapability {
	return []peripheralSDK.PeripheralCapability{
		peripheralSDK.DisplaySourceCapability,
	}
}

// Id returns the unique identifier of this peripheral.
func (source *DisplaySource) Id() peripheralSDK.PeripheralId {
	return source.id
}

// Terminate shuts down the display source.
func (source *DisplaySource) Terminate(ctx context.Context) error {
	err := source.ffmpegController.Stop(ctx)
	if err != nil {
		return fmt.Errorf("error stopping ffmpeg controller: %w", err)
	}

	source.pipeStop()

	return nil
}

// DisplayDataChannel returns a channel that emits display events containing frame data.
func (source *DisplaySource) DisplayDataChannel(ctx context.Context) <-chan peripheralSDK.DisplayDataEvent {
	return source.dataEventQueue.Listen(ctx)
}

// DisplayControlChannel returns a channel that emits display control events.
func (source *DisplaySource) DisplayControlChannel(ctx context.Context) <-chan peripheralSDK.DisplayControlEvent {
	return source.controlEventQueue.Listen(ctx)
}

// GetCurrentDisplayMode returns the current display mode configuration.
func (source *DisplaySource) GetCurrentDisplayMode() (*peripheralSDK.DisplayMode, error) {
	displayMode := source.input.GetCurrentDisplayMode()

	return &displayMode, nil
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

func (source *DisplaySource) emitDisplayDataEvent(event peripheralSDK.DisplayDataEvent) {
	source.dataEventQueue.Emit(event)
}

func (source *DisplaySource) ffmpegDataReadLoop(dataChannel <-chan []byte) {
	for dataChunk := range dataChannel {
		source.updateMetrics(func(metrics *peripheralSDK.DisplaySourceMetrics) {
			metrics.InputProcessedBytes += uint64(len(dataChunk))
			metrics.InputProcessedReadCalls += 1
		})

		var emittedDisplayFrameEndEventCount uint64
		var emittedDisplayFrameChunkEventCount uint64
		var emittedDisplayFrameStartEventCount uint64

		events, err := source.frameParser.Ingest(dataChunk)
		if err != nil {
			source.logger.Warn("Error while parsing PPM frame data.", slog.String("error", err.Error()))
		}

		for _, event := range events {
			switch event.(type) {
			case peripheralSDK.DisplayFrameStartEvent:
				emittedDisplayFrameStartEventCount += 1
			case peripheralSDK.DisplayFrameEndEvent:
				emittedDisplayFrameEndEventCount += 1
			case peripheralSDK.DisplayFrameChunkEvent:
				emittedDisplayFrameChunkEventCount += 1
			}
			source.emitDisplayDataEvent(event)
		}

		source.updateMetrics(func(metrics *peripheralSDK.DisplaySourceMetrics) {
			metrics.EmittedDisplayFrameStartEventCount += emittedDisplayFrameStartEventCount
			metrics.EmittedDisplayFrameEndEventCount += emittedDisplayFrameEndEventCount
			metrics.EmittedDisplayFrameChunkEventCount += emittedDisplayFrameChunkEventCount
		})
	}
}

var ErrDisplaySourceMissingInput = errors.New("display source missing input")
var ErrDisplaySourceInvalidChunkSize = errors.New("display source invalid chunk size")

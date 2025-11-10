//go:build linux

package v4l2

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"
	"github.com/mitchellh/mapstructure"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/driver"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/memory"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/linux/v4l2/tc358743"
	driverSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/driver"
	memorySDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/memory"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

const DisplaySourceDriverKind = driverSDK.Kind("v4l2-display-source")

var DisplaySourceDriver = driver.NewLocalDriver(DisplaySourceDriverKind, func(ctx context.Context, config any, name peripheralSDK.Name) (peripheralSDK.Peripheral, error) {
	driverConfig := DisplaySourceConfig{}

	err := mapstructure.Decode(config, &driverConfig)
	if err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	logger := slog.Default().With(slog.String("driverKind", DisplaySourceDriverKind.String()))

	displaySource, err := NewDisplaySource(ctx, driverConfig, name, WithDisplaySourceLogger(logger))
	if err != nil {
		return nil, fmt.Errorf("create peripheral: %w", err)
	}

	return displaySource, nil
})

func createDisplaySourceId(devicePath string, sourceType string) (peripheralSDK.Id, error) {
	devicePath = strings.TrimLeft(devicePath, "/")
	devicePath = strings.Replace(devicePath, "/", "-", -1)
	devicePath = strcase.ToKebab(devicePath)

	id := fmt.Sprintf("v4l2-%s-display-source-%s", sourceType, devicePath)

	return peripheralSDK.NewPeripheralId(id)
}

type DisplaySourceOptions struct {
	logger             *slog.Logger
	memoryPoolProvider memorySDK.PoolProvider
}

type DisplaySourceOpt func(*DisplaySourceOptions)

func defaultDisplaySourceOptions() *DisplaySourceOptions {
	return &DisplaySourceOptions{
		logger:             slog.New(slog.DiscardHandler),
		memoryPoolProvider: memory.DefaultMemoryPoolProvider,
	}
}

type DisplaySourceConfig struct {
	DevicePath string `json:"devicePath" validate:"required"`
}

type DisplaySource struct {
	id   peripheralSDK.Id
	name peripheralSDK.Name

	lifecycleCtx    context.Context
	lifecycleCancel context.CancelFunc

	frameBuffer     *peripheralSDK.DisplayFrameBuffer
	frameBufferLock *sync.RWMutex

	videoDevice *tc358743.Device

	memoryPool memorySDK.Pool
	logger     *slog.Logger
}

func WithDisplaySourceLogger(logger *slog.Logger) DisplaySourceOpt {
	return func(options *DisplaySourceOptions) {
		options.logger = logger
	}
}

func NewDisplaySource(ctx context.Context, config DisplaySourceConfig, name peripheralSDK.Name, opts ...DisplaySourceOpt) (*DisplaySource, error) {
	err := validator.New(validator.WithRequiredStructEnabled()).Struct(&config)
	if err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	id, err := createDisplaySourceId(config.DevicePath, "tc358743")
	if err != nil {
		return nil, fmt.Errorf("create display source id: %w", err)
	}

	options := defaultDisplaySourceOptions()

	for _, opt := range opts {
		opt(options)
	}

	memoryPool, err := options.memoryPoolProvider()
	if err != nil {
		return nil, fmt.Errorf("memory pool: %w", err)
	}

	lifecycleCtx, lifecycleCancel := context.WithCancel(ctx)

	logger := options.logger.With(slog.String("peripheralId", id.String()))

	source := &DisplaySource{
		id:   id,
		name: name,

		lifecycleCtx:    lifecycleCtx,
		lifecycleCancel: lifecycleCancel,

		frameBuffer:     nil,
		frameBufferLock: &sync.RWMutex{},

		memoryPool: memoryPool,
		logger:     logger,
	}

	videoDevice, err := tc358743.Open(lifecycleCtx, config.DevicePath,
		tc358743.WithLogger(logger),
		tc358743.WithFrameHandler(source.frameHandler),
	)
	if err != nil {
		lifecycleCancel()
		return nil, fmt.Errorf("open device: %w", err)
	}
	source.videoDevice = videoDevice

	return source, nil
}

func (source *DisplaySource) GetCapabilities() []peripheralSDK.PeripheralCapability {
	return []peripheralSDK.PeripheralCapability{
		peripheralSDK.DisplaySourceCapability,
	}
}

func (source *DisplaySource) GetId() peripheralSDK.Id {
	return source.id
}

func (source *DisplaySource) GetName() peripheralSDK.Name {
	return source.name
}

func (source *DisplaySource) Terminate(ctx context.Context) error {
	source.lifecycleCancel()
	return nil
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

func (source *DisplaySource) GetDisplayMode(ctx context.Context) (*peripheralSDK.DisplayMode, error) {
	return source.videoDevice.GetDisplayMode()
}

func (source *DisplaySource) GetDisplayPixelFormat(ctx context.Context) (*peripheralSDK.DisplayPixelFormat, error) {
	pixelFormat := peripheralSDK.DisplayPixelFormatRGB24
	return &pixelFormat, nil
}

func (source *DisplaySource) GetDisplaySourceMetrics() peripheralSDK.DisplaySourceMetrics {
	//TODO implement me
	panic("implement me")
}

func (source *DisplaySource) frameHandler(memoryBuffer memorySDK.Buffer) error {
	source.frameBufferLock.Lock()
	defer source.frameBufferLock.Unlock()

	if source.frameBuffer != nil {
		err := source.frameBuffer.Release()
		if err != nil {
			return fmt.Errorf("releasing previous frame buffer: %w", err)
		}
	}

	source.frameBuffer = peripheralSDK.NewDisplayFrameBuffer(memoryBuffer)

	return nil
}

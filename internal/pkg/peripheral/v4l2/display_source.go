package v4l2

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/iancoleman/strcase"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

const DisplaySourceDriver = peripheralSDK.PeripheralDriver("v4l2/display-source")

type DisplaySourceOptions struct {
	logger *slog.Logger
}

type DisplaySourceOpt func(*DisplaySourceOptions)

func defaultDisplaySourceOptions() *DisplaySourceOptions {
	return &DisplaySourceOptions{
		logger: slog.New(slog.DiscardHandler),
	}
}

type DisplaySourceConfig struct {
	DevicePath            string                        `json:"devicePath" validate:"required"`
	SupportedDisplayModes peripheralSDK.DisplayModeList `json:"supportedDisplayModes" validate:"required,dive"`
}

type DisplaySource struct {
	id   peripheralSDK.PeripheralId
	name peripheralSDK.PeripheralName

	devicePath string

	logger *slog.Logger
}

func WithDisplaySourceLogger(logger *slog.Logger) DisplaySourceOpt {
	return func(options *DisplaySourceOptions) {
		options.logger = logger
	}
}

func NewDisplaySource(ctx context.Context, config DisplaySourceConfig, name peripheralSDK.PeripheralName, opts ...DisplaySourceOpt) (*DisplaySource, error) {
	id, err := peripheralSDK.NewPeripheralId(fmt.Sprintf("v4l2-%f", strcase.ToKebab(config.DevicePath)))
	if err != nil {
		return nil, fmt.Errorf("new peripheral id: %w", err)
	}

	options := defaultDisplaySourceOptions()

	for _, opt := range opts {
		opt(options)
	}

	source := &DisplaySource{
		id:     id,
		name:   name,
		logger: options.logger,
	}

	return source, nil
}

func (source *DisplaySource) GetCapabilities() []peripheralSDK.PeripheralCapability {
	return []peripheralSDK.PeripheralCapability{
		peripheralSDK.DisplaySourceCapability,
	}
}

func (source *DisplaySource) GetId() peripheralSDK.PeripheralId {
	return source.id
}

func (source *DisplaySource) GetName() peripheralSDK.PeripheralName {
	return source.name
}

func (source *DisplaySource) Terminate(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (source *DisplaySource) GetDisplayFrameBuffer(ctx context.Context) (*peripheralSDK.DisplayFrameBuffer, error) {
	//TODO implement me
	panic("implement me")
}

func (source *DisplaySource) GetDisplayMode(ctx context.Context) (*peripheralSDK.DisplayMode, error) {
	//TODO implement me
	panic("implement me")
}

func (source *DisplaySource) GetDisplayPixelFormat(ctx context.Context) (*peripheralSDK.DisplayPixelFormat, error) {
	//TODO implement me
	panic("implement me")
}

func (source *DisplaySource) GetDisplaySourceMetrics() peripheralSDK.DisplaySourceMetrics {
	//TODO implement me
	panic("implement me")
}

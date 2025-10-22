package peripheral

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/mitchellh/mapstructure"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/peripheral/ffmpeg"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/peripheral/v4l2"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

// CreatePeripheralFromConfig creates a peripheral instance from the provided configuration.
// It delegates to driver-specific factory functions based on the peripheral driver.
func CreatePeripheralFromConfig(ctx context.Context, config PeripheralConfig) (peripheralSDK.Peripheral, error) {
	logger := slog.With(
		slog.String("peripheralDriver", string(config.Driver)),
	)

	switch config.Driver {
	case ffmpeg.DisplaySourceDriver:
		driverConfig := ffmpeg.DisplaySourceConfig{}

		err := mapstructure.Decode(config.Config, &driverConfig)
		if err != nil {
			return nil, fmt.Errorf("decode ffmpeg-display-source config: %w", err)
		}

		ffmpegDisplaySource, err := ffmpeg.NewDisplaySource(ctx, driverConfig, config.Name, ffmpeg.WithDisplaySourceLogger(logger))
		if err != nil {
			return nil, fmt.Errorf("create ffmpeg-display-source peripheral: %w", err)
		}

		return ffmpegDisplaySource, nil
	case ffmpeg.DisplaySinkDriver:
		driverConfig := ffmpeg.DisplaySinkConfig{}

		err := mapstructure.Decode(config.Config, &driverConfig)
		if err != nil {
			return nil, fmt.Errorf("decode ffmpeg-display-sink config: %w", err)
		}

		ffmpegDisplaySink, err := ffmpeg.NewDisplaySink(ctx, driverConfig, config.Name, ffmpeg.WithDisplaySinkLogger(logger))
		if err != nil {
			return nil, fmt.Errorf("create ffmpeg-display-sink peripheral: %w", err)
		}

		return ffmpegDisplaySink, nil
	case v4l2.DisplaySourceDriver:
		driverConfig := v4l2.DisplaySourceConfig{}

		err := mapstructure.Decode(config.Config, &driverConfig)
		if err != nil {
			return nil, fmt.Errorf("decode v4l2-display-source config: %w", err)
		}

		v4l2DisplaySource, err := v4l2.NewDisplaySource(ctx, driverConfig, config.Name)
		if err != nil {
			return nil, fmt.Errorf("create v4l2-display-source peripheral: %w", err)
		}

		return v4l2DisplaySource, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPeripheralDriver, config.Driver)
	}
}

var (
	// ErrUnsupportedPeripheralDriver indicates that the peripheral driver is not supported by the factory.
	ErrUnsupportedPeripheralDriver = errors.New("unsupported peripheral driver")
)

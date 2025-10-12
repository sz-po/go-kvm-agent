package peripheral

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/mitchellh/mapstructure"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/peripheral/ffmpeg"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/peripheral/mpv"
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
		ffmpegDisplaySourceConfig := ffmpeg.DisplaySourceConfig{}

		err := mapstructure.Decode(config.Config, &ffmpegDisplaySourceConfig)
		if err != nil {
			return nil, fmt.Errorf("decode ffmpeg-display-source config: %w", err)
		}

		ffmpegDisplaySource, err := ffmpeg.NewDisplaySource(ctx, ffmpegDisplaySourceConfig, ffmpeg.WithDisplaySourceLogger(logger))
		if err != nil {
			return nil, fmt.Errorf("create ffmpeg-display-source peripheral: %w", err)
		}

		return ffmpegDisplaySource, nil
	case mpv.WindowDriver:
		gioWindowConfig := mpv.WindowConfig{}

		err := mapstructure.Decode(config.Config, &gioWindowConfig)
		if err != nil {
			return nil, fmt.Errorf("decode gio-window config: %w", err)
		}

		gioWindow, err := mpv.NewMPVWindow(ctx, gioWindowConfig)
		if err != nil {
			return nil, fmt.Errorf("create gio-window peripheral: %w", err)
		}

		return gioWindow, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPeripheralDriver, config.Driver)
	}
}

var (
	// ErrUnsupportedPeripheralDriver indicates that the peripheral driver is not supported by the factory.
	ErrUnsupportedPeripheralDriver = errors.New("unsupported peripheral driver")
)

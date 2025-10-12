package mpv

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/mpv"

	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

// WindowDriver is the peripheral driver identifier for MPV window-based display sinks.
const WindowDriver = peripheralSDK.PeripheralDriver("mpv/window")

// WindowConfig holds configuration for creating an MPV window display sink.
type WindowConfig struct {
	Title                 *string                       `json:"title"`
	SupportedDisplayModes peripheralSDK.DisplayModeList `json:"supportedDisplayModes"`
}

// Window is a mpv implementation of a display sink that renders video in an MPV window.
type Window struct {
	mpvController         *mpv.Controller
	id                    peripheralSDK.PeripheralId
	supportedDisplayModes peripheralSDK.DisplayModeList
	currentDisplayMode    *peripheralSDK.DisplayMode
	title                 string
	logger                *slog.Logger
}

var _ peripheralSDK.DisplaySink = (*Window)(nil)

// NewMPVWindow creates a new MPV window display sink from the provided configuration.
func NewMPVWindow(ctx context.Context, config WindowConfig) (*Window, error) {
	id := peripheralSDK.CreatePeripheralRandomId("mpv-window")

	if len(config.SupportedDisplayModes) == 0 {
		return nil, ErrMissingSupportedDisplayMode
	}

	mpvController, err := mpv.NewController(
		mpv.WithStaticParameters(
			mpv.NoConfig.Render(true),
			mpv.ForceWindow.Render(mpv.ForceWindowImmediate),
			mpv.Untimed.Render(true),
			mpv.Cache.Render(false),
			mpv.Osc.Render(false),
			mpv.OsdLevel.Render(0),
			mpv.Demuxer.Render(mpv.DemuxerRawVideo),
			mpv.NoTerminal.Render(true),
			mpv.MsgLevel.Render(mpv.MsgLevelAllFatal),
			mpv.DemuxerRawVideoMpFormat.Render(mpv.RawVideoFormatRGB24),
		),
		mpv.WithRequiredParameters(
			mpv.DemuxerRawVideoWidth,
			mpv.DemuxerRawVideoHeight,
			mpv.DemuxerRawVideoFps,
			mpv.DemuxerRawVideoMpFormat,
			mpv.Title,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create mpv controller: %w", err)
	}

	err = mpvController.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("start mpv controller: %w", err)
	}

	window := &Window{
		id:                    id,
		mpvController:         mpvController,
		supportedDisplayModes: config.SupportedDisplayModes,
		title:                 utils.DefaultNil(config.Title, "mpv-window"),
		logger:                slog.Default(),
	}

	return window, nil
}

// Capabilities returns the list of peripheral capabilities supported by this display sink.
func (window *Window) Capabilities() []peripheralSDK.PeripheralCapability {
	return []peripheralSDK.PeripheralCapability{
		peripheralSDK.DisplaySinkCapability,
	}
}

// Id returns the unique identifier of this peripheral.
func (window *Window) Id() peripheralSDK.PeripheralId {
	return window.id
}

// HandleDisplayDataEvent processes incoming display data events and renders frames.
func (window *Window) HandleDisplayDataEvent(event peripheralSDK.DisplayDataEvent) error {
	panic("implement me")
}

func (window *Window) HandleDisplayControlEvent(event peripheralSDK.DisplayControlEvent) error {
	//TODO implement me
	panic("implement me")
}

// GetDisplayInfo returns information about the display including supported and current modes.
func (window *Window) GetDisplayInfo() (peripheralSDK.DisplayInfo, error) {
	return peripheralSDK.DisplayInfo{
		Manufacturer:   "MPV",
		Model:          "Test Window",
		SerialNumber:   window.id.String(),
		SupportedModes: window.supportedDisplayModes,
		CurrentMode:    window.currentDisplayMode,
	}, nil
}

// SetDisplayMode configures the MPV window to use the specified display mode.
func (window *Window) SetDisplayMode(requestedMode peripheralSDK.DisplayMode) error {
	window.logger.Info("Setting display requestedMode.",
		slog.Uint64("width", uint64(requestedMode.Width)),
		slog.Uint64("height", uint64(requestedMode.Height)),
		slog.Uint64("refreshRate", uint64(requestedMode.RefreshRate)),
	)

	if !window.supportedDisplayModes.Supports(requestedMode) {
		return fmt.Errorf("%w: %s", peripheralSDK.ErrUnsupportedDisplayMode, requestedMode)
	}

	window.currentDisplayMode = &requestedMode

	window.mpvController.SetParameters(
		mpv.DemuxerRawVideoHeight.Render(mpv.Integer(requestedMode.Height)),
		mpv.DemuxerRawVideoWidth.Render(mpv.Integer(requestedMode.Width)),
		mpv.DemuxerRawVideoFps.Render(mpv.Integer(requestedMode.RefreshRate)),
		mpv.Title.Render(mpv.String(fmt.Sprintf("%s [%s]", window.title, requestedMode))),
	)

	window.mpvController.Reload()

	return nil
}

// Terminate shuts down the MPV window display sink.
func (window *Window) Terminate(ctx context.Context) error {
	err := window.mpvController.Stop()
	if err != nil {
		return fmt.Errorf("stop mpv controller: %w", err)
	}

	return nil
}

var ErrMissingSupportedDisplayMode = errors.New("missing supported display modes")

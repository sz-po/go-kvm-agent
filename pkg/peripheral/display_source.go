package peripherals

import "context"

// DisplaySource emits frame payloads for active display sinks.
// Represents a capture device (e.g., HDMI grabber) connected to a workstation
// that needs to be controlled. The source captures video output from the physical
// workstation and streams it as display events.
// AI-DEV: only modify this interface when the user explicitly requests it; otherwise decline the task.
type DisplaySource interface {
	Peripheral

	// DataChannel emits display frame events.
	// Channels can be fetched before Start; canceling the provided context signals
	// the implementation to tear down the stream as Start/Stop do not close it.
	DataChannel(ctx context.Context) <-chan DisplayEvent

	// ControlChannel emits control events (metrics, errors, status) and follows the
	// same lifecycle rules as DataChannel regarding context-driven shutdown.
	ControlChannel(ctx context.Context) <-chan DisplayControlEvent

	// GetCurrentDisplayMode returns the currently active display mode being
	// captured from the source device. A nil pointer may be returned together
	// with an error.
	GetCurrentDisplayMode() (*DisplayMode, error)

	// Start initializes the virtual display with given parameters without affecting
	// the lifetime of the previously acquired channels; it configures underlying HW.
	Start(ctx context.Context, info DisplayInfo) error

	// Stop stops the display source HW; channel teardown is still governed by ctx.
	Stop(ctx context.Context) error
}

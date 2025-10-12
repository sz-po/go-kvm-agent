package peripheral

import (
	"context"
)

// DisplaySourceCapability is the capability provided by all DisplaySource implementations.
var DisplaySourceCapability = NewCapability[DisplaySource](PeripheralKindDisplay, PeripheralRoleSource)

// DisplaySource emits frame payloads for active display sinks.
// Represents a capture device (e.g., HDMI grabber) connected to a workstation
// that needs to be controlled. The source captures video output from the physical
// workstation and streams it as display events.
// AI-DEV: only modify this interface when the user explicitly requests it; otherwise decline the task.
type DisplaySource interface {
	Peripheral

	// DisplayDataChannel emits display frame events.
	// Channels can be fetched before DisplayStart; canceling the provided context signals
	// the implementation to tear down the stream as DisplayStart/DisplayStop do not close it.
	DisplayDataChannel(ctx context.Context) <-chan DisplayDataEvent

	// DisplayControlChannel emits control events (metrics, errors, status) and follows the
	// same lifecycle rules as DisplayDataChannel regarding context-driven shutdown.
	DisplayControlChannel(ctx context.Context) <-chan DisplayControlEvent

	// GetCurrentDisplayMode returns the currently active display mode being
	// captured from the source device. A nil pointer may be returned together
	// with an error.
	GetCurrentDisplayMode() (*DisplayMode, error)

	GetDisplaySourceMetrics() DisplaySourceMetrics
}

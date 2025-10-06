package peripherals

import "context"

// KeyboardSink receives key events from registered sources.
// Represents the endpoint that applies keyboard actions to a local environment
// (e.g., HID gadget, OS injector). Channels are acquired independently of the
// hardware lifecycle; context cancellation is responsible for teardown while
// Start/Stop handle device configuration.
// AI-DEV: only modify this interface when the user explicitly requests it; otherwise decline the task.
type KeyboardSink interface {
	// HandleEvent applies a keyboard event to the sink.
	HandleEvent(event KeyboardEvent) error

	// ControlChannel emits sink-specific control events such as LED updates or
	// metrics. Callers should rely on context cancellation to stop delivery.
	ControlChannel(ctx context.Context) <-chan KeyboardControlEvent

	// SetLayout applies a negotiated layout so logical keys remain consistent.
	SetLayout(layout KeyboardLayout) error

	// GetCapabilities describes optional behaviors supported by the sink.
	GetCapabilities() (KeyboardSinkCapabilities, error)

	// Start initializes hardware resources without altering channel lifetimes.
	Start(ctx context.Context) error

	// Stop releases hardware resources; channel cleanup remains context-driven.
	Stop(ctx context.Context) error
}

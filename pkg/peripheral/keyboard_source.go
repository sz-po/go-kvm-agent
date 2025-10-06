package peripherals

import "context"

// KeyboardSource emits keyboard events for downstream sinks.
// Represents a physical or virtual keyboard whose state needs to be captured
// and forwarded. Channels can be fetched before Start; canceling the supplied
// context is the canonical teardown signal because Start/Stop deal only with
// hardware initialization and release.
// AI-DEV: only modify this interface when the user explicitly requests it; otherwise decline the task.
type KeyboardSource interface {
	Peripheral

	// EventChannel emits keyboard data events.
	EventChannel(ctx context.Context) <-chan KeyboardEvent

	// ControlChannel emits control events (metrics, errors, layout changes). It
	// follows the same lifecycle rules as EventChannel and is closed via context.
	ControlChannel(ctx context.Context) <-chan KeyboardControlEvent

	// GetCurrentLayout returns the active layout negotiated with the device.
	GetCurrentLayout() (KeyboardLayout, error)

	// Start configures the source hardware without affecting channel lifetimes.
	Start(ctx context.Context, info KeyboardInfo) error

	// Stop releases the source hardware; channel cleanup remains context driven.
	Stop(ctx context.Context) error
}

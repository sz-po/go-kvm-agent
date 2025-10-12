package peripheral

import (
	"context"
)

// KeyboardSourceCapability is the capability provided by all KeyboardSource implementations.
var KeyboardSourceCapability = NewCapability[KeyboardSource](PeripheralKindKeyboard, PeripheralRoleSource)

// KeyboardSource emits keyboard events for downstream sinks.
// Represents a physical or virtual keyboard whose state needs to be captured
// and forwarded. Channels can be fetched before KeyboardStart; canceling the supplied
// context is the canonical teardown signal because KeyboardStart/KeyboardStop deal only with
// hardware initialization and release.
// AI-DEV: only modify this interface when the user explicitly requests it; otherwise decline the task.
type KeyboardSource interface {
	Peripheral

	// KeyboardDataChannel emits keyboard data events.
	KeyboardDataChannel(ctx context.Context) <-chan KeyboardEvent

	// KeyboardControlChannel emits control events (metrics, errors, layout changes). It
	// follows the same lifecycle rules as KeyboardDataChannel and is closed via context.
	KeyboardControlChannel(ctx context.Context) <-chan KeyboardControlEvent

	// GetCurrentLayout returns the active layout negotiated with the device.
	GetCurrentLayout() (KeyboardLayout, error)
}

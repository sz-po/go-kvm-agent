package peripheral

import "context"

// KeyboardSinkCapability is the capability provided by all KeyboardSink implementations.
var KeyboardSinkCapability = NewCapability[KeyboardSink](PeripheralKindKeyboard, PeripheralRoleSink)

// KeyboardSink receives key events from registered sources.
// Represents the endpoint that applies keyboard actions to a local environment
// (e.g., HID gadget, OS injector). Channels are acquired independently of the
// hardware lifecycle; context cancellation is responsible for teardown while
// KeyboardStart/KeyboardStop handle device configuration.
// AI-DEV: only modify this interface when the user explicitly requests it; otherwise decline the task.
type KeyboardSink interface {
	Peripheral

	// HandleKeyboardDataEvent applies a keyboard event to the sink.
	HandleKeyboardDataEvent(event KeyboardEvent) error

	// KeyboardControlChannel emits sink-specific control events such as LED updates or
	// metrics. Callers should rely on context cancellation to stop delivery.
	KeyboardControlChannel(ctx context.Context) <-chan KeyboardControlEvent
}

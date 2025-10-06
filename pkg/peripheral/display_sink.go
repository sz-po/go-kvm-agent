package peripheral

import "context"

// DisplaySink consumes rendered frames routed from display sources.
// Represents the actual physical display (monitor, screen) or virtual display
// where captured frames are rendered for viewing by the user.
// AI-DEV: only modify this interface when the user explicitly requests it; otherwise decline the task.
type DisplaySink interface {
	Peripheral

	// HandleDataEvent processes incoming frame data; callers should rely on context
	// cancellation to stop delivery, as Start/Stop leave channel lifetimes intact.
	HandleDataEvent(event DisplayEvent) error

	// ControlChannel emits sink-specific control events (metrics, errors) and can be
	// obtained before Start; canceling the context signals shutdown to the sink.
	ControlChannel(ctx context.Context) <-chan DisplayControlEvent

	// GetDisplayInfo returns information about the actual display
	// (the real monitor or rendering target).
	GetDisplayInfo() (DisplayInfo, error)

	// SetDisplayMode sets the desired display mode for rendering.
	SetDisplayMode(mode DisplayMode) error

	// Start initializes the display sink hardware, including pushing the negotiated
	// configuration (e.g., grabber display modes) without touching channel lifetime.
	Start(ctx context.Context) error

	// Stop stops the display sink hardware; channel cleanup remains context-driven.
	Stop(ctx context.Context) error
}

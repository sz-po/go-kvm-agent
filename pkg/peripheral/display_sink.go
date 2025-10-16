package peripheral

// DisplaySinkCapability is the capability provided by all DisplaySink implementations.
var DisplaySinkCapability = NewCapability[DisplaySink](PeripheralKindDisplay, PeripheralRoleSink)

// DisplaySink consumes rendered frames routed from display sources.
// Represents the actual physical display (monitor, screen) or virtual display
// where captured frames are rendered for viewing by the user.
// AI-DEV: only modify this interface when the user explicitly requests it; otherwise decline the task.
type DisplaySink interface {
	Peripheral

	SetDisplayFrameBufferProvider(provider DisplayFrameBufferProvider) error
	ClearDisplayFrameBufferProvider() error
}

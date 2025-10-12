package peripheral

// MouseSinkCapability is the capability provided by all MouseSink implementations.
var MouseSinkCapability = NewCapability[MouseSink](PeripheralKindMouse, PeripheralRoleSink)

// MouseSink applies pointer events received from mouse sources.
// AI-DEV: only modify this interface when the user explicitly requests it; otherwise decline the task.
type MouseSink interface {
	Peripheral
}

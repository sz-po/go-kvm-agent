package peripheral

// DisplaySourceCapability is the capability provided by all DisplaySource implementations.
var DisplaySourceCapability = NewCapability[DisplaySource](PeripheralKindDisplay, PeripheralRoleSource)

// DisplaySource emits frame payloads for active display sinks.
// Represents a capture device (e.g., HDMI grabber) connected to a workstation
// that needs to be controlled. The source captures video output from the physical
// workstation and streams it as display events.
// AI-DEV: only modify this interface when the user explicitly requests it; otherwise decline the task.
type DisplaySource interface {
	Peripheral
	DisplayFrameBufferProvider

	GetDisplaySourceMetrics() DisplaySourceMetrics
}

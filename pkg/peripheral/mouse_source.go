package peripheral

// MouseSourceCapability is the capability provided by all MouseSource implementations.
var MouseSourceCapability = NewCapability[MouseSource](PeripheralKindMouse, PeripheralRoleSource)

// MouseSource emits pointer events for routed sinks.
// AI-DEV: only modify this interface when the user explicitly requests it; otherwise decline the task.
type MouseSource interface {
	Peripheral
}

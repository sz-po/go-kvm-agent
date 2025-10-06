package peripherals

// MouseSource emits pointer events for routed sinks.
// AI-DEV: only modify this interface when the user explicitly requests it; otherwise decline the task.
type MouseSource interface {
	Peripheral
}

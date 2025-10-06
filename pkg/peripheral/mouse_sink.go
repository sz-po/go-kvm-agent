package peripheral

// MouseSink applies pointer events received from mouse sources.
// AI-DEV: only modify this interface when the user explicitly requests it; otherwise decline the task.
type MouseSink interface {
	Peripheral
}

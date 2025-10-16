package peripheral

// TODO: refine comments

// DisplaySourceMetrics captures throughput counters exposed by display sources.
type DisplaySourceMetrics struct {
	// InputProcessedBytes counts the bytes read from the raw input stream (for example HDMI or socket transport).
	InputProcessedBytes uint64 `json:"inputProcessedBytes"`

	// InputProcessedReadCalls counts how many read operations were executed against the input stream.
	InputProcessedReadCalls uint64 `json:"inputProcessedReadCalls"`

	// FrameBufferSwaps indicates how many times frame buffer was swapped.
	FrameBufferSwaps uint64 `json:"frameBufferSwaps"`

	// FrameBufferWrittenBytes indicates how many bytes was written to frame buffer.
	FrameBufferWrittenBytes uint64 `json:"frameBufferWrittenBytes"`

	// FramesPerSecond indicates how many frames per second source process.
	FramesPerSecond uint64 `json:"framesPerSecond"`

	// AdditionalMetrics may contain metrics related to underlying source.
	AdditionalMetrics map[string]interface{} `json:"additionalMetrics"`
}

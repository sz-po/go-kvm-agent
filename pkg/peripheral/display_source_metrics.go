package peripheral

// TODO: Fix and popraw komentarze

type DisplaySourceMetrics struct {
	// InputProcessedBytes indicates how many bytes are read from raw input (raw HDMI, socket, etc).
	InputProcessedBytes uint64 `json:"inputProcessedBytes"`

	// InputProcessedReadCalls indicates how many read calls was made to read data from input.
	InputProcessedReadCalls uint64 `json:"inputProcessedReadCalls"`

	EmittedDisplayFrameEndEventCount   uint64 `json:"emittedDisplayFrameEndEventCount"`
	EmittedDisplayFrameChunkEventCount uint64 `json:"emittedDisplayFrameChunkEventCount"`
	EmittedDisplayFrameStartEventCount uint64 `json:"emittedDisplayFrameStartEventCount"`
}

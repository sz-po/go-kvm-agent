package peripheral

import "time"

// DisplayEventType defines the type of display data event.
type DisplayEventType int

const (
	DisplayEventUnknown DisplayEventType = iota
	DisplayEventFrameStart
	DisplayEventFrameChunk
	DisplayEventFrameEnd
)

// DisplayEvent represents a display data event.
type DisplayEvent interface {
	Type() DisplayEventType
	Timestamp() time.Time
}

// DisplayFrameStartEvent signals the start of a new frame.
type DisplayFrameStartEvent struct {
	FrameID   uint64
	Width     uint32
	Height    uint32
	Format    DisplayPixelFormat
	timestamp time.Time
}

// NewDisplayFrameStartEvent constructs a DisplayFrameStartEvent with a preset timestamp.
func NewDisplayFrameStartEvent(frameID uint64, width, height uint32, format DisplayPixelFormat, timestamp time.Time) DisplayFrameStartEvent {
	return DisplayFrameStartEvent{
		FrameID:   frameID,
		Width:     width,
		Height:    height,
		Format:    format,
		timestamp: timestamp,
	}
}

// Type returns the event type.
func (e DisplayFrameStartEvent) Type() DisplayEventType {
	return DisplayEventFrameStart
}

// Timestamp returns the event timestamp.
func (e DisplayFrameStartEvent) Timestamp() time.Time {
	return e.timestamp
}

// DisplayFrameChunkEvent contains a chunk of frame data.
type DisplayFrameChunkEvent struct {
	FrameID    uint64
	ChunkIndex uint32
	Data       []byte
	timestamp  time.Time
}

// NewDisplayFrameChunkEvent constructs a DisplayFrameChunkEvent with a preset timestamp.
func NewDisplayFrameChunkEvent(frameID uint64, chunkIndex uint32, data []byte, timestamp time.Time) DisplayFrameChunkEvent {
	return DisplayFrameChunkEvent{
		FrameID:    frameID,
		ChunkIndex: chunkIndex,
		Data:       data,
		timestamp:  timestamp,
	}
}

// Type returns the event type.
func (e DisplayFrameChunkEvent) Type() DisplayEventType {
	return DisplayEventFrameChunk
}

// Timestamp returns the event timestamp.
func (e DisplayFrameChunkEvent) Timestamp() time.Time {
	return e.timestamp
}

// DisplayFrameEndEvent signals the end of a frame.
type DisplayFrameEndEvent struct {
	FrameID     uint64
	TotalChunks uint32
	timestamp   time.Time
}

// NewDisplayFrameEndEvent constructs a DisplayFrameEndEvent with a preset timestamp.
func NewDisplayFrameEndEvent(frameID uint64, totalChunks uint32, timestamp time.Time) DisplayFrameEndEvent {
	return DisplayFrameEndEvent{
		FrameID:     frameID,
		TotalChunks: totalChunks,
		timestamp:   timestamp,
	}
}

// Type returns the event type.
func (e DisplayFrameEndEvent) Type() DisplayEventType {
	return DisplayEventFrameEnd
}

// Timestamp returns the event timestamp.
func (e DisplayFrameEndEvent) Timestamp() time.Time {
	return e.timestamp
}

// DisplayControlEventType defines the type of display control event.
type DisplayControlEventType int

const (
	DisplayControlUnknown DisplayControlEventType = iota
	DisplayControlFrameDropped
	DisplayControlCaptureError
	DisplayControlMetrics
	DisplayControlModeChanged
	DisplayControlSourceStarted
	DisplayControlSourceStopped
	DisplayControlSinkStarted
	DisplayControlSinkStopped
)

// DisplayControlEvent represents a display control event.
type DisplayControlEvent interface {
	Type() DisplayControlEventType
	Timestamp() time.Time
}

// DisplayErrorSeverity defines the severity level of an error.
type DisplayErrorSeverity int

const (
	DisplayErrorUnknown DisplayErrorSeverity = iota
	DisplayErrorWarning
	DisplayErrorError
	DisplayErrorFatal
)

// DisplayFrameDroppedEvent signals that a frame was dropped.
type DisplayFrameDroppedEvent struct {
	timestamp time.Time
	FrameID   uint64
	Reason    string
}

// NewDisplayFrameDroppedEvent constructs a DisplayFrameDroppedEvent with a preset timestamp.
func NewDisplayFrameDroppedEvent(frameID uint64, reason string, timestamp time.Time) DisplayFrameDroppedEvent {
	return DisplayFrameDroppedEvent{
		timestamp: timestamp,
		FrameID:   frameID,
		Reason:    reason,
	}
}

// Type returns the event type.
func (e DisplayFrameDroppedEvent) Type() DisplayControlEventType {
	return DisplayControlFrameDropped
}

// Timestamp returns the event timestamp.
func (e DisplayFrameDroppedEvent) Timestamp() time.Time {
	return e.timestamp
}

// DisplayCaptureErrorEvent signals a capture error.
type DisplayCaptureErrorEvent struct {
	timestamp time.Time
	Error     error
	Severity  DisplayErrorSeverity
}

// NewDisplayCaptureErrorEvent constructs a DisplayCaptureErrorEvent with a preset timestamp.
func NewDisplayCaptureErrorEvent(err error, severity DisplayErrorSeverity, timestamp time.Time) DisplayCaptureErrorEvent {
	return DisplayCaptureErrorEvent{
		timestamp: timestamp,
		Error:     err,
		Severity:  severity,
	}
}

// Type returns the event type.
func (e DisplayCaptureErrorEvent) Type() DisplayControlEventType {
	return DisplayControlCaptureError
}

// Timestamp returns the event timestamp.
func (e DisplayCaptureErrorEvent) Timestamp() time.Time {
	return e.timestamp
}

// DisplayMetricsEvent contains display metrics.
type DisplayMetricsEvent struct {
	timestamp     time.Time
	FPS           float64
	DroppedFrames uint64
	TotalFrames   uint64
}

// NewDisplayMetricsEvent constructs a DisplayMetricsEvent with a preset timestamp.
func NewDisplayMetricsEvent(fps float64, droppedFrames, totalFrames uint64, timestamp time.Time) DisplayMetricsEvent {
	return DisplayMetricsEvent{
		timestamp:     timestamp,
		FPS:           fps,
		DroppedFrames: droppedFrames,
		TotalFrames:   totalFrames,
	}
}

// Type returns the event type.
func (e DisplayMetricsEvent) Type() DisplayControlEventType {
	return DisplayControlMetrics
}

// Timestamp returns the event timestamp.
func (e DisplayMetricsEvent) Timestamp() time.Time {
	return e.timestamp
}

// DisplayModeChangedEvent signals that the display mode has changed.
type DisplayModeChangedEvent struct {
	timestamp time.Time
	OldMode   DisplayMode
	NewMode   DisplayMode
}

// NewDisplayModeChangedEvent constructs a DisplayModeChangedEvent with a preset timestamp.
func NewDisplayModeChangedEvent(oldMode, newMode DisplayMode, timestamp time.Time) DisplayModeChangedEvent {
	return DisplayModeChangedEvent{
		timestamp: timestamp,
		OldMode:   oldMode,
		NewMode:   newMode,
	}
}

// Type returns the event type.
func (e DisplayModeChangedEvent) Type() DisplayControlEventType {
	return DisplayControlModeChanged
}

// Timestamp returns the event timestamp.
func (e DisplayModeChangedEvent) Timestamp() time.Time {
	return e.timestamp
}

// DisplaySourceStartedEvent signals that a display source has started.
type DisplaySourceStartedEvent struct {
	timestamp time.Time
}

// NewDisplaySourceStartedEvent constructs a DisplaySourceStartedEvent with a preset timestamp.
func NewDisplaySourceStartedEvent(timestamp time.Time) DisplaySourceStartedEvent {
	return DisplaySourceStartedEvent{timestamp: timestamp}
}

// Type returns the event type.
func (e DisplaySourceStartedEvent) Type() DisplayControlEventType {
	return DisplayControlSourceStarted
}

// Timestamp returns the event timestamp.
func (e DisplaySourceStartedEvent) Timestamp() time.Time {
	return e.timestamp
}

// DisplaySourceStoppedEvent signals that a display source has stopped.
type DisplaySourceStoppedEvent struct {
	timestamp time.Time
}

// NewDisplaySourceStoppedEvent constructs a DisplaySourceStoppedEvent with a preset timestamp.
func NewDisplaySourceStoppedEvent(timestamp time.Time) DisplaySourceStoppedEvent {
	return DisplaySourceStoppedEvent{timestamp: timestamp}
}

// Type returns the event type.
func (e DisplaySourceStoppedEvent) Type() DisplayControlEventType {
	return DisplayControlSourceStopped
}

// Timestamp returns the event timestamp.
func (e DisplaySourceStoppedEvent) Timestamp() time.Time {
	return e.timestamp
}

// DisplaySinkStartedEvent signals that a display sink has started.
type DisplaySinkStartedEvent struct {
	timestamp time.Time
}

// NewDisplaySinkStartedEvent constructs a DisplaySinkStartedEvent with a preset timestamp.
func NewDisplaySinkStartedEvent(timestamp time.Time) DisplaySinkStartedEvent {
	return DisplaySinkStartedEvent{timestamp: timestamp}
}

// Type returns the event type.
func (e DisplaySinkStartedEvent) Type() DisplayControlEventType {
	return DisplayControlSinkStarted
}

// Timestamp returns the event timestamp.
func (e DisplaySinkStartedEvent) Timestamp() time.Time {
	return e.timestamp
}

// DisplaySinkStoppedEvent signals that a display sink has stopped.
type DisplaySinkStoppedEvent struct {
	timestamp time.Time
}

// NewDisplaySinkStoppedEvent constructs a DisplaySinkStoppedEvent with a preset timestamp.
func NewDisplaySinkStoppedEvent(timestamp time.Time) DisplaySinkStoppedEvent {
	return DisplaySinkStoppedEvent{timestamp: timestamp}
}

// Type returns the event type.
func (e DisplaySinkStoppedEvent) Type() DisplayControlEventType {
	return DisplayControlSinkStopped
}

// Timestamp returns the event timestamp.
func (e DisplaySinkStoppedEvent) Timestamp() time.Time {
	return e.timestamp
}

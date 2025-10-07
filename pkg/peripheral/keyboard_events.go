package peripheral

import "time"

// KeyboardEventType defines the type of keyboard data event.
type KeyboardEventType int

const (
	// KeyboardEventUnknown represents an uninitialized or invalid keyboard event type.
	KeyboardEventUnknown KeyboardEventType = iota
	// KeyboardEventKey represents a keyboard key event.
	KeyboardEventKey
)

// KeyboardEvent represents a keyboard data event.
type KeyboardEvent interface {
	Type() KeyboardEventType
	Timestamp() time.Time
}

// KeyboardKeyState captures the transition of a key.
type KeyboardKeyState int

const (
	// KeyboardKeyStateUnknown represents an uninitialized or invalid key state.
	KeyboardKeyStateUnknown KeyboardKeyState = iota
	// KeyboardKeyStatePress represents a key press action.
	KeyboardKeyStatePress
	// KeyboardKeyStateRelease represents a key release action.
	KeyboardKeyStateRelease
	// KeyboardKeyStateRepeat represents a key repeat action.
	KeyboardKeyStateRepeat
)

// KeyboardModifiers represents the active modifier set as a bitmask.
type KeyboardModifiers uint32

const (
	// KeyboardModifierNone indicates no modifiers are active.
	KeyboardModifierNone KeyboardModifiers = 0
	// KeyboardModifierShift indicates the Shift modifier is active.
	KeyboardModifierShift KeyboardModifiers = 1 << iota
	// KeyboardModifierControl indicates the Control modifier is active.
	KeyboardModifierControl
	// KeyboardModifierAlt indicates the Alt modifier is active.
	KeyboardModifierAlt
	// KeyboardModifierMeta indicates the Meta (Windows/Command) modifier is active.
	KeyboardModifierMeta
	// KeyboardModifierAltGr indicates the AltGr modifier is active.
	KeyboardModifierAltGr
	// KeyboardModifierFunction indicates the Function modifier is active.
	KeyboardModifierFunction
)

// KeyboardKeyEvent encapsulates a physical key transition along with logical context.
type KeyboardKeyEvent struct {
	timestamp        time.Time
	HIDUsage         KeyboardHIDUsage
	PhysicalScanCode string
	LogicalKey       KeyboardLogicalKey
	Modifiers        KeyboardModifiers
	State            KeyboardKeyState
	Text             string
	SourceID         string
}

// NewKeyboardKeyEvent constructs a KeyboardKeyEvent with an explicit timestamp.
func NewKeyboardKeyEvent(hid KeyboardHIDUsage, scanCode string, logicalKey KeyboardLogicalKey, modifiers KeyboardModifiers, state KeyboardKeyState, text string, sourceID string, timestamp time.Time) KeyboardKeyEvent {
	return KeyboardKeyEvent{
		timestamp:        timestamp,
		HIDUsage:         hid,
		PhysicalScanCode: scanCode,
		LogicalKey:       logicalKey,
		Modifiers:        modifiers,
		State:            state,
		Text:             text,
		SourceID:         sourceID,
	}
}

// Type returns the event type.
func (e KeyboardKeyEvent) Type() KeyboardEventType {
	return KeyboardEventKey
}

// Timestamp returns the event timestamp.
func (e KeyboardKeyEvent) Timestamp() time.Time {
	return e.timestamp
}

// KeyboardControlEventType defines the type of keyboard control event.
type KeyboardControlEventType int

const (
	// KeyboardControlUnknown represents an uninitialized or invalid control event type.
	KeyboardControlUnknown KeyboardControlEventType = iota
	// KeyboardControlError signals a keyboard error event.
	KeyboardControlError
	// KeyboardControlMetrics signals keyboard metrics information.
	KeyboardControlMetrics
	// KeyboardControlLayoutChanged signals that the keyboard layout has changed.
	KeyboardControlLayoutChanged
	// KeyboardControlLEDStateChanged signals that the LED state has changed.
	KeyboardControlLEDStateChanged
	// KeyboardControlSourceStarted signals that a keyboard source has started.
	KeyboardControlSourceStarted
	// KeyboardControlSourceStopped signals that a keyboard source has stopped.
	KeyboardControlSourceStopped
	// KeyboardControlSinkStarted signals that a keyboard sink has started.
	KeyboardControlSinkStarted
	// KeyboardControlSinkStopped signals that a keyboard sink has stopped.
	KeyboardControlSinkStopped
)

// KeyboardControlEvent represents a keyboard control event.
type KeyboardControlEvent interface {
	Type() KeyboardControlEventType
	Timestamp() time.Time
}

// KeyboardErrorSeverity defines severity levels for keyboard errors.
type KeyboardErrorSeverity int

const (
	// KeyboardErrorUnknown represents an uninitialized or invalid error severity.
	KeyboardErrorUnknown KeyboardErrorSeverity = iota
	// KeyboardErrorWarning represents a warning that does not stop operation.
	KeyboardErrorWarning
	// KeyboardErrorRecoverable represents a recoverable error.
	KeyboardErrorRecoverable
	// KeyboardErrorFatal represents an unrecoverable error.
	KeyboardErrorFatal
)

// KeyboardErrorEvent signals an error encountered by a keyboard peripheral.
type KeyboardErrorEvent struct {
	timestamp time.Time
	Error     error
	Severity  KeyboardErrorSeverity
	SourceID  string
}

// NewKeyboardErrorEvent constructs a KeyboardErrorEvent with a preset timestamp.
func NewKeyboardErrorEvent(err error, severity KeyboardErrorSeverity, sourceID string, timestamp time.Time) KeyboardErrorEvent {
	return KeyboardErrorEvent{
		timestamp: timestamp,
		Error:     err,
		Severity:  severity,
		SourceID:  sourceID,
	}
}

// Type returns the event type.
func (e KeyboardErrorEvent) Type() KeyboardControlEventType {
	return KeyboardControlError
}

// Timestamp returns the event timestamp.
func (e KeyboardErrorEvent) Timestamp() time.Time {
	return e.timestamp
}

// KeyboardMetricsEvent captures high-level metrics about keyboard streams.
type KeyboardMetricsEvent struct {
	timestamp        time.Time
	TotalEvents      uint64
	DroppedEvents    uint64
	AverageLatencyMs float64
	SourceID         string
}

// NewKeyboardMetricsEvent constructs a KeyboardMetricsEvent with a preset timestamp.
func NewKeyboardMetricsEvent(totalEvents, droppedEvents uint64, avgLatencyMs float64, sourceID string, timestamp time.Time) KeyboardMetricsEvent {
	return KeyboardMetricsEvent{
		timestamp:        timestamp,
		TotalEvents:      totalEvents,
		DroppedEvents:    droppedEvents,
		AverageLatencyMs: avgLatencyMs,
		SourceID:         sourceID,
	}
}

// Type returns the event type.
func (e KeyboardMetricsEvent) Type() KeyboardControlEventType {
	return KeyboardControlMetrics
}

// Timestamp returns the event timestamp.
func (e KeyboardMetricsEvent) Timestamp() time.Time {
	return e.timestamp
}

// KeyboardLayoutChangedEvent notifies listeners about layout changes.
type KeyboardLayoutChangedEvent struct {
	timestamp time.Time
	OldLayout KeyboardLayout
	NewLayout KeyboardLayout
	Reason    string
	SourceID  string
}

// NewKeyboardLayoutChangedEvent constructs a KeyboardLayoutChangedEvent with a preset timestamp.
func NewKeyboardLayoutChangedEvent(oldLayout, newLayout KeyboardLayout, reason string, sourceID string, timestamp time.Time) KeyboardLayoutChangedEvent {
	return KeyboardLayoutChangedEvent{
		timestamp: timestamp,
		OldLayout: oldLayout,
		NewLayout: newLayout,
		Reason:    reason,
		SourceID:  sourceID,
	}
}

// Type returns the event type.
func (e KeyboardLayoutChangedEvent) Type() KeyboardControlEventType {
	return KeyboardControlLayoutChanged
}

// Timestamp returns the event timestamp.
func (e KeyboardLayoutChangedEvent) Timestamp() time.Time {
	return e.timestamp
}

// KeyboardLEDStateChangedEvent reports LED state updates coming from a sink.
type KeyboardLEDStateChangedEvent struct {
	timestamp time.Time
	State     KeyboardLEDState
	SourceID  string
}

// NewKeyboardLEDStateChangedEvent constructs a KeyboardLEDStateChangedEvent with a preset timestamp.
func NewKeyboardLEDStateChangedEvent(state KeyboardLEDState, sourceID string, timestamp time.Time) KeyboardLEDStateChangedEvent {
	return KeyboardLEDStateChangedEvent{
		timestamp: timestamp,
		State:     state,
		SourceID:  sourceID,
	}
}

// Type returns the event type.
func (e KeyboardLEDStateChangedEvent) Type() KeyboardControlEventType {
	return KeyboardControlLEDStateChanged
}

// Timestamp returns the event timestamp.
func (e KeyboardLEDStateChangedEvent) Timestamp() time.Time {
	return e.timestamp
}

// KeyboardSourceStartedEvent signals that a keyboard source has started.
type KeyboardSourceStartedEvent struct {
	timestamp time.Time
	SourceID  string
}

// NewKeyboardSourceStartedEvent constructs a KeyboardSourceStartedEvent with a preset timestamp.
func NewKeyboardSourceStartedEvent(sourceID string, timestamp time.Time) KeyboardSourceStartedEvent {
	return KeyboardSourceStartedEvent{
		timestamp: timestamp,
		SourceID:  sourceID,
	}
}

// Type returns the event type.
func (e KeyboardSourceStartedEvent) Type() KeyboardControlEventType {
	return KeyboardControlSourceStarted
}

// Timestamp returns the event timestamp.
func (e KeyboardSourceStartedEvent) Timestamp() time.Time {
	return e.timestamp
}

// KeyboardSourceStoppedEvent signals that a keyboard source has stopped.
type KeyboardSourceStoppedEvent struct {
	timestamp time.Time
	SourceID  string
}

// NewKeyboardSourceStoppedEvent constructs a KeyboardSourceStoppedEvent with a preset timestamp.
func NewKeyboardSourceStoppedEvent(sourceID string, timestamp time.Time) KeyboardSourceStoppedEvent {
	return KeyboardSourceStoppedEvent{
		timestamp: timestamp,
		SourceID:  sourceID,
	}
}

// Type returns the event type.
func (e KeyboardSourceStoppedEvent) Type() KeyboardControlEventType {
	return KeyboardControlSourceStopped
}

// Timestamp returns the event timestamp.
func (e KeyboardSourceStoppedEvent) Timestamp() time.Time {
	return e.timestamp
}

// KeyboardSinkStartedEvent signals that a keyboard sink has started.
type KeyboardSinkStartedEvent struct {
	timestamp time.Time
	SinkID    string
}

// NewKeyboardSinkStartedEvent constructs a KeyboardSinkStartedEvent with a preset timestamp.
func NewKeyboardSinkStartedEvent(sinkID string, timestamp time.Time) KeyboardSinkStartedEvent {
	return KeyboardSinkStartedEvent{
		timestamp: timestamp,
		SinkID:    sinkID,
	}
}

// Type returns the event type.
func (e KeyboardSinkStartedEvent) Type() KeyboardControlEventType {
	return KeyboardControlSinkStarted
}

// Timestamp returns the event timestamp.
func (e KeyboardSinkStartedEvent) Timestamp() time.Time {
	return e.timestamp
}

// KeyboardSinkStoppedEvent signals that a keyboard sink has stopped.
type KeyboardSinkStoppedEvent struct {
	timestamp time.Time
	SinkID    string
}

// NewKeyboardSinkStoppedEvent constructs a KeyboardSinkStoppedEvent with a preset timestamp.
func NewKeyboardSinkStoppedEvent(sinkID string, timestamp time.Time) KeyboardSinkStoppedEvent {
	return KeyboardSinkStoppedEvent{
		timestamp: timestamp,
		SinkID:    sinkID,
	}
}

// Type returns the event type.
func (e KeyboardSinkStoppedEvent) Type() KeyboardControlEventType {
	return KeyboardControlSinkStopped
}

// Timestamp returns the event timestamp.
func (e KeyboardSinkStoppedEvent) Timestamp() time.Time {
	return e.timestamp
}

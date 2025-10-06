package peripheral

import (
	"errors"
	"fmt"
)

// PeripheralType defines the category of peripheral device.
type PeripheralType int

const (
	PeripheralTypeUnknown PeripheralType = iota
	PeripheralTypeDisplay
	PeripheralTypeKeyboard
	PeripheralTypeMouse
)

// String returns the string representation of the peripheral type.
func (pt PeripheralType) String() string {
	switch pt {
	case PeripheralTypeDisplay:
		return "display"
	case PeripheralTypeKeyboard:
		return "keyboard"
	case PeripheralTypeMouse:
		return "mouse"
	case PeripheralTypeUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// PeripheralRole defines whether a peripheral is a source or sink.
type PeripheralRole int

const (
	PeripheralRoleUnknown PeripheralRole = iota
	PeripheralRoleSource
	PeripheralRoleSink
)

// String returns the string representation of the peripheral role.
func (pr PeripheralRole) String() string {
	switch pr {
	case PeripheralRoleSource:
		return "source"
	case PeripheralRoleSink:
		return "sink"
	case PeripheralRoleUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// PeripheralID uniquely identifies a peripheral device.
type PeripheralID struct {
	pType    PeripheralType
	role     PeripheralRole
	specific string
}

// NewPeripheralID constructs a new PeripheralID with the given parameters.
// Returns an error if pType is Unknown, role is Unknown, or specific is empty.
func NewPeripheralID(pType PeripheralType, role PeripheralRole, specific string) (PeripheralID, error) {
	if pType == PeripheralTypeUnknown {
		return PeripheralID{}, errors.New("peripheral type cannot be unknown")
	}
	if role == PeripheralRoleUnknown {
		return PeripheralID{}, errors.New("peripheral role cannot be unknown")
	}
	if specific == "" {
		return PeripheralID{}, errors.New("specific identifier cannot be empty")
	}

	return PeripheralID{
		pType:    pType,
		role:     role,
		specific: specific,
	}, nil
}

// Type returns the peripheral type.
func (pid PeripheralID) Type() PeripheralType {
	return pid.pType
}

// Role returns the peripheral role.
func (pid PeripheralID) Role() PeripheralRole {
	return pid.role
}

// Specific returns the device-specific identifier part.
func (pid PeripheralID) Specific() string {
	return pid.specific
}

// String returns the formatted peripheral ID in the form: {type}/{role}/{specific}
func (pid PeripheralID) String() string {
	return fmt.Sprintf("%s/%s/%s", pid.pType, pid.role, pid.specific)
}

// Peripheral is the base interface for all peripheral devices.
type Peripheral interface {
	ID() PeripheralID
}

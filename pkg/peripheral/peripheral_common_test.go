package peripheral

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPeripheralType_String(t *testing.T) {
	assert.Equal(t, "display", PeripheralTypeDisplay.String())
	assert.Equal(t, "keyboard", PeripheralTypeKeyboard.String())
	assert.Equal(t, "mouse", PeripheralTypeMouse.String())
	assert.Equal(t, "unknown", PeripheralTypeUnknown.String())
	assert.Equal(t, "unknown", PeripheralType(999).String())
}

func TestPeripheralRole_String(t *testing.T) {
	assert.Equal(t, "source", PeripheralRoleSource.String())
	assert.Equal(t, "sink", PeripheralRoleSink.String())
	assert.Equal(t, "unknown", PeripheralRoleUnknown.String())
	assert.Equal(t, "unknown", PeripheralRole(999).String())
}

func TestNewPeripheralID(t *testing.T) {
	pid, err := NewPeripheralID(PeripheralTypeDisplay, PeripheralRoleSource, "HDMI-1")
	assert.NoError(t, err)
	assert.Equal(t, PeripheralTypeDisplay, pid.pType)
	assert.Equal(t, PeripheralRoleSource, pid.role)
	assert.Equal(t, "HDMI-1", pid.specific)
}

func TestPeripheralID_Getters(t *testing.T) {
	pid, err := NewPeripheralID(PeripheralTypeDisplay, PeripheralRoleSource, "HDMI-1")
	assert.NoError(t, err)

	assert.Equal(t, PeripheralTypeDisplay, pid.Type())
	assert.Equal(t, PeripheralRoleSource, pid.Role())
	assert.Equal(t, "HDMI-1", pid.Specific())
}

func TestPeripheralID_String(t *testing.T) {
	pid1, err := NewPeripheralID(PeripheralTypeDisplay, PeripheralRoleSource, "HDMI-1")
	assert.NoError(t, err)
	assert.Equal(t, "display/source/HDMI-1", pid1.String())

	pid2, err := NewPeripheralID(PeripheralTypeKeyboard, PeripheralRoleSink, "USB-KB-001")
	assert.NoError(t, err)
	assert.Equal(t, "keyboard/sink/USB-KB-001", pid2.String())

	pid3, err := NewPeripheralID(PeripheralTypeMouse, PeripheralRoleSource, "BT-Mouse-1")
	assert.NoError(t, err)
	assert.Equal(t, "mouse/source/BT-Mouse-1", pid3.String())
}

func TestNewPeripheralID_ValidationErrors(t *testing.T) {
	_, err := NewPeripheralID(PeripheralTypeUnknown, PeripheralRoleSource, "test")
	assert.Error(t, err)
	assert.Equal(t, "peripheral type cannot be unknown", err.Error())

	_, err = NewPeripheralID(PeripheralTypeDisplay, PeripheralRoleUnknown, "test")
	assert.Error(t, err)
	assert.Equal(t, "peripheral role cannot be unknown", err.Error())

	_, err = NewPeripheralID(PeripheralTypeDisplay, PeripheralRoleSource, "")
	assert.Error(t, err)
	assert.Equal(t, "specific identifier cannot be empty", err.Error())
}

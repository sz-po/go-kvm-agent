package peripheral

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPeripheralName(t *testing.T) {
	t.Run("creates valid peripheral name", func(t *testing.T) {
		name, err := NewPeripheralName("mpv-peripheral")
		assert.NoError(t, err)
		assert.Equal(t, PeripheralName("mpv-peripheral"), name)
	})

	t.Run("returns error for empty name", func(t *testing.T) {
		name, err := NewPeripheralName("")
		assert.Error(t, err)
		assert.Equal(t, PeripheralName(""), name)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("returns error for non-kebab-case name", func(t *testing.T) {
		name, err := NewPeripheralName("TestPeripheral")
		assert.Error(t, err)
		assert.Equal(t, PeripheralName(""), name)
		assert.Contains(t, err.Error(), "must be kebab-case")
	})

	t.Run("returns error for name with underscores", func(t *testing.T) {
		name, err := NewPeripheralName("test_peripheral")
		assert.Error(t, err)
		assert.Equal(t, PeripheralName(""), name)
		assert.Contains(t, err.Error(), "must be kebab-case")
	})
}

func TestNewPeripheralId(t *testing.T) {
	t.Run("creates valid peripheral id", func(t *testing.T) {
		id, err := NewPeripheralId("mpv-peripheral-123")
		assert.NoError(t, err)
		assert.Equal(t, PeripheralId("mpv-peripheral-123"), id)
	})

	t.Run("returns error for empty id", func(t *testing.T) {
		id, err := NewPeripheralId("")
		assert.Error(t, err)
		assert.Equal(t, PeripheralId(""), id)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("returns error for non-kebab-case id", func(t *testing.T) {
		id, err := NewPeripheralId("TestPeripheral123")
		assert.Error(t, err)
		assert.Equal(t, PeripheralId(""), id)
		assert.Contains(t, err.Error(), "should be kebab-case")
	})

	t.Run("returns error for id with underscores", func(t *testing.T) {
		id, err := NewPeripheralId("test_peripheral_123")
		assert.Error(t, err)
		assert.Equal(t, PeripheralId(""), id)
		assert.Contains(t, err.Error(), "should be kebab-case")
	})
}

func TestNewPeripheralRandomId(t *testing.T) {
	t.Run("generates id with kebab-case prefix", func(t *testing.T) {
		id := CreatePeripheralRandomId("TestPrefix")
		assert.NotEmpty(t, id)
		assert.True(t, strings.HasPrefix(string(id), "test-prefix-"))
	})

	t.Run("generates unique ids", func(t *testing.T) {
		id1 := CreatePeripheralRandomId("mpv")
		id2 := CreatePeripheralRandomId("mpv")
		assert.NotEqual(t, id1, id2)
	})

	t.Run("generates valid uuid format", func(t *testing.T) {
		id := CreatePeripheralRandomId("mpv")
		parts := strings.Split(string(id), "-")
		// Format: mpv-{uuid} where uuid has 5 parts separated by hyphens
		assert.Greater(t, len(parts), 1, "id should contain at least prefix and uuid parts")
	})
}

func TestPeripheralKind_String(t *testing.T) {
	t.Run("returns string representation", func(t *testing.T) {
		assert.Equal(t, "display", PeripheralKindDisplay.String())
		assert.Equal(t, "keyboard", PeripheralKindKeyboard.String())
		assert.Equal(t, "mouse", PeripheralKindMouse.String())
		assert.Equal(t, "", PeripheralKindUnknown.String())
	})
}

func TestPeripheralRole_String(t *testing.T) {
	t.Run("returns string representation", func(t *testing.T) {
		assert.Equal(t, "source", PeripheralRoleSource.String())
		assert.Equal(t, "sink", PeripheralRoleSink.String())
		assert.Equal(t, "", PeripheralRoleUnknown.String())
	})
}

func TestPeripheralName_String(t *testing.T) {
	t.Run("returns string representation", func(t *testing.T) {
		name := PeripheralName("mpv-name")
		assert.Equal(t, "mpv-name", name.String())
	})
}

func TestPeripheralId_String(t *testing.T) {
	t.Run("returns string representation", func(t *testing.T) {
		id := PeripheralId("mpv-id-123")
		assert.Equal(t, "mpv-id-123", id.String())
	})
}

func TestNewCapability(t *testing.T) {
	t.Run("creates capability with kind and role", func(t *testing.T) {
		capability := NewCapability[DisplaySource](PeripheralKindDisplay, PeripheralRoleSource)
		assert.Equal(t, PeripheralKindDisplay, capability.Kind)
		assert.Equal(t, PeripheralRoleSource, capability.Role)
		assert.NotNil(t, capability.validationFn)
	})

	t.Run("creates capability for display sink", func(t *testing.T) {
		capability := NewCapability[DisplaySink](PeripheralKindDisplay, PeripheralRoleSink)
		assert.Equal(t, PeripheralKindDisplay, capability.Kind)
		assert.Equal(t, PeripheralRoleSink, capability.Role)
		assert.NotNil(t, capability.validationFn)
	})
}

func TestPeripheralCapability_String(t *testing.T) {
	t.Run("formats as kind-role", func(t *testing.T) {
		capability := PeripheralCapability{
			Kind: PeripheralKindDisplay,
			Role: PeripheralRoleSource,
		}
		assert.Equal(t, "display-source", capability.String())
	})

	t.Run("formats display sink capability", func(t *testing.T) {
		capability := PeripheralCapability{
			Kind: PeripheralKindDisplay,
			Role: PeripheralRoleSink,
		}
		assert.Equal(t, "display-sink", capability.String())
	})

	t.Run("formats keyboard source capability", func(t *testing.T) {
		capability := PeripheralCapability{
			Kind: PeripheralKindKeyboard,
			Role: PeripheralRoleSource,
		}
		assert.Equal(t, "keyboard-source", capability.String())
	})
}

func TestPeripheralCapability_Validate(t *testing.T) {
	t.Run("validates display source implements interface", func(t *testing.T) {
		capability := NewCapability[DisplaySource](PeripheralKindDisplay, PeripheralRoleSource)
		displaySource := NewDisplaySourceMock(t)

		err := capability.Validate(displaySource)
		assert.NoError(t, err)
	})

	t.Run("returns error when peripheral does not implement interface", func(t *testing.T) {
		capability := NewCapability[DisplaySource](PeripheralKindDisplay, PeripheralRoleSource)
		id := PeripheralId("mpv-sink")
		displaySink := NewDisplaySinkMock(t)
		displaySink.EXPECT().GetId().Return(id)

		err := capability.Validate(displaySink)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not implement")
		assert.Contains(t, err.Error(), "mpv-sink")
	})

	t.Run("validates display sink implements interface", func(t *testing.T) {
		capability := NewCapability[DisplaySink](PeripheralKindDisplay, PeripheralRoleSink)
		displaySink := NewDisplaySinkMock(t)

		err := capability.Validate(displaySink)
		assert.NoError(t, err)
	})

	t.Run("returns error when display sink does not implement display source", func(t *testing.T) {
		capability := NewCapability[DisplaySource](PeripheralKindDisplay, PeripheralRoleSource)
		id := PeripheralId("mpv-sink")
		displaySink := NewDisplaySinkMock(t)
		displaySink.EXPECT().GetId().Return(id)

		err := capability.Validate(displaySink)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not implement")
	})
}

func TestValidatePeripheralCapabilities(t *testing.T) {
	t.Run("validates all capabilities successfully", func(t *testing.T) {
		displaySource := NewDisplaySourceMock(t)
		displaySource.EXPECT().GetCapabilities().Return([]PeripheralCapability{DisplaySourceCapability})

		err := ValidatePeripheralCapabilities(displaySource)
		assert.NoError(t, err)
	})

	t.Run("validates keyboard source capability", func(t *testing.T) {
		keyboardSource := NewKeyboardSourceMock(t)
		keyboardSource.EXPECT().GetCapabilities().Return([]PeripheralCapability{KeyboardSourceCapability})

		err := ValidatePeripheralCapabilities(keyboardSource)
		assert.NoError(t, err)
	})

	t.Run("returns error when keyboard source capability validation fails", func(t *testing.T) {
		keyboardSourceID := PeripheralId("keyboard-source")
		capability := NewCapability[KeyboardSource](PeripheralKindKeyboard, PeripheralRoleSource)

		mockedPeripheral := NewPeripheralMock(t)
		mockedPeripheral.EXPECT().GetId().Return(keyboardSourceID)
		mockedPeripheral.EXPECT().GetCapabilities().Return([]PeripheralCapability{capability})

		err := ValidatePeripheralCapabilities(mockedPeripheral)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "keyboard-source")
		assert.Contains(t, err.Error(), capability.String())
	})

	t.Run("validates keyboard sink capability", func(t *testing.T) {
		keyboardSink := NewKeyboardSinkMock(t)
		keyboardSink.EXPECT().GetCapabilities().Return([]PeripheralCapability{KeyboardSinkCapability})

		err := ValidatePeripheralCapabilities(keyboardSink)
		assert.NoError(t, err)
	})

	t.Run("returns error when keyboard sink capability validation fails", func(t *testing.T) {
		keyboardSinkID := PeripheralId("keyboard-sink")
		capability := NewCapability[KeyboardSink](PeripheralKindKeyboard, PeripheralRoleSink)

		mockedPeripheral := NewPeripheralMock(t)
		mockedPeripheral.EXPECT().GetId().Return(keyboardSinkID)
		mockedPeripheral.EXPECT().GetCapabilities().Return([]PeripheralCapability{capability})

		err := ValidatePeripheralCapabilities(mockedPeripheral)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "keyboard-sink")
		assert.Contains(t, err.Error(), capability.String())
	})

	t.Run("validates mouse source capability", func(t *testing.T) {
		mouseSource := NewMouseSourceMock(t)
		mouseSource.EXPECT().GetCapabilities().Return([]PeripheralCapability{MouseSourceCapability})

		err := ValidatePeripheralCapabilities(mouseSource)
		assert.NoError(t, err)
	})

	t.Run("validates mouse sink capability", func(t *testing.T) {
		mouseSink := NewMouseSinkMock(t)
		mouseSink.EXPECT().GetCapabilities().Return([]PeripheralCapability{MouseSinkCapability})

		err := ValidatePeripheralCapabilities(mouseSink)
		assert.NoError(t, err)
	})

	t.Run("returns error when capability validation fails", func(t *testing.T) {
		id := PeripheralId("mpv-sink")

		// Create a capability that expects DisplaySource but give it a DisplaySink
		capability := NewCapability[DisplaySource](PeripheralKindDisplay, PeripheralRoleSource)

		// Create a mock peripheral that claims to have DisplaySource capability but is actually a DisplaySink
		mockedPeripheral := NewPeripheralMock(t)
		mockedPeripheral.EXPECT().GetId().Return(id)
		mockedPeripheral.EXPECT().GetCapabilities().Return([]PeripheralCapability{capability})

		err := ValidatePeripheralCapabilities(mockedPeripheral)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mpv-sink")
		assert.Contains(t, err.Error(), "display-source")
		assert.Contains(t, err.Error(), "capability")
	})

	t.Run("validates peripheral with no capabilities", func(t *testing.T) {
		mockedPeripheral := NewPeripheralMock(t)
		mockedPeripheral.EXPECT().GetCapabilities().Return([]PeripheralCapability{})

		err := ValidatePeripheralCapabilities(mockedPeripheral)
		assert.NoError(t, err)
	})
}

package machine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	internalPeripheral "github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/peripheral"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

// mockPeripheral is a test implementation of the Peripheral interface.
type mockPeripheral struct {
	id peripheralSDK.PeripheralID
}

func (mp *mockPeripheral) ID() peripheralSDK.PeripheralID {
	return mp.id
}

func TestMachine_GetPeripherals_Empty(t *testing.T) {
	machineName, err := NewMachineName("test-machine")
	if !assert.NoError(t, err, "NewMachineName should succeed.") {
		return
	}

	machine := &Machine{
		name:        machineName,
		peripherals: make(map[peripheralSDK.PeripheralID]peripheralSDK.Peripheral),
	}

	peripherals := machine.GetPeripherals()
	assert.Len(t, peripherals, 0, "GetPeripherals should return no peripherals for an empty machine.")
}

func TestMachine_GetPeripherals_WithDevices(t *testing.T) {
	machineName, err := NewMachineName("test-machine")
	if !assert.NoError(t, err, "NewMachineName should succeed.") {
		return
	}

	id1, err := peripheralSDK.NewPeripheralID(peripheralSDK.PeripheralTypeDisplay, peripheralSDK.PeripheralRoleSource, "hdmi0")
	if !assert.NoError(t, err, "Creating first peripheral ID should succeed.") {
		return
	}

	id2, err := peripheralSDK.NewPeripheralID(peripheralSDK.PeripheralTypeKeyboard, peripheralSDK.PeripheralRoleSource, "usb-kbd0")
	if !assert.NoError(t, err, "Creating second peripheral ID should succeed.") {
		return
	}

	mock1 := &mockPeripheral{id: id1}
	mock2 := &mockPeripheral{id: id2}

	machine := &Machine{
		name: machineName,
		peripherals: map[peripheralSDK.PeripheralID]peripheralSDK.Peripheral{
			id1: mock1,
			id2: mock2,
		},
	}

	peripherals := machine.GetPeripherals()
	assert.Len(t, peripherals, 2, "GetPeripherals should return all registered peripherals.")

	// Verify returned slice is a copy (modifying it should not affect the machine)
	originalLen := len(machine.peripherals)
	peripherals = append(peripherals, nil)
	assert.Equal(t, originalLen, len(machine.peripherals), "Modifying the returned slice should not change machine peripherals.")
}

func TestMachine_GetPeripheralByID_Found(t *testing.T) {
	machineName, err := NewMachineName("test-machine")
	if !assert.NoError(t, err, "NewMachineName should succeed.") {
		return
	}

	id, err := peripheralSDK.NewPeripheralID(peripheralSDK.PeripheralTypeDisplay, peripheralSDK.PeripheralRoleSource, "hdmi0")
	if !assert.NoError(t, err, "Creating peripheral ID should succeed.") {
		return
	}

	mock := &mockPeripheral{id: id}

	machine := &Machine{
		name: machineName,
		peripherals: map[peripheralSDK.PeripheralID]peripheralSDK.Peripheral{
			id: mock,
		},
	}

	result, err := machine.GetPeripheralByID(id)
	assert.NoError(t, err, "GetPeripheralByID should find existing peripheral.")
	if !assert.NotNil(t, result, "GetPeripheralByID should return a peripheral instance.") {
		return
	}
	assert.Equal(t, id, result.ID(), "GetPeripheralByID should return the requested peripheral ID.")
}

func TestMachine_GetPeripheralByID_NotFound(t *testing.T) {
	machineName, err := NewMachineName("test-machine")
	if !assert.NoError(t, err, "NewMachineName should succeed.") {
		return
	}

	id1, err := peripheralSDK.NewPeripheralID(peripheralSDK.PeripheralTypeDisplay, peripheralSDK.PeripheralRoleSource, "hdmi0")
	if !assert.NoError(t, err, "Creating existing peripheral ID should succeed.") {
		return
	}

	id2, err := peripheralSDK.NewPeripheralID(peripheralSDK.PeripheralTypeKeyboard, peripheralSDK.PeripheralRoleSource, "usb-kbd0")
	if !assert.NoError(t, err, "Creating missing peripheral ID should succeed.") {
		return
	}

	mock := &mockPeripheral{id: id1}

	machine := &Machine{
		name: machineName,
		peripherals: map[peripheralSDK.PeripheralID]peripheralSDK.Peripheral{
			id1: mock,
		},
	}

	result, err := machine.GetPeripheralByID(id2)
	assert.Error(t, err, "GetPeripheralByID should fail for missing peripheral.")
	assert.ErrorIs(t, err, ErrPeripheralNotFound, "GetPeripheralByID should wrap ErrPeripheralNotFound.")
	assert.Nil(t, result, "GetPeripheralByID should return nil peripheral when not found.")
}

func TestMachine_CreateMachineFromConfig(t *testing.T) {
	machineName, err := NewMachineName("test-machine")
	if !assert.NoError(t, err, "NewMachineName should succeed.") {
		return
	}

	config := &MachineConfig{
		Name:        machineName,
		Peripherals: []internalPeripheral.PeripheralConfig{},
	}

	machine, err := CreateMachineFromConfig(config)

	assert.NoError(t, err, "CreateMachineFromConfig should succeed for valid config.")
	if !assert.NotNil(t, machine, "CreateMachineFromConfig should return a machine instance.") {
		return
	}

	assert.Equal(t, machineName.String(), machine.Name(), "Created machine should have the configured name.")
	assert.NotNil(t, machine.peripherals, "Created machine should initialize peripherals map.")
	assert.Empty(t, machine.GetPeripherals(), "Created machine should have no peripherals initially.")
}

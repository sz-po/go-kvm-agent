package machine

import (
	"context"
	"errors"
	"fmt"

	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

var (
	// ErrPeripheralNotFound indicates that the requested peripheralSDK does not exist in the machine.
	ErrPeripheralNotFound = errors.New("peripheralSDK not found")
)

// MachineOpt is a functional option for configuring a Machine during creation.
type MachineOpt func(*Machine)

// Machine represents a virtual machine instance with its runtime state.
type Machine struct {
	name        MachineName
	peripherals map[peripheralSDK.PeripheralId]peripheralSDK.Peripheral
}

// WithPeripheral returns a MachineOpt that adds a peripheral to the machine.
func WithPeripheral(peripheral peripheralSDK.Peripheral) MachineOpt {
	return func(machine *Machine) {
		machine.peripherals[peripheral.Id()] = peripheral
	}
}

// NewMachine creates a new Machine instance with the given name and options.
func NewMachine(name MachineName, opts ...MachineOpt) *Machine {
	machine := &Machine{
		name:        name,
		peripherals: make(map[peripheralSDK.PeripheralId]peripheralSDK.Peripheral),
	}

	for _, opt := range opts {
		opt(machine)
	}

	return machine
}

// Name returns the machine name as a string.
func (machine *Machine) Name() string {
	return machine.name.String()
}

// GetPeripherals returns all peripherals attached to this machine.
// The returned slice is a copy and modifications to it will not affect the machine's state.
func (machine *Machine) GetPeripherals() []peripheralSDK.Peripheral {
	peripherals := make([]peripheralSDK.Peripheral, 0, len(machine.peripherals))
	for _, p := range machine.peripherals {
		peripherals = append(peripherals, p)
	}
	return peripherals
}

// GetPeripheralByID returns a specific peripheralSDK by its Name.
// Returns ErrPeripheralNotFound if the peripheralSDK does not exist.
func (machine *Machine) GetPeripheralByID(id peripheralSDK.PeripheralId) (peripheralSDK.Peripheral, error) {
	p, ok := machine.peripherals[id]
	if !ok {
		return nil, ErrPeripheralNotFound
	}
	return p, nil
}

func (machine *Machine) Terminate(ctx context.Context) error {
	for _, peripheral := range machine.peripherals {
		err := peripheral.Terminate(ctx)
		if err != nil {
			return fmt.Errorf("peripheral %s terminate failed: %w", peripheral.Id(), err)
		}
	}

	return nil
}

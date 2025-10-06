package machine

import (
	"errors"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

var (
	// ErrPeripheralNotFound indicates that the requested peripheral does not exist in the machine.
	ErrPeripheralNotFound = errors.New("peripheral not found")
)

// Machine represents a virtual machine instance with its runtime state.
type Machine struct {
	name        MachineName
	peripherals map[peripheral.PeripheralID]peripheral.Peripheral
}

// CreateMachineFromConfig creates a new Machine instance from the provided configuration.
func CreateMachineFromConfig(config *MachineConfig) (*Machine, error) {
	return &Machine{
		name:        config.Name,
		peripherals: make(map[peripheral.PeripheralID]peripheral.Peripheral),
	}, nil
}

// Name returns the machine name as a string.
func (machine *Machine) Name() string {
	return machine.name.String()
}

// GetPeripherals returns all peripherals attached to this machine.
// The returned slice is a copy and modifications to it will not affect the machine's state.
func (machine *Machine) GetPeripherals() []peripheral.Peripheral {
	peripherals := make([]peripheral.Peripheral, 0, len(machine.peripherals))
	for _, p := range machine.peripherals {
		peripherals = append(peripherals, p)
	}
	return peripherals
}

// GetPeripheralByID returns a specific peripheral by its ID.
// Returns ErrPeripheralNotFound if the peripheral does not exist.
func (machine *Machine) GetPeripheralByID(id peripheral.PeripheralID) (peripheral.Peripheral, error) {
	p, ok := machine.peripherals[id]
	if !ok {
		return nil, ErrPeripheralNotFound
	}
	return p, nil
}

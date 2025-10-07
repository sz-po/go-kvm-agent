package machine

import (
	"fmt"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/peripheral"
)

// CreateMachineFromConfig creates a new Machine instance from the provided configuration.
func CreateMachineFromConfig(config *MachineConfig) (*Machine, error) {
	machineOpts := []MachineOpt{}

	for _, peripheralConfig := range config.Peripherals {
		machinePeripheral, err := peripheral.CreatePeripheralFromConfig(peripheralConfig)
		if err != nil {
			return nil, fmt.Errorf("error creating peripheral: %w", err)
		}

		machineOpts = append(machineOpts, WithPeripheral(machinePeripheral))
	}

	return NewMachine(config.Name, machineOpts...), nil
}

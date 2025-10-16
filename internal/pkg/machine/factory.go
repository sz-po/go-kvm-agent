package machine

import (
	"context"
	"fmt"

	peripheralInternal "github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/peripheral"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

// CreateMachineFromConfig creates a new LocalMachine instance from the provided configuration.
func CreateMachineFromConfig(ctx context.Context, config MachineConfig) (*LocalMachine, error) {
	if config.Local != nil {
		return createLocalMachineFromConfig(ctx, config.Name, *config.Local)
	} else if config.Remote != nil {
		return createRemoteMachineFromConfig(ctx, config.Name, *config.Remote)
	} else {
		return nil, fmt.Errorf("machine type must be local or remote")
	}
}

func createLocalMachineFromConfig(ctx context.Context, name machineSDK.MachineName, config LocalMachineConfig) (*LocalMachine, error) {
	var peripherals []peripheralSDK.Peripheral
	for _, peripheralConfig := range config.Peripherals {
		peripheral, err := peripheralInternal.CreatePeripheralFromConfig(ctx, peripheralConfig)
		if err != nil {
			return nil, fmt.Errorf("error creating peripheralInternal: %w", err)
		}

		peripherals = append(peripherals, peripheral)
	}

	return newLocalMachine(name, peripherals)
}

func createRemoteMachineFromConfig(ctx context.Context, name machineSDK.MachineName, config RemoteMachineConfig) (*LocalMachine, error) {
	panic("implement me")
}

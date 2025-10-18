package machine

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	peripheralInternal "github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/peripheral"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type LocalMachineConfig struct {
	Peripherals []peripheralInternal.PeripheralConfig `json:"peripherals"`
}

// LocalMachine represents a machine instance with its runtime state.
type LocalMachine struct {
	name        machineSDK.MachineName
	id          machineSDK.MachineId
	peripherals peripheralSDK.Repository
}

var _ machineSDK.Machine = (*LocalMachine)(nil)

// newLocalMachine creates a new LocalMachine instance with the given name and options.
func newLocalMachine(name machineSDK.MachineName, peripherals []peripheralSDK.Peripheral) (*LocalMachine, error) {
	id := machineSDK.MachineId(uuid.NewString())

	peripheralRepository, err := peripheralInternal.NewLocalRepository(peripheralInternal.WithPeripherals(peripherals...))
	if err != nil {
		return nil, fmt.Errorf("create local peripheral repository: %w", err)
	}

	machine := &LocalMachine{
		name:        name,
		id:          id,
		peripherals: peripheralRepository,
	}

	return machine, nil
}

func (machine *LocalMachine) GetName() machineSDK.MachineName {
	return machine.name
}

func (machine *LocalMachine) GetId() machineSDK.MachineId {
	return machine.id
}

func (machine *LocalMachine) Peripherals() peripheralSDK.Repository {
	return machine.peripherals
}

func (machine *LocalMachine) Terminate(ctx context.Context) error {
	peripherals, err := machine.peripherals.GetAllPeripherals(ctx)
	if err != nil {
		return err
	}

	for _, peripheral := range peripherals {
		err = peripheral.Terminate(ctx)
		if err != nil {
			return fmt.Errorf("peripheralInternal %s terminate: %w", peripheral.GetId(), err)
		}
	}

	return nil
}

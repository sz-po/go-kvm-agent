package machine

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/peripheral"
)

// CreateMachineFromConfig creates a new Machine instance from the provided configuration.
func CreateMachineFromConfig(ctx context.Context, config MachineConfig) (*Machine, error) {
	machineOpts := []MachineOpt{}

	for _, peripheralConfig := range config.Peripherals {
		machinePeripheral, err := peripheral.CreatePeripheralFromConfig(ctx, peripheralConfig)
		if err != nil {
			return nil, fmt.Errorf("error creating peripheral: %w", err)
		}

		slog.Info("Peripheral created.",
			slog.String("machineName", config.Name.String()),
			slog.String("peripheralId", machinePeripheral.Id().String()),
		)

		machineOpts = append(machineOpts, WithPeripheral(machinePeripheral))
	}

	return NewMachine(config.Name, machineOpts...), nil
}

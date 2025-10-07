package go_kvm_agent

import (
	"fmt"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/peripheral"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/routing"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

func createDisplayRouterFromPeripheralRepository(repository *peripheral.PeripheralRepository, opts ...routing.DisplayRouterOption) (*routing.DisplayRouter, error) {
	sourcePeripherals := repository.GetAll(
		peripheral.FilterType(peripheralSDK.PeripheralTypeDisplay),
		peripheral.FilterRole(peripheralSDK.PeripheralRoleSource),
	)

	for _, sourcePeripheral := range sourcePeripherals {
		if displaySource, ok := sourcePeripheral.(peripheralSDK.DisplaySource); ok {
			opts = append(opts, routing.WithDisplaySource(displaySource))
		}
	}

	sinkPeripherals := repository.GetAll(
		peripheral.FilterType(peripheralSDK.PeripheralTypeDisplay),
		peripheral.FilterRole(peripheralSDK.PeripheralRoleSink),
	)

	for _, sinkPeripheral := range sinkPeripherals {
		if displaySink, ok := sinkPeripheral.(peripheralSDK.DisplaySink); ok {
			opts = append(opts, routing.WithDisplaySink(displaySink))
		}
	}

	return routing.NewDisplayRouter(opts...)
}

func createMachineFromConfigPath(configPath string) ([]*machine.Machine, error) {
	configs, err := loadMachineConfigFromPath(configPath)
	if err != nil {
		return nil, err
	}

	machines := make([]*machine.Machine, 0, len(configs))
	for i := range configs {
		machineConfig, err := machine.CreateMachineFromConfig(&configs[i])
		if err != nil {
			return nil, fmt.Errorf("create machine from config: %w", err)
		}
		machines = append(machines, machineConfig)
	}

	return machines, nil
}

func createPeripheralRepositoryFromMachines(machines []*machine.Machine) *peripheral.PeripheralRepository {
	opts := make([]peripheral.RepositoryOpt, 0, len(machines))
	for _, m := range machines {
		opts = append(opts, peripheral.WithPeripheralsFromSource(m))
	}
	return peripheral.NewPeripheralRepository(opts...)
}

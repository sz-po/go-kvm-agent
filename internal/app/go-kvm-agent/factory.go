package go_kvm_agent

import (
	"fmt"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/routing"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

func createDisplayRouter(sources []peripheral.DisplaySource, sinks []peripheral.DisplaySink, opts ...routing.DisplayRouterOption) (*routing.DisplayRouter, error) {
	for _, source := range sources {
		opts = append(opts, routing.WithDisplaySource(source))
	}

	for _, sink := range sinks {
		opts = append(opts, routing.WithDisplaySink(sink))
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

package go_kvm_agent

import (
	"context"
	"fmt"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/routing"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	routingSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/routing"
)

func createLocalDisplayRouterFromMachineRepository(ctx context.Context, repository machineSDK.Repository, opts ...routing.LocalDisplayRouterOpt) (routingSDK.DisplayRouter, error) {
	machines, err := repository.GetAllMachines(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all machines: %w", err)
	}

	for _, machine := range machines {
		displaySources, err := machine.Peripherals().GetAllDisplaySources(ctx)
		if err != nil {
			return nil, fmt.Errorf("get all display sources: %w", err)
		}

		for _, displaySourcePeripheral := range displaySources {
			opts = append(opts, routing.WithDisplaySource(displaySourcePeripheral))
		}

		displaySinks, err := machine.Peripherals().GetAllDisplaySinks(ctx)
		if err != nil {
			return nil, fmt.Errorf("get all display sinks: %w", err)
		}

		for _, displaySinkPeripheral := range displaySinks {
			opts = append(opts, routing.WithDisplaySink(displaySinkPeripheral))
		}
	}

	return routing.NewLocalDisplayRouter(opts...)
}

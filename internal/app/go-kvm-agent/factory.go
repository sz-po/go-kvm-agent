package go_kvm_agent

import (
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/peripheral"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/routing"
	routingSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/routing"
)

func createLocalDisplayRouterFromPeripheralRepository(repository *peripheral.PeripheralRepository, opts ...routing.LocalDisplayRouterOpt) (routingSDK.DisplayRouter, error) {
	for _, displaySourcePeripheral := range repository.GetAllDisplaySources() {
		opts = append(opts, routing.WithDisplaySource(displaySourcePeripheral))
	}

	for _, displaySinkPeripheral := range repository.GetAllDisplaySinks() {
		opts = append(opts, routing.WithDisplaySink(displaySinkPeripheral))
	}

	return routing.NewLocalDisplayRouter(opts...)
}

func createPeripheralRepositoryFromMachines(machines []*machine.Machine) (*peripheral.PeripheralRepository, error) {
	opts := make([]peripheral.RepositoryOpt, 0, len(machines))
	for _, m := range machines {
		opts = append(opts, peripheral.WithPeripheralsFromProvider(m))
	}
	return peripheral.NewPeripheralRepository(opts...)
}

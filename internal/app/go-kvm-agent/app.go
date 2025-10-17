package go_kvm_agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/handler/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/handler/router"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/http"
	internalMachine "github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/memory"
)

func Start(ctx context.Context, wg *sync.WaitGroup, config Config) error {
	memoryPool, err := memory.NewHeapPool(1024*1024*16, 32)
	if err != nil {
		return fmt.Errorf("create heap pool: %w", err)
	}

	err = memory.SetDefaultMemoryPool(memoryPool)
	if err != nil {
		return fmt.Errorf("set default memory pool: %w", err)
	}

	var machineRepositoryOpts []internalMachine.LocalRepositoryOpt

	for _, machineConfig := range config.Machines {
		machine, err := internalMachine.CreateMachineFromConfig(ctx, machineConfig)
		if err != nil {
			return fmt.Errorf("create machine: %w", err)
		}

		machineRepositoryOpts = append(machineRepositoryOpts, internalMachine.WithLocalRepositoryMachines(machine))
	}

	machinesRepository := internalMachine.NewLocalRepository(machineRepositoryOpts...)

	displayRouter, err := createLocalDisplayRouterFromMachineRepository(ctx, machinesRepository)
	if err != nil {
		return fmt.Errorf("create display routing: %w", err)
	}

	err = http.Listen(ctx, config.Api.ControlApi.Server,
		http.WithHandler(router.HandlerProvider(displayRouter, machinesRepository)),
		http.WithHandler(machine.HandlerProvider(machinesRepository)),
	)
	if err != nil {
		return fmt.Errorf("control api: http server: listen: %w", err)
	}

	wg.Add(1)
	go func() {
		<-ctx.Done()
		wg.Done()
	}()

	return nil
}

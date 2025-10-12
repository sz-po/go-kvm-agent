package go_kvm_agent

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/control/handler"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/http"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/machine"
)

func Start(ctx context.Context, wg *sync.WaitGroup, config Config) error {
	var machines []*machine.Machine

	for _, machineConfig := range config.Machines {
		newMachine, err := machine.CreateMachineFromConfig(ctx, machineConfig)
		if err != nil {
			return fmt.Errorf("create machine: %s: %w", machineConfig.Name, err)
		}

		peripherals := newMachine.GetPeripherals()

		slog.Info("Machine created.",
			slog.String("machineName", string(machineConfig.Name)),
			slog.Int("peripheralsCount", len(peripherals)),
		)

		machines = append(machines, newMachine)
	}

	peripheralRepository, err := createPeripheralRepositoryFromMachines(machines)
	if err != nil {
		return fmt.Errorf("create peripheral repository: %w", err)
	}

	displayRouter, err := createLocalDisplayRouterFromPeripheralRepository(peripheralRepository)
	if err != nil {
		return fmt.Errorf("create display routing: %w", err)
	}

	err = http.Listen(ctx, config.Api.ControlApi.Server,
		http.WithHandler(handler.NewDisplayRouterHandler(displayRouter)),
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

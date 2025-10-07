package go_kvm_agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/routing"
)

func Start(config Config, wg *sync.WaitGroup, ctx context.Context) error {
	machines, err := createMachineFromConfigPath(config.Machine.ConfigPath)
	if err != nil {
		return fmt.Errorf("create machines: %w", err)
	}

	peripheralRepository := createPeripheralRepositoryFromMachines(machines)

	router, err := createDisplayRouterFromPeripheralRepository(peripheralRepository)
	if err != nil {
		return fmt.Errorf("create display router: %w", err)
	}

	if err := router.Start(ctx); err != nil {
		return fmt.Errorf("start display router: %w", err)
	}

	if wg != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()

			<-ctx.Done()

			stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := router.Stop(stopCtx); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, routing.ErrDisplayRouterNotStarted) {
				slog.Error("Failed to stop display router.", slog.String("error", err.Error()))
			}
		}()
	}

	return nil
}

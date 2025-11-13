package orbiqd_peripheral

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/transport/p2p"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/cli"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/memory"
	nodeInternal "github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/node"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/peripheral"
	apiSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/api"
	nodeAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/service/node"
	peripheralAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/service/node/peripheral"
	driverSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/driver"
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

func Start(ctx context.Context, wg *sync.WaitGroup, config Config) error {
	logger := slog.Default()

	if err := setupMemoryPool(); err != nil {
		return fmt.Errorf("setup memory pool: %w", err)
	}

	driverRepository, err := createDriverRepository()
	if err != nil {
		return fmt.Errorf("create driver repository: %w", err)
	}

	driverList, err := driverRepository.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("get all drivers: %w", err)
	}

	for _, driver := range driverList {
		logger.Info("Found driver in repository.", slog.String("driverKind", driver.GetKind().String()))
	}

	nodeRepository := nodeInternal.NewNodeRepository()
	nodeRegistrar := nodeInternal.NewNodeRegistrar(nodeRepository)

	peripheralServices, err := setupPeripherals(ctx, wg, driverRepository, config.Peripheral)
	if err != nil {
		return fmt.Errorf("setup peripherals: %w", err)
	}

	transport, err := setupTransport(ctx, wg, config.Transport,
		p2p.WithTransportServices(peripheralServices...),
		p2p.WithTransportNodeRegistrar(nodeRegistrar),
	)
	if err != nil {
		return fmt.Errorf("setup transport: %w", err)
	}

	transport.GetLocalNodeId()

	return nil
}

func setupTransport(ctx context.Context, wg *sync.WaitGroup, config cli.TransportConfig, opts ...p2p.TransportOpt) (apiSDK.Transport, error) {
	logger := slog.Default()

	identity, err := p2p.NewIdentity(config.IdentityPath)
	if err != nil {
		return nil, fmt.Errorf("create identity: %w", err)
	}

	node := nodeInternal.NewNode(identity.GetId(), nodeInternal.WithNodeRole(nodeSDK.Peripheral))
	nodeService := nodeAPI.NewNodeAdapter(node)

	opts = append(opts,
		p2p.WithTransportServices(nodeService),
		p2p.WithTransportIdentity(identity),
		p2p.WithTransportLogger(logger),
		p2p.WithTransportBindAddress(config.BindAddress),
	)

	transport, err := p2p.NewTransport(opts...)
	if err != nil {
		return nil, fmt.Errorf("create transport: %w", err)
	}

	discoverer, err := p2p.NewDiscoverer(transport,
		p2p.WithMulticastDNSDiscovery(),
		p2p.WithDiscovererLogger(logger),
	)
	if err != nil {
		return nil, fmt.Errorf("create discoverer: %w", err)
	}

	wg.Add(1)
	go func() {
		<-ctx.Done()

		if err := discoverer.Terminate(ctx); err != nil {
			logger.Warn("Discoverer termination failed.", slog.String("error", err.Error()))
		}

		if err := transport.Terminate(ctx); err != nil {
			logger.Warn("Transport termination failed.", slog.String("error", err.Error()))
		}

		wg.Done()
		logger.Debug("Transport terminated.")
	}()

	logger.Debug("Waiting for discovery bootstrap finished.")
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(config.DiscoveryBootstrapWindow):
	}

	logger.Info("Transport ready.", slog.String("localNodeId", string(transport.GetLocalNodeId())))

	return transport, nil
}

func setupMemoryPool() error {
	memoryPool, err := memory.NewHeapPool(1024*1024*16, 32)
	if err != nil {
		return fmt.Errorf("create heap pool: %w", err)
	}

	err = memory.SetDefaultMemoryPool(memoryPool)
	if err != nil {
		return fmt.Errorf("set as default: %w", err)
	}

	return nil
}

func setupPeripherals(ctx context.Context, wg *sync.WaitGroup, driverRepository driverSDK.DriverRepository, peripheralConfigList []PeripheralConfig) ([]nodeSDK.Service, error) {
	var services []nodeSDK.Service
	var repositoryOpts []peripheral.RepositoryOpt

	for _, peripheralConfig := range peripheralConfigList {
		logger := slog.Default().With(
			slog.String("driverKind", peripheralConfig.DriverKind.String()),
			slog.String("peripheralName", peripheralConfig.Name.String()),
		)

		driver, err := driverRepository.GetByKind(ctx, peripheralConfig.DriverKind)
		if err != nil {
			return nil, fmt.Errorf("get driver by kind: %s: %w", peripheralConfig.DriverKind, err)
		}

		peripheralInstance, err := driver.CreatePeripheral(ctx, peripheralConfig.Config, peripheralConfig.Name)
		if err != nil {
			return nil, err
		}
		services = append(services, peripheralAPI.NewPeripheralAdapter(peripheralInstance,
			peripheralAPI.WithPeripheralAdapterLogger(logger),
		))

		if displaySource, isDisplaySource := peripheralInstance.(peripheralSDK.DisplaySource); isDisplaySource {
			services = append(services, peripheralAPI.NewDisplaySourceAdapter(displaySource,
				peripheralAPI.WithDisplaySourceAdapterLogger(logger),
			))
		}

		if displaySink, isDisplaySink := peripheralInstance.(peripheralSDK.DisplaySink); isDisplaySink {
			services = append(services, peripheralAPI.NewDisplaySinkAdapter(displaySink,
				peripheralAPI.WithDisplaySinkAdapterLogger(logger),
			))
		}

		repositoryOpts = append(repositoryOpts, peripheral.WithPeripheral(peripheralInstance))

		wg.Add(1)
		go func() {
			<-ctx.Done()

			if err := peripheralInstance.Terminate(ctx); err != nil {
				logger.Warn("Peripheral termination failed.", slog.String("error", err.Error()))
			}

			wg.Done()
			logger.Debug("Peripheral terminated.")
		}()

		logger.Info("Peripheral ready.")
	}

	logger := slog.Default()

	peripheralRepository, err := peripheral.NewRepository(repositoryOpts...)
	if err != nil {
		return nil, fmt.Errorf("create peripheral repository: %w", err)
	}
	services = append(services, peripheralAPI.NewRepositoryAdapter(peripheralRepository,
		peripheralAPI.WithRepositoryAdapterLogger(logger),
	))

	return services, nil
}

package orbiqd_ctl

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/transport/p2p"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/cli"
	nodeInternal "github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/node"
	apiSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/api"
	nodeAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/service/node"
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
)

func ProvideNodeRepository() *nodeInternal.NodeRepository {
	return nodeInternal.NewNodeRepository()
}

func ProvideTransport(ctx context.Context, wg *sync.WaitGroup, config cli.TransportConfig, nodeRepository *nodeInternal.NodeRepository) (apiSDK.Transport, error) {
	logger := slog.Default()

	nodeRegistrar := nodeInternal.NewNodeRegistrar(nodeRepository)

	identity, err := p2p.NewIdentity(config.IdentityPath)
	if err != nil {
		return nil, fmt.Errorf("create identity: %w", err)
	}

	node := nodeInternal.NewNode(identity.GetId(), nodeInternal.WithNodeRole(
		nodeSDK.CLI,
	))
	nodeService := nodeAPI.NewNodeAdapter(node)

	transport, err := p2p.NewTransport(
		p2p.WithTransportServices(nodeService),
		p2p.WithTransportIdentity(identity),
		p2p.WithTransportLogger(logger),
		p2p.WithTransportBindAddress(config.BindAddress),
		p2p.WithTransportNodeRegistrar(nodeRegistrar),
	)
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

		logger.Debug("Terminating transport.")

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

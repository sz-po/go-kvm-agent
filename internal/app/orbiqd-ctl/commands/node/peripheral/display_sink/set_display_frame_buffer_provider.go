package display_sink

import (
	"context"
	"fmt"
	"log/slog"

	apiSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/api"
	peripheralAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/service/node/peripheral"
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type SetDisplayFrameBufferProvider struct {
	NodeId                string `help:"Identifier of the node containing the display sink." required:"true" short:"n" long:"node-id"`
	PeripheralId          string `help:"Identifier of the display sink peripheral." required:"true" short:"p" long:"peripheral-id"`
	ProviderNodeId        string `help:"Identifier of the node containing the display source provider." required:"true" long:"provider-node-id"`
	ProviderPeripheralId  string `help:"Identifier of the display source provider peripheral." required:"true" long:"provider-peripheral-id"`
}

func (command *SetDisplayFrameBufferProvider) Run(ctx context.Context, transport apiSDK.Transport, logger *slog.Logger) error {
	nodeId := nodeSDK.NodeId(command.NodeId)
	peripheralId := peripheralSDK.Id(command.PeripheralId)
	providerNodeId := nodeSDK.NodeId(command.ProviderNodeId)
	providerPeripheralId := peripheralSDK.Id(command.ProviderPeripheralId)

	logger = logger.With(
		slog.String("nodeId", string(nodeId)),
		slog.String("peripheralId", string(peripheralId)),
		slog.String("providerNodeId", string(providerNodeId)),
		slog.String("providerPeripheralId", string(providerPeripheralId)),
	)

	// Get the display sink peripheral
	sinkRepositoryClient := peripheralAPI.NewRepositoryClient(nodeId, transport)

	sinkPeripheral, err := sinkRepositoryClient.GetPeripheralById(ctx, peripheralId)
	if err != nil {
		return fmt.Errorf("get display sink peripheral: %w", err)
	}

	sinkPeripheralClient, isSinkPeripheralClient := sinkPeripheral.(*peripheralAPI.PeripheralClient)
	if !isSinkPeripheralClient {
		return fmt.Errorf("peripheral %s is not a peripheral api client", peripheralId)
	}

	displaySink := peripheralAPI.AsDisplaySink(sinkPeripheralClient)

	// Get the display source provider peripheral
	providerRepositoryClient := peripheralAPI.NewRepositoryClient(providerNodeId, transport)

	providerPeripheral, err := providerRepositoryClient.GetPeripheralById(ctx, providerPeripheralId)
	if err != nil {
		return fmt.Errorf("get display source provider peripheral: %w", err)
	}

	providerPeripheralClient, isProviderPeripheralClient := providerPeripheral.(*peripheralAPI.PeripheralClient)
	if !isProviderPeripheralClient {
		return fmt.Errorf("peripheral %s is not a peripheral api client", providerPeripheralId)
	}

	displaySourceProvider := peripheralAPI.AsDisplaySource(providerPeripheralClient)

	// Set the frame buffer provider
	err = displaySink.SetDisplayFrameBufferProvider(displaySourceProvider)
	if err != nil {
		return fmt.Errorf("set display frame buffer provider: %w", err)
	}

	logger.Info("Display frame buffer provider set successfully.")

	return nil
}

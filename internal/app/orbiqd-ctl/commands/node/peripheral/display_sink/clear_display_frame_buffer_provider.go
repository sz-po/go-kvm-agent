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

type ClearDisplayFrameBufferProvider struct {
	NodeId       string `help:"Identifier of the node containing the display sink." required:"true" short:"n" long:"node-id"`
	PeripheralId string `help:"Identifier of the display sink peripheral." required:"true" short:"p" long:"peripheral-id"`
}

func (command *ClearDisplayFrameBufferProvider) Run(ctx context.Context, transport apiSDK.Transport, logger *slog.Logger) error {
	nodeId := nodeSDK.NodeId(command.NodeId)
	peripheralId := peripheralSDK.Id(command.PeripheralId)

	logger = logger.With(
		slog.String("nodeId", string(nodeId)),
		slog.String("peripheralId", string(peripheralId)),
	)

	// Get the display sink peripheral
	repositoryClient := peripheralAPI.NewRepositoryClient(nodeId, transport)

	peripheral, err := repositoryClient.GetPeripheralById(ctx, peripheralId)
	if err != nil {
		return fmt.Errorf("get display sink peripheral: %w", err)
	}

	peripheralClient, isPeripheralClient := peripheral.(*peripheralAPI.PeripheralClient)
	if !isPeripheralClient {
		return fmt.Errorf("peripheral %s is not a peripheral api client", peripheralId)
	}

	displaySink := peripheralAPI.AsDisplaySink(peripheralClient)

	// Clear the frame buffer provider
	err = displaySink.ClearDisplayFrameBufferProvider()
	if err != nil {
		return fmt.Errorf("clear display frame buffer provider: %w", err)
	}

	logger.Info("Display frame buffer provider cleared successfully.")

	return nil
}

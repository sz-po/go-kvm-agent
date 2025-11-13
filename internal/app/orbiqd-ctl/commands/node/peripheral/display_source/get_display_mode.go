package display_source

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/lensesio/tableprinter"
	apiSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/api"
	peripheralAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/service/node/peripheral"
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type GetDisplayMode struct {
	NodeId       string `help:"Identifier of the node to query." required:"true" short:"n" long:"node-id"`
	PeripheralId string `help:"Identifier of the display source." required:"true" short:"p" long:"peripheral-id"`
}

func (command *GetDisplayMode) Run(ctx context.Context, transport apiSDK.Transport, logger *slog.Logger) error {
	nodeId := nodeSDK.NodeId(command.NodeId)
	peripheralId := peripheralSDK.Id(command.PeripheralId)

	logger = logger.With(
		slog.String("nodeId", string(nodeId)),
		slog.String("peripheralId", string(peripheralId)),
	)

	repositoryClient := peripheralAPI.NewRepositoryClient(nodeId, transport)

	peripheral, err := repositoryClient.GetPeripheralById(ctx, peripheralId)
	if err != nil {
		return fmt.Errorf("get display sources: %w", err)
	}

	peripheralClient, isPeripheralClient := peripheral.(*peripheralAPI.PeripheralClient)
	if !isPeripheralClient {
		return fmt.Errorf("peripheral %s is not a peripheral api client", peripheralId)
	}

	displaySource := peripheralAPI.AsDisplaySource(peripheralClient)

	displayMode, err := displaySource.GetDisplayMode(ctx)
	if err != nil {
		return fmt.Errorf("get display mode: %w", err)
	}

	output := []displayModeOutput{
		{
			NodeId:       nodeId,
			PeripheralId: peripheralId,
		},
	}

	if displayMode != nil {
		output[0].Width = displayMode.Width
		output[0].Height = displayMode.Height
		output[0].RefreshRate = displayMode.RefreshRate
	}

	tableprinter.Print(os.Stdout, output)

	logger.Info("Display mode fetched.")

	return nil
}

type displayModeOutput struct {
	NodeId       nodeSDK.NodeId   `json:"nodeId" header:"Node ID"`
	PeripheralId peripheralSDK.Id `json:"peripheralId" header:"Peripheral ID"`
	Width        uint32           `json:"width" header:"Width"`
	Height       uint32           `json:"height" header:"Height"`
	RefreshRate  uint32           `json:"refreshRate" header:"Refresh Rate"`
}

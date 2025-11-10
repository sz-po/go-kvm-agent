package peripheral

import (
	"context"
	"log/slog"
	"os"

	"github.com/lensesio/tableprinter"
	apiSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/api"
	peripheralAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/service/node/peripheral"
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type List struct {
	NodeId string `help:"Identifier of the node to query for peripherals." required:"true" short:"n" long:"node-id"`
}

type listOutput struct {
	NodeId       nodeSDK.NodeId     `json:"nodeId" header:"Node ID"`
	PeripheralId peripheralSDK.Id   `json:"peripheralId" header:"Peripheral ID"`
	Name         peripheralSDK.Name `json:"name" header:"Name"`
	Capabilities []string           `json:"capabilities" header:"Capabilities"`
	Error        string             `json:"error" header:"Error"`
}

func (cmd *List) Run(ctx context.Context, transport apiSDK.Transport, logger *slog.Logger) error {
	nodeId := nodeSDK.NodeId(cmd.NodeId)
	logger = logger.With(slog.String("nodeId", string(nodeId)))

	peripheralRepository := peripheralAPI.NewRepositoryClient(nodeId, transport)

	logger.Debug("Fetching peripherals for node.")
	peripherals, err := peripheralRepository.GetAllPeripherals(ctx)

	output := make([]listOutput, 0)

	if err != nil {
		output = append(output, listOutput{
			NodeId: nodeId,
			Error:  err.Error(),
		})
		tableprinter.Print(os.Stdout, output)
		return nil
	}

	for _, peripheral := range peripherals {
		output = append(output, listOutput{
			NodeId:       nodeId,
			PeripheralId: peripheral.GetId(),
			Name:         peripheral.GetName(),
			Capabilities: formatCapabilities(peripheral.GetCapabilities()),
		})
	}

	tableprinter.Print(os.Stdout, output)
	logger.Info("Peripherals listed.", slog.Int("peripheralCount", len(output)))

	return nil
}

func formatCapabilities(capabilities []peripheralSDK.PeripheralCapability) []string {
	formatted := make([]string, 0, len(capabilities))
	for _, capability := range capabilities {
		formatted = append(formatted, capability.String())
	}
	return formatted
}

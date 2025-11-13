package display_source

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/memory"
	apiSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/api"
	peripheralAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/service/node/peripheral"
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type GetDisplayFrameBuffer struct {
	NodeId       string `help:"Identifier of the node to query." required:"true" short:"n" long:"node-id"`
	PeripheralId string `help:"Identifier of the display source." required:"true" short:"p" long:"peripheral-id"`
	OutputFile   string `help:"Output file." required:"true" short:"o" long:"output-file"`
}

func (command *GetDisplayFrameBuffer) Run(ctx context.Context, transport apiSDK.Transport, logger *slog.Logger) error {
	memoryPool, err := memory.NewHeapPool(1024*1024*16, 16)
	if err != nil {
		return fmt.Errorf("create memory pool: %w", err)
	}
	err = memory.SetDefaultMemoryPool(memoryPool)
	if err != nil {
		return fmt.Errorf("set memory pool as default: %w", err)
	}

	nodeId := nodeSDK.NodeId(command.NodeId)
	peripheralId := peripheralSDK.Id(command.PeripheralId)
	outputFile := command.OutputFile

	logger = logger.With(
		slog.String("nodeId", string(nodeId)),
		slog.String("peripheralId", peripheralId.String()),
		slog.String("outputFile", outputFile),
	)

	peripheralRepository := peripheralAPI.NewRepositoryClient(nodeId, transport)
	peripheral, err := peripheralRepository.GetPeripheralById(ctx, peripheralId)
	if err != nil {
		return fmt.Errorf("get peripheral: %w", err)
	}

	peripheralClient, isPeripheralClient := peripheral.(*peripheralAPI.PeripheralClient)
	if !isPeripheralClient {
		return fmt.Errorf("peripheral %s is not a peripheral api client", peripheralId)
	}

	displaySource := peripheralAPI.AsDisplaySource(peripheralClient)

	frameBuffer, err := displaySource.GetDisplayFrameBuffer(ctx)
	if err != nil {
		return fmt.Errorf("get display frame buffer: %w", err)
	}
	defer func() {
		if releaseErr := frameBuffer.Release(); releaseErr != nil {
			logger.Warn("Failed to release frame buffer.", slog.String("error", releaseErr.Error()))
		}
	}()

	var output *os.File
	switch outputFile {
	case "-":
		output = os.Stdout
	default:
		output, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("create output file %s: %w", outputFile, err)
		}
		defer func() {
			if closeErr := output.Close(); closeErr != nil {
				logger.Warn("Failed to close output file.", slog.String("error", closeErr.Error()))
			}
		}()
	}

	writtenBytes, err := frameBuffer.WriteTo(output)
	if err != nil {
		return fmt.Errorf("write frame buffer to output: %w", err)
	}

	logger.Info("Display frame buffer saved.", slog.Int64("bytesWritten", writtenBytes))

	return nil
}

package display_source

import (
	"context"
	"fmt"
	"log/slog"

	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
)

type GetDisplayPixelFormat struct {
	NodeId       string `help:"Identifier of the node to query." required:"true" short:"n" long:"node-id"`
	PeripheralId string `help:"Identifier of the display source." required:"true" short:"p" long:"peripheral-id"`
}

func (command *GetDisplayPixelFormat) Run(ctx context.Context, repository nodeSDK.NodeRepository, logger *slog.Logger) error {
	return fmt.Errorf("display source get display pixel format: not implemented")
}

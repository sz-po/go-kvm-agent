package node

import "github.com/szymonpodeszwa/go-kvm-agent/internal/app/orbiqd-ctl/commands/node/peripheral"

type Commands struct {
	List       List                `cmd:"true" help:"List all nodes with theirs details and capabilities."`
	Peripheral peripheral.Commands `cmd:"true" help:"Peripheral-related commands for a specific node."`
}

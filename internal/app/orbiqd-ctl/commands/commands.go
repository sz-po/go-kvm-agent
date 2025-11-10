package commands

import "github.com/szymonpodeszwa/go-kvm-agent/internal/app/orbiqd-ctl/commands/node"

type Commands struct {
	Node node.Commands `cmd:"true" help:"Node-related commands."`
}

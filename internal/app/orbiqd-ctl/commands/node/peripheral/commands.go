package peripheral

import (
	"github.com/szymonpodeszwa/go-kvm-agent/internal/app/orbiqd-ctl/commands/node/peripheral/display_sink"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/app/orbiqd-ctl/commands/node/peripheral/display_source"
)

type Commands struct {
	List          List                    `cmd:"true" help:"List peripherals registered on a specific node."`
	DisplaySource display_source.Commands `cmd:"true" help:"Display source related commands."`
	DisplaySink   display_sink.Commands   `cmd:"true" help:"Display sink related commands."`
}

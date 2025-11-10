package main

import (
	"github.com/alecthomas/kong"
	orbiqd_ctl "github.com/szymonpodeszwa/go-kvm-agent/internal/app/orbiqd-ctl"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/app/orbiqd-ctl/commands"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/cli"
)

func main() {
	cli.BootstrapCommands[orbiqd_ctl.Config, commands.Commands](
		cli.WithApplicationName("orbiqd-ctl"),
		cli.WithKongOptions(
			kong.BindSingletonProvider(orbiqd_ctl.ProvideNodeRepository),
			kong.BindSingletonProvider(orbiqd_ctl.ProvideTransport),
		),
	)
}

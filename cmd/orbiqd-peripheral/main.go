package main

import (
	orbiqd_peripheral "github.com/szymonpodeszwa/go-kvm-agent/internal/app/orbiqd-peripheral"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/cli"
)

func main() {
	cli.BootstrapDaemon(orbiqd_peripheral.Start,
		cli.WithApplicationName("orbiqd-peripheral"),
	)
}

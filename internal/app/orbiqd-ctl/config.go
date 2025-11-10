package orbiqd_ctl

import "github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/cli"

type Config struct {
	cli.LogConfigHelper
	cli.TransportConfigHelper
}

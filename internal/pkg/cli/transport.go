package cli

import "time"

type TransportConfig struct {
	IdentityPath             string        `help:"Path to the identity private key file." placeholder:"FILE" required:"true" type:"path"`
	BindAddress              string        `help:"IP address to bind on." default:"0.0.0.0"`
	DiscoveryBootstrapWindow time.Duration `help:"How long transport will wait for discovery finished after startup." default:"5s"`
}

type SupportTransportConfig interface {
	GetTransportConfig() TransportConfig
}

type TransportConfigHelper struct {
	Transport TransportConfig `kong:"embed,prefix='transport-',group='Transport configuration'"`
}

func (helper TransportConfigHelper) GetTransportConfig() TransportConfig {
	return helper.Transport
}

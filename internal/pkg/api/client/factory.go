package client

import (
	"fmt"

	api "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/client"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/client/transport"
)

func CreateClientFromConfig(config Config) (*api.Client, error) {
	var roundTripper transport.RoundTripper
	var err error

	switch config.Protocol {
	case "http", "https":
		roundTripper, err = transport.NewHTTPRoundTripper(config.Protocol, config.Address, config.Port)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", config.Protocol)
	}

	if err != nil {
		return nil, err
	}

	client := api.NewClient(roundTripper)

	return client, nil
}

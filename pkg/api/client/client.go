package client

import (
	"fmt"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/client/service/machine"
	clientTransport "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/client/transport"
	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
)

type Client struct {
	roundTripper clientTransport.RoundTripper
}

func NewClient(roundTripper clientTransport.RoundTripper) *Client {
	return &Client{
		roundTripper: roundTripper,
	}
}

func (client *Client) Machine(machineIdentifier machineAPI.MachineIdentifier) (*machine.Service, error) {
	machineService, err := machine.NewService(client.roundTripper, machineIdentifier)
	if err != nil {
		return nil, fmt.Errorf("machine service: %w", err)
	}

	return machineService, nil
}

func (client *Client) Machines() (*machine.Iterator, error) {
	iterator, err := machine.NewIterator(client.roundTripper)
	if err != nil {
		return nil, fmt.Errorf("machine iterator: %w", err)
	}

	return iterator, nil
}

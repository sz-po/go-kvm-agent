package peripheral

import (
	"context"
	"fmt"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/client/transport"
	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	peripheralAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
)

type Iterator struct {
	machineIdentifier machineAPI.MachineIdentifier

	roundTripper transport.RoundTripper
}

func NewIterator(roundTripper transport.RoundTripper, machineIdentifier machineAPI.MachineIdentifier) (*Iterator, error) {
	if err := machineIdentifier.Validate(); err != nil {
		return nil, fmt.Errorf("machine identifier: %w", err)
	}

	return &Iterator{
		machineIdentifier: machineIdentifier,

		roundTripper: roundTripper,
	}, nil
}

func (iterator *Iterator) List(ctx context.Context) ([]peripheralAPI.Peripheral, error) {
	request := peripheralAPI.ListRequest{
		Path: peripheralAPI.ListRequestPath{
			MachineIdentifier: iterator.machineIdentifier,
		},
	}

	response, err := transport.CallUsingRequestProvider[peripheralAPI.ListResponse](ctx, iterator.roundTripper, &request, peripheralAPI.ParseListResponse)
	if err != nil {
		return nil, fmt.Errorf("call: %w", err)
	}

	return response.Body.Peripherals, nil
}

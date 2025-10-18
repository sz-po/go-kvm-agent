package machine

import (
	"context"
	"fmt"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/client/transport"
	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
)

type Iterator struct {
	roundTripper transport.RoundTripper
}

func NewIterator(roundTripper transport.RoundTripper) (*Iterator, error) {
	return &Iterator{
		roundTripper: roundTripper,
	}, nil
}

func (iterator *Iterator) List(ctx context.Context) ([]machineAPI.Machine, error) {
	request := machineAPI.ListRequest{}

	response, err := transport.CallUsingRequestProvider[machineAPI.ListResponse](ctx, iterator.roundTripper, &request, machineAPI.ParseListResponse)
	if err != nil {
		return nil, fmt.Errorf("call: %w", err)
	}

	return response.Body.Machines, nil
}

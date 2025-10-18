package machine

import (
	"context"
	"fmt"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/client/service/machine/peripheral"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/client/transport"
	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	peripheralAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
)

type Service struct {
	machineIdentifier machineAPI.MachineIdentifier

	roundTripper transport.RoundTripper
}

func NewService(roundTripper transport.RoundTripper, machineIdentifier machineAPI.MachineIdentifier) (*Service, error) {
	if err := machineIdentifier.Validate(); err != nil {
		return nil, fmt.Errorf("machine identifier: %w", err)
	}

	return &Service{
		machineIdentifier: machineIdentifier,

		roundTripper: roundTripper,
	}, nil
}

func (service *Service) Get(ctx context.Context) (*machineAPI.Machine, error) {
	request := machineAPI.GetRequest{
		Path: machineAPI.GetRequestPath{
			MachineIdentifier: service.machineIdentifier,
		},
	}

	response, err := transport.CallUsingRequestProvider[machineAPI.GetResponse](ctx, service.roundTripper, &request, machineAPI.ParseGetResponse)
	if err != nil {
		return nil, fmt.Errorf("call: %w", err)
	}

	return &response.Body.Machine, nil
}

func (service *Service) Peripherals() (*peripheral.Iterator, error) {
	iterator, err := peripheral.NewIterator(service.roundTripper, service.machineIdentifier)
	if err != nil {
		return nil, fmt.Errorf("peripheral iterator: %w", err)
	}

	return iterator, nil
}

func (service *Service) Peripheral(identifier peripheralAPI.PeripheralIdentifier) (*peripheral.Service, error) {
	return peripheral.NewService(service.roundTripper, service.machineIdentifier, identifier)
}

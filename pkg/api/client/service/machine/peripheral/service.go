package peripheral

import (
	"context"
	"fmt"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/client/service/machine/peripheral/display_source"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/client/transport"
	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	peripheralAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
)

type Service struct {
	machineIdentifier    machineAPI.MachineIdentifier
	peripheralIdentifier peripheralAPI.PeripheralIdentifier

	roundTripper transport.RoundTripper
}

func NewService(roundTripper transport.RoundTripper, machineIdentifier machineAPI.MachineIdentifier, peripheralIdentifier peripheralAPI.PeripheralIdentifier) (*Service, error) {
	if err := machineIdentifier.Validate(); err != nil {
		return nil, fmt.Errorf("machine identifier: %w", err)
	}

	if err := peripheralIdentifier.Validate(); err != nil {
		return nil, fmt.Errorf("peripheral identifier: %w", err)
	}

	return &Service{
		machineIdentifier:    machineIdentifier,
		peripheralIdentifier: peripheralIdentifier,

		roundTripper: roundTripper,
	}, nil
}

func (service *Service) Get(ctx context.Context) (*peripheralAPI.Peripheral, error) {
	request := &peripheralAPI.GetRequest{
		Path: peripheralAPI.GetRequestPath{
			MachineIdentifier:    service.machineIdentifier,
			PeripheralIdentifier: service.peripheralIdentifier,
		},
	}

	response, err := transport.CallUsingRequestProvider[peripheralAPI.GetResponse](ctx, service.roundTripper, request, peripheralAPI.ParseGetResponse)
	if err != nil {
		return nil, fmt.Errorf("call: %w", err)
	}

	return &response.Body.Peripheral, nil
}

func (service *Service) DisplaySource() (*display_source.Service, error) {
	return display_source.NewService(service.roundTripper, service.machineIdentifier, service.peripheralIdentifier)
}

package remote

import (
	"context"
	"fmt"
	"slices"

	peripheralAPIService "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/client/service/machine/peripheral"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type Peripheral struct {
	peripheralService peripheralAPIService.Service

	id           peripheralSDK.PeripheralId
	name         peripheralSDK.PeripheralName
	capabilities []peripheralSDK.PeripheralCapability
}

func NewPeripheral(ctx context.Context, peripheralService peripheralAPIService.Service) (*Peripheral, error) {
	peripheral, err := peripheralService.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("get peripheral: %w", err)
	}

	return &Peripheral{
		peripheralService: peripheralService,

		id:           peripheral.Id,
		name:         peripheral.Name,
		capabilities: peripheral.Capabilities,
	}, nil
}

func (peripheral *Peripheral) GetCapabilities() []peripheralSDK.PeripheralCapability {
	return peripheral.capabilities
}

func (peripheral *Peripheral) GetId() peripheralSDK.PeripheralId {
	return peripheral.id
}

func (peripheral *Peripheral) GetName() peripheralSDK.PeripheralName {
	return peripheral.name
}

func (peripheral *Peripheral) Terminate(ctx context.Context) error {
	return nil
}

func (peripheral *Peripheral) IsDisplaySource() bool {
	return slices.ContainsFunc(peripheral.capabilities, func(capability peripheralSDK.PeripheralCapability) bool {
		return capability.Equals(peripheralSDK.DisplaySourceCapability)
	})
}

func (peripheral *Peripheral) AsDisplaySource() (peripheralSDK.DisplaySource, error) {
	if !peripheral.IsDisplaySource() {
		return nil, peripheralSDK.ErrNotDisplaySource
	}

	displaySourceService, err := peripheral.peripheralService.DisplaySource()
	if err != nil {
		return nil, fmt.Errorf("display source service: %w", err)
	}

	return &DisplaySource{
		Peripheral:           peripheral,
		displaySourceService: displaySourceService,
	}, nil
}

package helper

import (
	"context"
	"fmt"

	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	peripheralAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

func GetPeripheralByIdentifier(ctx context.Context, repository peripheralSDK.Repository, identifier peripheralAPI.PeripheralIdentifier) (peripheralSDK.Peripheral, error) {
	switch {
	case identifier.Id != nil:
		return repository.GetPeripheralById(ctx, *identifier.Id)
	case identifier.Name != nil:
		return repository.GetPeripheralByName(ctx, *identifier.Name)
	default:
		return nil, fmt.Errorf("invalid identifier")
	}
}

func GetMachinePeripheralByIdentifier(ctx context.Context, repository machineSDK.Repository, machineIdentifier machineAPI.MachineIdentifier, peripheralIdentifier peripheralAPI.PeripheralIdentifier) (peripheralSDK.Peripheral, error) {
	machine, err := GetMachineByIdentifier(ctx, repository, machineIdentifier)
	if err != nil {
		return nil, err
	}

	peripheral, err := GetPeripheralByIdentifier(ctx, machine.Peripherals(), peripheralIdentifier)
	if err != nil {
		return nil, err
	}

	return peripheral, nil
}

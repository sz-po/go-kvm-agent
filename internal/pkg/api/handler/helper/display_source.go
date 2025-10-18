package helper

import (
	"context"

	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	peripheralAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

func GetMachineDisplaySourceByIdentifier(ctx context.Context, repository machineSDK.Repository, machineIdentifier machineAPI.MachineIdentifier, peripheralIdentifier peripheralAPI.PeripheralIdentifier) (peripheralSDK.DisplaySource, error) {
	machine, err := GetMachineByIdentifier(ctx, repository, machineIdentifier)
	if err != nil {
		return nil, err
	}

	peripheral, err := GetPeripheralByIdentifier(ctx, machine.Peripherals(), peripheralIdentifier)
	if err != nil {
		return nil, err
	}

	displaySource, err := peripheralSDK.AsDisplaySource(peripheral)
	if err != nil {
		return nil, err
	}

	return displaySource, nil
}

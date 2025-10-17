package helper

import (
	"context"
	"fmt"

	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
)

func GetMachineByIdentifier(ctx context.Context, repository machineSDK.Repository, identifier machineAPI.MachineIdentifier) (machineSDK.Machine, error) {
	switch {
	case identifier.Id != nil:
		return repository.GetMachineById(ctx, *identifier.Id)
	case identifier.Name != nil:
		return repository.GetMachineByName(ctx, *identifier.Name)
	default:
		return nil, fmt.Errorf("invalid identifier")
	}
}

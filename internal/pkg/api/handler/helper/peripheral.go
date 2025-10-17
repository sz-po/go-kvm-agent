package helper

import (
	"context"
	"fmt"

	peripheralAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
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

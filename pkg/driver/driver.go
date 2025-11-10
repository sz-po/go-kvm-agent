package driver

import (
	"context"

	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type Kind string

func (kind Kind) String() string {
	return string(kind)
}

type Driver interface {
	GetKind() Kind
	CreatePeripheral(ctx context.Context, config any, name peripheralSDK.Name) (peripheralSDK.Peripheral, error)
}

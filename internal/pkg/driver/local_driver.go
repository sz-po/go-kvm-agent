package driver

import (
	"context"

	driverSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/driver"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type LocalDriverCreateFn func(ctx context.Context, config any, name peripheralSDK.Name) (peripheralSDK.Peripheral, error)

type LocalDriver struct {
	kind     driverSDK.Kind
	createFn LocalDriverCreateFn
}

func NewLocalDriver(kind driverSDK.Kind, createFn LocalDriverCreateFn) *LocalDriver {
	return &LocalDriver{
		kind:     kind,
		createFn: createFn,
	}
}

func (driver *LocalDriver) GetKind() driverSDK.Kind {
	return driver.kind
}

func (driver *LocalDriver) CreatePeripheral(ctx context.Context, config any, name peripheralSDK.Name) (peripheralSDK.Peripheral, error) {
	return driver.createFn(ctx, config, name)
}

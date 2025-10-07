package peripheral

import (
	"errors"
	"fmt"
	
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

// CreatePeripheralFromConfig creates a peripheral instance from the provided configuration.
// It delegates to type-specific factory functions based on the peripheral type.
func CreatePeripheralFromConfig(config PeripheralConfig) (peripheralSDK.Peripheral, error) {
	switch config.Type {
	case peripheralSDK.PeripheralTypeDisplay:
		peripheral, err := createDisplayPeripheralFromConfig(config)
		if err != nil {
			return nil, fmt.Errorf("error creating display peripheral: %w", err)
		}

		return peripheral, nil
	default:
		return nil, ErrUnsupportedPeripheralType
	}
}

// createDisplayPeripheralFromConfig creates a display peripheral from configuration.
// It delegates to role-specific factory functions based on the peripheral role.
func createDisplayPeripheralFromConfig(config PeripheralConfig) (peripheralSDK.Peripheral, error) {
	if config.Type != peripheralSDK.PeripheralTypeDisplay {
		panic("peripheralSDK must be a display peripheralSDK")
	}

	switch config.Role {
	case peripheralSDK.PeripheralRoleSource:
		peripheral, err := createDisplaySourcePeripheralFromConfig(config)
		if err != nil {
			return nil, fmt.Errorf("error creating display source peripheral: %w", err)
		}

		return peripheral, nil
	default:
		return nil, ErrUnsupportedPeripheralRole
	}
}

// createDisplaySourcePeripheralFromConfig creates a display source peripheral from configuration.
// It delegates to driver-specific factory functions based on the peripheral driver.
func createDisplaySourcePeripheralFromConfig(config PeripheralConfig) (peripheralSDK.Peripheral, error) {
	switch config.Driver {
	default:
		return nil, ErrUnsupportedPeripheralDriver
	}
}

var (
	// ErrUnsupportedPeripheralType indicates that the peripheral type is not supported by the factory.
	ErrUnsupportedPeripheralType = errors.New("unsupported peripheralSDK type")
	// ErrUnsupportedPeripheralRole indicates that the peripheral role is not supported by the factory.
	ErrUnsupportedPeripheralRole = errors.New("unsupported peripheralSDK role")
	// ErrUnsupportedPeripheralDriver indicates that the peripheral driver is not supported by the factory.
	ErrUnsupportedPeripheralDriver = errors.New("unsupported peripheralSDK driver")
)

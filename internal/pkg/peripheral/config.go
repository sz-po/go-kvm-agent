package peripheral

import peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"

// PeripheralConfig represents the configuration for creating a peripheral device.
// It specifies the peripheral ID, driver, and driver-specific configuration.
type PeripheralConfig struct {
	Name   peripheralSDK.PeripheralName   `json:"name"`
	Driver peripheralSDK.PeripheralDriver `json:"driver"`
	Config any                            `json:"config"`
}

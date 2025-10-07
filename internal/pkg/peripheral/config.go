package peripheral

import peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"

// PeripheralConfig represents the configuration for creating a peripheral device.
// It specifies the peripheral type, role, driver, and driver-specific configuration.
type PeripheralConfig struct {
	Type   peripheralSDK.PeripheralType   `json:"type"`
	Role   peripheralSDK.PeripheralRole   `json:"role"`
	Driver peripheralSDK.PeripheralDriver `json:"driver"`
	Config any                            `json:"config"`
}

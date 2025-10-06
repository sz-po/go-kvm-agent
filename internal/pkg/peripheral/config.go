package peripherals

import "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripherals"

type PeripheralConfig struct {
	Type   peripherals.PeripheralType `json:"type"`
	Role   peripherals.PeripheralRole `json:"role"`
	Config any                        `json:"config"`
}

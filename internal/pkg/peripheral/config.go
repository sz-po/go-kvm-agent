package peripheral

import "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"

type PeripheralConfig struct {
	Type   peripheral.PeripheralType `json:"type"`
	Role   peripheral.PeripheralRole `json:"role"`
	Config any                       `json:"config"`
}

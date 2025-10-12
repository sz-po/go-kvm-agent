package machine

import "github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/peripheral"

// MachineConfig represents the configuration for creating a virtual machine.
// It includes the machine name and attached peripheral configurations.
type MachineConfig struct {
	Name        MachineName                   `json:"name"`
	Peripherals []peripheral.PeripheralConfig `json:"peripherals"`
}

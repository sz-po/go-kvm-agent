package machine

import machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"

// MachineConfig represents the configuration for creating a virtual machine.
// It includes the machine name and attached peripheral configurations.
type MachineConfig struct {
	Name   machineSDK.MachineName `json:"name"`
	Local  *LocalMachineConfig    `json:"local,omitempty"`
	Remote *RemoteMachineConfig   `json:"remote,omitempty"`
}

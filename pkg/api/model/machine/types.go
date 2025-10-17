package machine

import (
	"fmt"
	"strings"

	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
)

// MachineIdentifier represents a machine reference that can be provided either
// by id or by name. Only one of the fields is expected to be non-nil for a
// given identifier instance.
type MachineIdentifier struct {
	Id   *machineSDK.MachineId   `json:"id"`
	Name *machineSDK.MachineName `json:"name"`
}

// Validate ensures the identifier contains exactly one of id or name. It
// returns an error if neither or both of the fields are set.
func (machineIdentifier *MachineIdentifier) Validate() error {
	if machineIdentifier == nil {
		return fmt.Errorf("machine identifier: identifier is nil")
	}

	hasId := machineIdentifier.Id != nil
	hasName := machineIdentifier.Name != nil

	if !hasId && !hasName {
		return fmt.Errorf("machine identifier: either id or name must be provided")
	}

	if hasId && hasName {
		return fmt.Errorf("machine identifier: id and name are mutually exclusive")
	}

	return nil
}

type Machine struct {
	Id   machineSDK.MachineId   `json:"id"`
	Name machineSDK.MachineName `json:"name"`
}

type MachineList []Machine

// ParseMachineIdentifier converts a path segment formatted as
// "id:<machine-id>" or "name:<machine-name>" into a MachineIdentifier. The
// prefix determines whether the returned identifier targets the id or the name,
// and it validates the referenced value using the domain constructors. The
// function returns an error when the prefix is missing, unknown, or the value
// fails validation.
func ParseMachineIdentifier(machineIdentifier string) (*MachineIdentifier, error) {
	machineIdentifierType, machineIdentifierValue, found := strings.Cut(machineIdentifier, ":")
	if !found {
		return nil, fmt.Errorf("missing identifier type")
	}

	switch machineIdentifierType {
	case "id":
		machineId, err := machineSDK.NewMachineId(machineIdentifierValue)
		if err != nil {
			return nil, fmt.Errorf("invalid machine id: %w", err)
		}

		return &MachineIdentifier{
			Id: &machineId,
		}, nil
	case "name":
		machineName, err := machineSDK.NewMachineName(machineIdentifierValue)
		if err != nil {
			return nil, fmt.Errorf("invalid machine name: %w", err)
		}

		return &MachineIdentifier{
			Name: &machineName,
		}, nil
	default:
		return nil, fmt.Errorf("unknown machine type: %s", machineIdentifierType)
	}
}

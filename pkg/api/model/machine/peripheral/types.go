package peripheral

import (
	"fmt"
	"strings"

	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

const (
	PeripheralIdentifierPathFieldName = "peripheralIdentifier"
)

// PeripheralIdentifier represents a peripheral reference that can be provided
// either by id or by name. Only one of the fields should be defined for a given
// identifier instance.
type PeripheralIdentifier struct {
	Id   *peripheralSDK.PeripheralId   `json:"id"`
	Name *peripheralSDK.PeripheralName `json:"name"`
}

// Validate ensures the identifier contains exactly one of id or name. It
// returns an error if neither or both of the fields are set.
func (peripheralIdentifier *PeripheralIdentifier) Validate() error {
	if peripheralIdentifier == nil {
		return fmt.Errorf("peripheral identifier: identifier is nil")
	}

	hasId := peripheralIdentifier.Id != nil
	hasName := peripheralIdentifier.Name != nil

	if !hasId && !hasName {
		return fmt.Errorf("peripheral identifier: either id or name must be provided")
	}

	if hasId && hasName {
		return fmt.Errorf("peripheral identifier: id and name are mutually exclusive")
	}

	return nil
}

func (peripheralIdentifier *PeripheralIdentifier) String() (*string, error) {
	if err := peripheralIdentifier.Validate(); err != nil {
		return nil, err
	}

	var result string

	switch {
	case peripheralIdentifier.Id != nil:
		result = fmt.Sprintf("id:%s", *peripheralIdentifier.Id)
	case peripheralIdentifier.Name != nil:
		result = fmt.Sprintf("name:%s", *peripheralIdentifier.Name)
	default:
		return nil, fmt.Errorf("peripheral identifier: either id or name must be provided")
	}

	return &result, nil
}

// ParsePeripheralIdentifier converts a path segment formatted as
// "id:<peripheral-id>" or "name:<peripheral-name>" into a PeripheralIdentifier.
// The prefix determines whether the resulting identifier targets the id or the
// name, and the value is validated using the domain constructors. The function
// returns an error if the prefix is missing, unknown, or the value fails
// validation.
func ParsePeripheralIdentifier(peripheralIdentifier string) (*PeripheralIdentifier, error) {
	peripheralIdentifierType, peripheralIdentifierValue, found := strings.Cut(peripheralIdentifier, ":")
	if !found {
		return nil, fmt.Errorf("missing identifier type")
	}

	switch peripheralIdentifierType {
	case "id":
		peripheralId, err := peripheralSDK.NewPeripheralId(peripheralIdentifierValue)
		if err != nil {
			return nil, fmt.Errorf("invalid peripheral id: %w", err)
		}

		return &PeripheralIdentifier{
			Id: &peripheralId,
		}, nil
	case "name":
		peripheralName, err := peripheralSDK.NewPeripheralName(peripheralIdentifierValue)
		if err != nil {
			return nil, fmt.Errorf("invalid peripheral name: %w", err)
		}

		return &PeripheralIdentifier{
			Name: &peripheralName,
		}, nil
	default:
		return nil, fmt.Errorf("unknown peripheral type: %s", peripheralIdentifierType)
	}
}

type Peripheral struct {
	Id           peripheralSDK.PeripheralId           `json:"id"`
	Name         peripheralSDK.PeripheralName         `json:"name"`
	Capabilities []peripheralSDK.PeripheralCapability `json:"capabilities"`
}

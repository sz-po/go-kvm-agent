package peripheral

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"github.com/iancoleman/strcase"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils"
)

// PeripheralDriver identifies the driver implementation used for a peripheral.
type PeripheralDriver string

// PeripheralKind defines the category of peripheral device.
type PeripheralKind string

const (
	// PeripheralKindUnknown represents an uninitialized or invalid peripheral kind.
	PeripheralKindUnknown PeripheralKind = ""
	// PeripheralKindDisplay represents a display peripheral.
	PeripheralKindDisplay PeripheralKind = "display"
	// PeripheralKindKeyboard represents a keyboard peripheral.
	PeripheralKindKeyboard PeripheralKind = "keyboard"
	// PeripheralKindMouse represents a mouse peripheral.
	PeripheralKindMouse PeripheralKind = "mouse"
)

// String returns the string representation of the peripheral kind.
func (pk PeripheralKind) String() string {
	return string(pk)
}

// PeripheralRole defines whether a peripheral is a source or sink.
type PeripheralRole string

const (
	// PeripheralRoleUnknown represents an uninitialized or invalid peripheral role.
	PeripheralRoleUnknown PeripheralRole = ""
	// PeripheralRoleSource represents a peripheral that emits events or data.
	PeripheralRoleSource PeripheralRole = "source"
	// PeripheralRoleSink represents a peripheral that consumes events or data.
	PeripheralRoleSink PeripheralRole = "sink"
)

// String returns the string representation of the peripheral role.
func (pr PeripheralRole) String() string {
	return string(pr)
}

type PeripheralValidationFn func(peripheral Peripheral) error

// PeripheralCapability describes what a peripheral can do.
type PeripheralCapability struct {
	Kind         PeripheralKind
	Role         PeripheralRole
	validationFn PeripheralValidationFn
}

// NewCapability constructs a new PeripheralCapability with the given kind and role.
func NewCapability[T Peripheral](kind PeripheralKind, role PeripheralRole) PeripheralCapability {
	return PeripheralCapability{
		Kind: kind,
		Role: role,
		validationFn: func(peripheral Peripheral) error {
			if _, ok := peripheral.(T); !ok {
				return errors.New(fmt.Sprintf("peripheral %s does not implement %v",
					peripheral.GetId(),
					reflect.TypeOf((*T)(nil)).Elem(),
				))
			}
			return nil
		},
	}
}

// String returns the formatted capability in the form: {kind}-{role}
func (pc PeripheralCapability) String() string {
	return fmt.Sprintf("%s-%s", pc.Kind, pc.Role)
}

func (pc PeripheralCapability) Validate(peripheral Peripheral) error {
	return pc.validationFn(peripheral)
}

// PeripheralName uniquely identifies a peripheral device.
type PeripheralName string

// NewPeripheralName constructs a new PeripheralName with the given identifier.
// Returns an error if the id is empty or not in kebab-case format.
func NewPeripheralName(name string) (PeripheralName, error) {
	if name == "" {
		return "", errors.New("peripheral name cannot be empty")
	}
	if !utils.IsKebabCase(name) {
		return "", errors.New("peripheral name must be kebab-case")
	}
	return PeripheralName(name), nil
}

// String returns the string representation of the peripheral Name.
func (name PeripheralName) String() string {
	return string(name)
}

// UnmarshalJSON implements json.Unmarshaler interface with validation.
func (name *PeripheralName) UnmarshalJSON(data []byte) error {
	var nameStr string
	if err := json.Unmarshal(data, &nameStr); err != nil {
		return fmt.Errorf("unmarshal peripheral name: %w", err)
	}

	validated, err := NewPeripheralName(nameStr)
	if err != nil {
		return fmt.Errorf("unmarshal peripheral name: %w", err)
	}

	*name = validated
	return nil
}

// MarshalJSON implements json.Marshaler interface.
func (name PeripheralName) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(name))
}

type PeripheralId string

func CreatePeripheralRandomId(prefix string) PeripheralId {
	return PeripheralId(fmt.Sprintf("%s-%s", strcase.ToKebab(prefix), uuid.NewString()))
}

func NewPeripheralId(id string) (PeripheralId, error) {
	if id == "" {
		return "", errors.New("peripheral id cannot be empty")
	}

	if !utils.IsKebabCase(id) {
		return "", errors.New("peripheral id should be kebab-case")
	}

	return PeripheralId(id), nil
}

func (id PeripheralId) String() string {
	return string(id)
}

// UnmarshalJSON implements json.Unmarshaler interface with validation.
func (id *PeripheralId) UnmarshalJSON(data []byte) error {
	var idStr string
	if err := json.Unmarshal(data, &idStr); err != nil {
		return fmt.Errorf("unmarshal peripheral id: %w", err)
	}

	validated, err := NewPeripheralId(idStr)
	if err != nil {
		return fmt.Errorf("unmarshal peripheral id: %w", err)
	}

	*id = validated
	return nil
}

// MarshalJSON implements json.Marshaler interface.
func (id PeripheralId) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(id))
}

func ValidatePeripheralCapabilities(peripheral Peripheral) error {
	for _, capability := range peripheral.GetCapabilities() {
		if err := capability.Validate(peripheral); err != nil {
			return fmt.Errorf("peripheral %s capability %s validation: %w", peripheral.GetId(), capability, err)
		}
	}

	return nil
}

// Peripheral is the base interface for all peripheral devices.
type Peripheral interface {
	GetCapabilities() []PeripheralCapability

	GetId() PeripheralId

	GetName() PeripheralName

	// Terminate peripheral. It must be called only once in peripheral life.
	Terminate(ctx context.Context) error
}

type PeripheralProvider interface {
	GetAllPeripherals(ctx context.Context) ([]Peripheral, error)
}

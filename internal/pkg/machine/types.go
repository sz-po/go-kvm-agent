package machine

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

// MachineName represents a validated machine name in kebab-case format.
// Valid names contain only lowercase letters, digits, and hyphens,
// following the pattern: lowercase-with-hyphens
type MachineName string

var (
	// ErrInvalidMachineName indicates that the provided machine name doesn't meet validation requirements.
	ErrInvalidMachineName = errors.New("invalid machine name")

	// machineNameRegex validates kebab-case format: lowercase letters, digits, and hyphens.
	// Pattern: starts and ends with alphanumeric, hyphens only as separators (no consecutive hyphens).
	machineNameRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
)

// NewMachineName creates and validates a new MachineName.
// The name must follow kebab-case convention: only lowercase letters (a-z),
// digits (0-9), and hyphens (-) as separators. It cannot be empty, start or end
// with a hyphen, or contain consecutive hyphens.
//
// Valid examples: "my-machine", "mpv-vm-1", "server"
// Invalid examples: "My-Machine", "my machine", "-start", "end-", "double--dash", ""
func NewMachineName(name string) (MachineName, error) {
	if name == "" {
		return "", fmt.Errorf("%w: name cannot be empty", ErrInvalidMachineName)
	}

	if !machineNameRegex.MatchString(name) {
		return "", fmt.Errorf("%w: name must be in kebab-case format (lowercase letters, digits, and hyphens only)", ErrInvalidMachineName)
	}

	return MachineName(name), nil
}

// String returns the string representation of the MachineName.
func (machineName MachineName) String() string {
	return string(machineName)
}

// UnmarshalJSON implements json.Unmarshaler interface with validation.
func (machineName *MachineName) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return fmt.Errorf("unmarshal machine name: %w", err)
	}

	validated, err := NewMachineName(name)
	if err != nil {
		return fmt.Errorf("unmarshal machine name: %w", err)
	}

	*machineName = validated
	return nil
}

// MarshalJSON implements json.Marshaler interface.
func (machineName MachineName) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(machineName))
}

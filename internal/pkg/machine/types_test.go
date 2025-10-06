package machine

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMachineName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		errorType error
	}{
		// Valid cases
		{
			name:      "valid simple name",
			input:     "myvm",
			wantError: false,
		},
		{
			name:      "valid kebab-case name",
			input:     "my-machine",
			wantError: false,
		},
		{
			name:      "valid name with numbers",
			input:     "test-vm-1",
			wantError: false,
		},
		{
			name:      "valid name starting with number",
			input:     "1-test-vm",
			wantError: false,
		},
		{
			name:      "valid complex kebab-case",
			input:     "my-test-vm-123",
			wantError: false,
		},
		// Invalid cases
		{
			name:      "empty string",
			input:     "",
			wantError: true,
			errorType: ErrInvalidMachineName,
		},
		{
			name:      "uppercase letters",
			input:     "My-Machine",
			wantError: true,
			errorType: ErrInvalidMachineName,
		},
		{
			name:      "all uppercase",
			input:     "MYMACHINE",
			wantError: true,
			errorType: ErrInvalidMachineName,
		},
		{
			name:      "contains space",
			input:     "my machine",
			wantError: true,
			errorType: ErrInvalidMachineName,
		},
		{
			name:      "starts with hyphen",
			input:     "-myvm",
			wantError: true,
			errorType: ErrInvalidMachineName,
		},
		{
			name:      "ends with hyphen",
			input:     "myvm-",
			wantError: true,
			errorType: ErrInvalidMachineName,
		},
		{
			name:      "consecutive hyphens",
			input:     "my--machine",
			wantError: true,
			errorType: ErrInvalidMachineName,
		},
		{
			name:      "contains underscore",
			input:     "my_machine",
			wantError: true,
			errorType: ErrInvalidMachineName,
		},
		{
			name:      "contains dot",
			input:     "my.machine",
			wantError: true,
			errorType: ErrInvalidMachineName,
		},
		{
			name:      "contains special characters",
			input:     "my@machine",
			wantError: true,
			errorType: ErrInvalidMachineName,
		},
		{
			name:      "only hyphen",
			input:     "-",
			wantError: true,
			errorType: ErrInvalidMachineName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewMachineName(tt.input)

			if tt.wantError {
				assert.Error(t, err, "NewMachineName should fail for invalid input.")
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType, "NewMachineName should wrap expected error.")
				}
				return
			}

			if !assert.NoError(t, err, "NewMachineName should succeed for valid input.") {
				return
			}
			assert.Equal(t, tt.input, result.String(), "NewMachineName should preserve original value.")
		})
	}
}

func TestMachineNameString(t *testing.T) {
	name, err := NewMachineName("test-machine")
	if !assert.NoError(t, err, "NewMachineName should succeed.") {
		return
	}

	expected := "test-machine"
	assert.Equal(t, expected, name.String(), "String should return the canonical value.")
}

func TestMachineNameJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{
			name:      "valid JSON",
			input:     `"my-machine"`,
			wantError: false,
		},
		{
			name:      "invalid name in JSON",
			input:     `"My-Machine"`,
			wantError: true,
		},
		{
			name:      "empty name in JSON",
			input:     `""`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var machineName MachineName
			err := json.Unmarshal([]byte(tt.input), &machineName)

			if tt.wantError {
				assert.Error(t, err, "UnmarshalJSON should fail for invalid machine name.")
				return
			}

			assert.NoError(t, err, "UnmarshalJSON should succeed for valid machine name.")
		})
	}
}

func TestMachineNameJSONRoundTrip(t *testing.T) {
	original := "test-vm-123"
	machineName, err := NewMachineName(original)
	if !assert.NoError(t, err, "NewMachineName should succeed.") {
		return
	}

	// Marshal
	data, err := json.Marshal(machineName)
	if !assert.NoError(t, err, "Marshal should succeed.") {
		return
	}

	// Unmarshal
	var decoded MachineName
	if !assert.NoError(t, json.Unmarshal(data, &decoded), "Unmarshal should succeed.") {
		return
	}

	assert.Equal(t, original, decoded.String(), "Round trip should preserve machine name value.")
}

func TestMachineNameInStruct(t *testing.T) {
	type TestStruct struct {
		Name MachineName `json:"name"`
	}

	validJSON := `{"name":"my-test-vm"}`
	var valid TestStruct
	if !assert.NoError(t, json.Unmarshal([]byte(validJSON), &valid), "Unmarshal should succeed for valid JSON.") {
		return
	}
	assert.Equal(t, "my-test-vm", valid.Name.String(), "Struct should contain decoded machine name.")

	invalidJSON := `{"name":"Invalid-Name"}`
	var invalid TestStruct
	assert.Error(t, json.Unmarshal([]byte(invalidJSON), &invalid), "Unmarshal should fail for invalid machine name.")
}

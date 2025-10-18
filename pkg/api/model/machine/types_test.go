package machine_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	apiMachine "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
)

func TestParseMachineIdentifierById(t *testing.T) {
	t.Parallel()

	parsedIdentifier, err := apiMachine.ParseMachineIdentifier("id:test-machine")

	assert.NoError(t, err)
	if assert.NotNil(t, parsedIdentifier) {
		if assert.NotNil(t, parsedIdentifier.Id) {
			assert.Equal(t, machineSDK.MachineId("test-machine"), *parsedIdentifier.Id)
		}

		assert.Nil(t, parsedIdentifier.Name)
	}
}

func TestParseMachineIdentifierByName(t *testing.T) {
	t.Parallel()

	parsedIdentifier, err := apiMachine.ParseMachineIdentifier("name:test-machine")

	assert.NoError(t, err)
	if assert.NotNil(t, parsedIdentifier) {
		assert.Nil(t, parsedIdentifier.Id)

		if assert.NotNil(t, parsedIdentifier.Name) {
			assert.Equal(t, machineSDK.MachineName("test-machine"), *parsedIdentifier.Name)
		}
	}
}

func TestParseMachineIdentifierMissingColon(t *testing.T) {
	t.Parallel()

	parsedIdentifier, err := apiMachine.ParseMachineIdentifier("invalid")

	assert.Error(t, err)
	assert.Nil(t, parsedIdentifier)
	assert.EqualError(t, err, "missing identifier type")
}

func TestParseMachineIdentifierUnknownPrefix(t *testing.T) {
	t.Parallel()

	parsedIdentifier, err := apiMachine.ParseMachineIdentifier("uuid:test-machine")

	assert.Error(t, err)
	assert.Nil(t, parsedIdentifier)
	assert.EqualError(t, err, "unknown machine type: uuid")
}

func TestParseMachineIdentifierInvalidId(t *testing.T) {
	t.Parallel()

	parsedIdentifier, err := apiMachine.ParseMachineIdentifier("id:")

	assert.Error(t, err)
	assert.Nil(t, parsedIdentifier)
	assert.EqualError(t, err, "invalid machine id: machine id cannot be empty")
}

func TestParseMachineIdentifierInvalidName(t *testing.T) {
	t.Parallel()

	parsedIdentifier, err := apiMachine.ParseMachineIdentifier("name:Invalid-Name")

	assert.Error(t, err)
	assert.Nil(t, parsedIdentifier)
	assert.Contains(t, err.Error(), "invalid machine name")
}

func TestMachineIdentifierUnmarshalJSONById(t *testing.T) {
	t.Parallel()

	var machineIdentifier apiMachine.MachineIdentifier
	err := json.Unmarshal([]byte(`{"id":"test-machine"}`), &machineIdentifier)

	assert.NoError(t, err)
	assert.NoError(t, machineIdentifier.Validate())
	if assert.NotNil(t, machineIdentifier.Id) {
		assert.Equal(t, machineSDK.MachineId("test-machine"), *machineIdentifier.Id)
	}
	assert.Nil(t, machineIdentifier.Name)
}

func TestMachineIdentifierUnmarshalJSONByName(t *testing.T) {
	t.Parallel()

	var machineIdentifier apiMachine.MachineIdentifier
	err := json.Unmarshal([]byte(`{"name":"test-machine"}`), &machineIdentifier)

	assert.NoError(t, err)
	assert.NoError(t, machineIdentifier.Validate())
	assert.Nil(t, machineIdentifier.Id)
	if assert.NotNil(t, machineIdentifier.Name) {
		assert.Equal(t, machineSDK.MachineName("test-machine"), *machineIdentifier.Name)
	}
}

func TestMachineIdentifierUnmarshalJSONMissingValues(t *testing.T) {
	t.Parallel()

	var machineIdentifier apiMachine.MachineIdentifier
	assert.NoError(t, json.Unmarshal([]byte(`{}`), &machineIdentifier))

	assert.EqualError(t, machineIdentifier.Validate(), "either id or name must be provided")
}

func TestMachineIdentifierUnmarshalJSONAmbiguousValues(t *testing.T) {
	t.Parallel()

	var machineIdentifier apiMachine.MachineIdentifier
	assert.NoError(t, json.Unmarshal([]byte(`{"id":"test-machine","name":"test-machine"}`), &machineIdentifier))

	assert.EqualError(t, machineIdentifier.Validate(), "id and name are mutually exclusive")
}

func TestMachineIdentifierUnmarshalJSONInvalidId(t *testing.T) {
	t.Parallel()

	var machineIdentifier apiMachine.MachineIdentifier
	err := json.Unmarshal([]byte(`{"id":""}`), &machineIdentifier)

	assert.EqualError(t, err, "unmarshal machine id: machine id cannot be empty")
}

func TestMachineIdentifierUnmarshalJSONInvalidName(t *testing.T) {
	t.Parallel()

	var machineIdentifier apiMachine.MachineIdentifier
	err := json.Unmarshal([]byte(`{"name":"Invalid-Name"}`), &machineIdentifier)

	assert.Contains(t, err.Error(), "unmarshal machine name")
}

func TestMachineIdentifierStringWithId(t *testing.T) {
	t.Parallel()

	machineId := machineSDK.MachineId("test-machine")
	identifier := apiMachine.MachineIdentifier{Id: &machineId}

	result, err := identifier.String()

	assert.NoError(t, err)
	if assert.NotNil(t, result) {
		assert.Equal(t, "id:test-machine", *result)
	}
}

func TestMachineIdentifierStringWithName(t *testing.T) {
	t.Parallel()

	machineName := machineSDK.MachineName("test-machine")
	identifier := apiMachine.MachineIdentifier{Name: &machineName}

	result, err := identifier.String()

	assert.NoError(t, err)
	if assert.NotNil(t, result) {
		assert.Equal(t, "name:test-machine", *result)
	}
}

func TestMachineIdentifierStringWithNilIdentifier(t *testing.T) {
	t.Parallel()

	var identifier *apiMachine.MachineIdentifier

	result, err := identifier.String()

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "identifier is nil")
}

func TestMachineIdentifierStringWithEmptyIdentifier(t *testing.T) {
	t.Parallel()

	identifier := apiMachine.MachineIdentifier{}

	result, err := identifier.String()

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "either id or name must be provided")
}

func TestMachineIdentifierStringWithBothFields(t *testing.T) {
	t.Parallel()

	machineId := machineSDK.MachineId("test-machine")
	machineName := machineSDK.MachineName("test-machine")
	identifier := apiMachine.MachineIdentifier{Id: &machineId, Name: &machineName}

	result, err := identifier.String()

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "id and name are mutually exclusive")
}

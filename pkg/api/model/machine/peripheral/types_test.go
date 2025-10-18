package peripheral_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	apiPeripheral "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

func TestParsePeripheralIdentifierById(t *testing.T) {
	t.Parallel()

	parsedIdentifier, err := apiPeripheral.ParsePeripheralIdentifier("id:test-peripheral")

	assert.NoError(t, err)
	if assert.NotNil(t, parsedIdentifier) {
		if assert.NotNil(t, parsedIdentifier.Id) {
			assert.Equal(t, peripheralSDK.PeripheralId("test-peripheral"), *parsedIdentifier.Id)
		}

		assert.Nil(t, parsedIdentifier.Name)
	}
}

func TestParsePeripheralIdentifierByName(t *testing.T) {
	t.Parallel()

	parsedIdentifier, err := apiPeripheral.ParsePeripheralIdentifier("name:test-peripheral")

	assert.NoError(t, err)
	if assert.NotNil(t, parsedIdentifier) {
		assert.Nil(t, parsedIdentifier.Id)

		if assert.NotNil(t, parsedIdentifier.Name) {
			assert.Equal(t, peripheralSDK.PeripheralName("test-peripheral"), *parsedIdentifier.Name)
		}
	}
}

func TestParsePeripheralIdentifierMissingColon(t *testing.T) {
	t.Parallel()

	parsedIdentifier, err := apiPeripheral.ParsePeripheralIdentifier("invalid")

	assert.Error(t, err)
	assert.Nil(t, parsedIdentifier)
	assert.EqualError(t, err, "missing identifier type")
}

func TestParsePeripheralIdentifierUnknownPrefix(t *testing.T) {
	t.Parallel()

	parsedIdentifier, err := apiPeripheral.ParsePeripheralIdentifier("uuid:test-peripheral")

	assert.Error(t, err)
	assert.Nil(t, parsedIdentifier)
	assert.EqualError(t, err, "unknown peripheral type: uuid")
}

func TestParsePeripheralIdentifierInvalidId(t *testing.T) {
	t.Parallel()

	parsedIdentifier, err := apiPeripheral.ParsePeripheralIdentifier("id:")

	assert.Error(t, err)
	assert.Nil(t, parsedIdentifier)
	assert.EqualError(t, err, "invalid peripheral id: peripheral id cannot be empty")
}

func TestParsePeripheralIdentifierInvalidName(t *testing.T) {
	t.Parallel()

	parsedIdentifier, err := apiPeripheral.ParsePeripheralIdentifier("name:Invalid-Peripheral")

	assert.Error(t, err)
	assert.Nil(t, parsedIdentifier)
	assert.Contains(t, err.Error(), "invalid peripheral name")
}

func TestPeripheralIdentifierUnmarshalJSONById(t *testing.T) {
	t.Parallel()

	var peripheralIdentifier apiPeripheral.PeripheralIdentifier
	err := json.Unmarshal([]byte(`{"id":"test-peripheral"}`), &peripheralIdentifier)

	assert.NoError(t, err)
	assert.NoError(t, peripheralIdentifier.Validate())
	if assert.NotNil(t, peripheralIdentifier.Id) {
		assert.Equal(t, peripheralSDK.PeripheralId("test-peripheral"), *peripheralIdentifier.Id)
	}
	assert.Nil(t, peripheralIdentifier.Name)
}

func TestPeripheralIdentifierUnmarshalJSONByName(t *testing.T) {
	t.Parallel()

	var peripheralIdentifier apiPeripheral.PeripheralIdentifier
	err := json.Unmarshal([]byte(`{"name":"test-peripheral"}`), &peripheralIdentifier)

	assert.NoError(t, err)
	assert.NoError(t, peripheralIdentifier.Validate())
	assert.Nil(t, peripheralIdentifier.Id)
	if assert.NotNil(t, peripheralIdentifier.Name) {
		assert.Equal(t, peripheralSDK.PeripheralName("test-peripheral"), *peripheralIdentifier.Name)
	}
}

func TestPeripheralIdentifierUnmarshalJSONMissingValues(t *testing.T) {
	t.Parallel()

	var peripheralIdentifier apiPeripheral.PeripheralIdentifier
	assert.NoError(t, json.Unmarshal([]byte(`{}`), &peripheralIdentifier))

	assert.EqualError(t, peripheralIdentifier.Validate(), "peripheral identifier: either id or name must be provided")
}

func TestPeripheralIdentifierUnmarshalJSONAmbiguousValues(t *testing.T) {
	t.Parallel()

	var peripheralIdentifier apiPeripheral.PeripheralIdentifier
	assert.NoError(t, json.Unmarshal([]byte(`{"id":"test-peripheral","name":"test-peripheral"}`), &peripheralIdentifier))

	assert.EqualError(t, peripheralIdentifier.Validate(), "peripheral identifier: id and name are mutually exclusive")
}

func TestPeripheralIdentifierUnmarshalJSONInvalidId(t *testing.T) {
	t.Parallel()

	var peripheralIdentifier apiPeripheral.PeripheralIdentifier
	err := json.Unmarshal([]byte(`{"id":""}`), &peripheralIdentifier)

	assert.EqualError(t, err, "unmarshal peripheral id: peripheral id cannot be empty")
}

func TestPeripheralIdentifierUnmarshalJSONInvalidName(t *testing.T) {
	t.Parallel()

	var peripheralIdentifier apiPeripheral.PeripheralIdentifier
	err := json.Unmarshal([]byte(`{"name":"Invalid-Peripheral"}`), &peripheralIdentifier)

	assert.Contains(t, err.Error(), "unmarshal peripheral name")
}

func TestPeripheralIdentifierStringWithId(t *testing.T) {
	t.Parallel()

	peripheralId := peripheralSDK.PeripheralId("test-peripheral")
	identifier := apiPeripheral.PeripheralIdentifier{Id: &peripheralId}

	result, err := identifier.String()

	assert.NoError(t, err)
	if assert.NotNil(t, result) {
		assert.Equal(t, "id:test-peripheral", *result)
	}
}

func TestPeripheralIdentifierStringWithName(t *testing.T) {
	t.Parallel()

	peripheralName := peripheralSDK.PeripheralName("test-peripheral")
	identifier := apiPeripheral.PeripheralIdentifier{Name: &peripheralName}

	result, err := identifier.String()

	assert.NoError(t, err)
	if assert.NotNil(t, result) {
		assert.Equal(t, "name:test-peripheral", *result)
	}
}

func TestPeripheralIdentifierStringWithNilIdentifier(t *testing.T) {
	t.Parallel()

	var identifier *apiPeripheral.PeripheralIdentifier

	result, err := identifier.String()

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "peripheral identifier: identifier is nil")
}

func TestPeripheralIdentifierStringWithEmptyIdentifier(t *testing.T) {
	t.Parallel()

	identifier := apiPeripheral.PeripheralIdentifier{}

	result, err := identifier.String()

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "peripheral identifier: either id or name must be provided")
}

func TestPeripheralIdentifierStringWithBothFields(t *testing.T) {
	t.Parallel()

	peripheralId := peripheralSDK.PeripheralId("test-peripheral")
	peripheralName := peripheralSDK.PeripheralName("test-peripheral")
	identifier := apiPeripheral.PeripheralIdentifier{Id: &peripheralId, Name: &peripheralName}

	result, err := identifier.String()

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "peripheral identifier: id and name are mutually exclusive")
}

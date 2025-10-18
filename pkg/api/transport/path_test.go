package transport

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPath_Require(t *testing.T) {
	t.Run("returns no error when all required keys are present", func(t *testing.T) {
		path := PathParams{
			"machineId":    "machine-123",
			"peripheralId": "peripheral-456",
		}

		err := path.Require("machineId", "peripheralId")

		assert.NoError(t, err)
	})

	t.Run("returns error when single required key is missing", func(t *testing.T) {
		path := PathParams{
			"machineId": "machine-123",
		}

		err := path.Require("machineId", "peripheralId")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrMissingPathParamKey))
		assert.Contains(t, err.Error(), "peripheralId")
	})

	t.Run("returns error when multiple required keys are missing", func(t *testing.T) {
		path := PathParams{
			"otherId": "other-123",
		}

		err := path.Require("machineId", "peripheralId")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrMissingPathParamKey))
	})

	t.Run("returns no error when no keys are required", func(t *testing.T) {
		path := PathParams{
			"someId": "some-123",
		}

		err := path.Require()

		assert.NoError(t, err)
	})

	t.Run("returns no error when empty path and no keys required", func(t *testing.T) {
		path := PathParams{}

		err := path.Require()

		assert.NoError(t, err)
	})

	t.Run("returns error when empty path and keys are required", func(t *testing.T) {
		path := PathParams{}

		err := path.Require("machineId")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrMissingPathParamKey))
	})
}

func TestPathParams_Clone(t *testing.T) {
	t.Run("returns independent copy with all values", func(t *testing.T) {
		original := PathParams{
			"machineId":    "machine-123",
			"peripheralId": "peripheral-456",
		}

		cloned := original.Clone()

		assert.Equal(t, original, cloned)
		assert.Equal(t, "machine-123", cloned["machineId"])
		assert.Equal(t, "peripheral-456", cloned["peripheralId"])
	})

	t.Run("returns independent copy that can be modified without affecting original", func(t *testing.T) {
		original := PathParams{
			"machineId": "machine-123",
		}

		cloned := original.Clone()
		cloned["machineId"] = "machine-999"
		cloned["newKey"] = "new-value"

		assert.Equal(t, "machine-123", original["machineId"])
		assert.Equal(t, "machine-999", cloned["machineId"])
		assert.Equal(t, "", original["newKey"])
		assert.Equal(t, "new-value", cloned["newKey"])
	})

	t.Run("returns empty map when original is empty", func(t *testing.T) {
		original := PathParams{}

		cloned := original.Clone()

		assert.NotNil(t, cloned)
		assert.Equal(t, 0, len(cloned))
	})

	t.Run("returns nil when original is nil", func(t *testing.T) {
		var original PathParams

		cloned := original.Clone()

		assert.Nil(t, cloned)
	})
}

package transport

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPath_Require(t *testing.T) {
	t.Run("returns no error when all required keys are present", func(t *testing.T) {
		path := Path{
			"machineId":    "machine-123",
			"peripheralId": "peripheral-456",
		}

		err := path.Require("machineId", "peripheralId")

		assert.NoError(t, err)
	})

	t.Run("returns error when single required key is missing", func(t *testing.T) {
		path := Path{
			"machineId": "machine-123",
		}

		err := path.Require("machineId", "peripheralId")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrMissingPathKey))
		assert.Contains(t, err.Error(), "peripheralId")
	})

	t.Run("returns error when multiple required keys are missing", func(t *testing.T) {
		path := Path{
			"otherId": "other-123",
		}

		err := path.Require("machineId", "peripheralId")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrMissingPathKey))
	})

	t.Run("returns no error when no keys are required", func(t *testing.T) {
		path := Path{
			"someId": "some-123",
		}

		err := path.Require()

		assert.NoError(t, err)
	})

	t.Run("returns no error when empty path and no keys required", func(t *testing.T) {
		path := Path{}

		err := path.Require()

		assert.NoError(t, err)
	})

	t.Run("returns error when empty path and keys are required", func(t *testing.T) {
		path := Path{}

		err := path.Require("machineId")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrMissingPathKey))
	})
}

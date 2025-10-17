package transport

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeader_Require(t *testing.T) {
	t.Run("returns no error when all required keys are present", func(t *testing.T) {
		header := Header{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token123",
		}

		err := header.Require("Content-Type", "Authorization")

		assert.NoError(t, err)
	})

	t.Run("returns error when single required key is missing", func(t *testing.T) {
		header := Header{
			"Content-Type": "application/json",
		}

		err := header.Require("Content-Type", "Authorization")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrMissingHeaderKey))
		assert.Contains(t, err.Error(), "Authorization")
	})

	t.Run("returns error when multiple required keys are missing", func(t *testing.T) {
		header := Header{
			"Other-Header": "value",
		}

		err := header.Require("Content-Type", "Authorization")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrMissingHeaderKey))
	})

	t.Run("returns no error when no keys are required", func(t *testing.T) {
		header := Header{
			"Content-Type": "application/json",
		}

		err := header.Require()

		assert.NoError(t, err)
	})

	t.Run("returns no error when empty header and no keys required", func(t *testing.T) {
		header := Header{}

		err := header.Require()

		assert.NoError(t, err)
	})

	t.Run("returns error when empty header and keys are required", func(t *testing.T) {
		header := Header{}

		err := header.Require("Content-Type")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrMissingHeaderKey))
	})
}

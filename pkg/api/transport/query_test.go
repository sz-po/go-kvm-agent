package transport

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuery_Require(t *testing.T) {
	t.Run("returns no error when all required keys are present", func(t *testing.T) {
		query := Query{
			"filter": "active",
			"limit":  "10",
		}

		err := query.Require("filter", "limit")

		assert.NoError(t, err)
	})

	t.Run("returns error when single required key is missing", func(t *testing.T) {
		query := Query{
			"filter": "active",
		}

		err := query.Require("filter", "limit")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrMissingQueryKey))
		assert.Contains(t, err.Error(), "limit")
	})

	t.Run("returns error when multiple required keys are missing", func(t *testing.T) {
		query := Query{
			"other": "value",
		}

		err := query.Require("filter", "limit")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrMissingQueryKey))
	})

	t.Run("returns no error when no keys are required", func(t *testing.T) {
		query := Query{
			"filter": "active",
		}

		err := query.Require()

		assert.NoError(t, err)
	})

	t.Run("returns no error when empty query and no keys required", func(t *testing.T) {
		query := Query{}

		err := query.Require()

		assert.NoError(t, err)
	})

	t.Run("returns error when empty query and keys are required", func(t *testing.T) {
		query := Query{}

		err := query.Require("filter")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrMissingQueryKey))
	})
}

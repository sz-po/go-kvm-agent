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

func TestQuery_Clone(t *testing.T) {
	t.Run("returns independent copy with all values", func(t *testing.T) {
		original := Query{
			"filter": "active",
			"limit":  "10",
		}

		cloned := original.Clone()

		assert.Equal(t, original, cloned)
		assert.Equal(t, "active", cloned["filter"])
		assert.Equal(t, "10", cloned["limit"])
	})

	t.Run("returns independent copy that can be modified without affecting original", func(t *testing.T) {
		original := Query{
			"filter": "active",
		}

		cloned := original.Clone()
		cloned["filter"] = "inactive"
		cloned["newKey"] = "new-value"

		assert.Equal(t, "active", original["filter"])
		assert.Equal(t, "inactive", cloned["filter"])
		assert.Equal(t, "", original["newKey"])
		assert.Equal(t, "new-value", cloned["newKey"])
	})

	t.Run("returns empty map when original is empty", func(t *testing.T) {
		original := Query{}

		cloned := original.Clone()

		assert.NotNil(t, cloned)
		assert.Equal(t, 0, len(cloned))
	})

	t.Run("returns nil when original is nil", func(t *testing.T) {
		var original Query

		cloned := original.Clone()

		assert.Nil(t, cloned)
	})
}

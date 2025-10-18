package transport

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeader_Require(t *testing.T) {
	t.Run("returns no error when all required keys are present", func(t *testing.T) {
		header := Header{
			HeaderContentType: "application/json",
			"authorization":   "Bearer token123",
		}

		err := header.Require(HeaderContentType, "authorization")

		assert.NoError(t, err)
	})

	t.Run("returns error when single required key is missing", func(t *testing.T) {
		header := Header{
			HeaderContentType: "application/json",
		}

		err := header.Require(HeaderContentType, "authorization")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrMissingHeaderKey))
		assert.Contains(t, err.Error(), "authorization")
	})

	t.Run("returns error when multiple required keys are missing", func(t *testing.T) {
		header := Header{
			"other-header": "value",
		}

		err := header.Require(HeaderContentType, "authorization")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrMissingHeaderKey))
	})

	t.Run("returns no error when no keys are required", func(t *testing.T) {
		header := Header{
			HeaderContentType: "application/json",
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

		err := header.Require(HeaderContentType)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrMissingHeaderKey))
	})

	t.Run("is case-insensitive when checking required keys", func(t *testing.T) {
		header := Header{
			HeaderContentType: "application/json",
			"authorization":   "Bearer token123",
		}

		err := header.Require("Content-Type", "Authorization")

		assert.NoError(t, err)
	})

	t.Run("is case-insensitive when reporting missing keys", func(t *testing.T) {
		header := Header{
			HeaderContentType: "application/json",
		}

		err := header.Require("Content-Type", "Authorization")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrMissingHeaderKey))
		assert.Contains(t, err.Error(), "Authorization")
	})
}

func TestHeader_Get(t *testing.T) {
	t.Run("returns value when key exists", func(t *testing.T) {
		header := Header{
			HeaderContentType: "application/json",
		}

		value := header.Get(HeaderContentType)

		assert.Equal(t, "application/json", value)
	})

	t.Run("returns empty string when key does not exist", func(t *testing.T) {
		header := Header{
			HeaderContentType: "application/json",
		}

		value := header.Get("authorization")

		assert.Equal(t, "", value)
	})

	t.Run("returns empty string when header is empty", func(t *testing.T) {
		header := Header{}

		value := header.Get(HeaderContentType)

		assert.Equal(t, "", value)
	})

	t.Run("is case-insensitive", func(t *testing.T) {
		header := Header{
			HeaderContentType: "application/json",
		}

		value := header.Get("Content-Type")

		assert.Equal(t, "application/json", value)
	})

	t.Run("is case-insensitive with mixed case", func(t *testing.T) {
		header := Header{
			HeaderContentType: "application/json",
		}

		value := header.Get("CONTENT-TYPE")

		assert.Equal(t, "application/json", value)
	})
}

func TestHeader_Clone(t *testing.T) {
	t.Run("returns independent copy with all values", func(t *testing.T) {
		original := Header{
			HeaderContentType: "application/json",
			"authorization":   "Bearer token123",
		}

		cloned := original.Clone()

		assert.Equal(t, original, cloned)
		assert.Equal(t, "application/json", cloned[HeaderContentType])
		assert.Equal(t, "Bearer token123", cloned["authorization"])
	})

	t.Run("returns independent copy that can be modified without affecting original", func(t *testing.T) {
		original := Header{
			HeaderContentType: "application/json",
		}

		cloned := original.Clone()
		cloned[HeaderContentType] = "text/html"
		cloned["newKey"] = "new-value"

		assert.Equal(t, "application/json", original[HeaderContentType])
		assert.Equal(t, "text/html", cloned[HeaderContentType])
		assert.Equal(t, "", original["newKey"])
		assert.Equal(t, "new-value", cloned["newKey"])
	})

	t.Run("returns empty map when original is empty", func(t *testing.T) {
		original := Header{}

		cloned := original.Clone()

		assert.NotNil(t, cloned)
		assert.Equal(t, 0, len(cloned))
	})

	t.Run("returns nil when original is nil", func(t *testing.T) {
		var original Header

		cloned := original.Clone()

		assert.Nil(t, cloned)
	})
}

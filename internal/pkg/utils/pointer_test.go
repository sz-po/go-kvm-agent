package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultNil(t *testing.T) {
	t.Run("returns default value when pointer is nil", func(t *testing.T) {
		var nilInt *int
		result := DefaultNil(nilInt, 42)
		assert.Equal(t, 42, result)

		var nilString *string
		result2 := DefaultNil(nilString, "default")
		assert.Equal(t, "default", result2)

		var nilBool *bool
		result3 := DefaultNil(nilBool, true)
		assert.Equal(t, true, result3)
	})

	t.Run("returns pointer value when pointer is not nil", func(t *testing.T) {
		value := 100
		result := DefaultNil(&value, 42)
		assert.Equal(t, 100, result)

		str := "actual"
		result2 := DefaultNil(&str, "default")
		assert.Equal(t, "actual", result2)

		boolean := false
		result3 := DefaultNil(&boolean, true)
		assert.Equal(t, false, result3)
	})

	t.Run("works with complex types", func(t *testing.T) {
		type testStruct struct {
			Name  string
			Value int
		}

		defaultStruct := testStruct{Name: "default", Value: 0}

		var nilStruct *testStruct
		result := DefaultNil(nilStruct, defaultStruct)
		assert.Equal(t, defaultStruct, result)

		actualStruct := testStruct{Name: "actual", Value: 42}
		result2 := DefaultNil(&actualStruct, defaultStruct)
		assert.Equal(t, actualStruct, result2)
	})

	t.Run("works with zero values", func(t *testing.T) {
		zeroInt := 0
		result := DefaultNil(&zeroInt, 42)
		assert.Equal(t, 0, result)

		emptyString := ""
		result2 := DefaultNil(&emptyString, "default")
		assert.Equal(t, "", result2)
	})
}

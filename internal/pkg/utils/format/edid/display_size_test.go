package edid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func intPtr(value int) *int {
	return &value
}

func TestCreateDisplaySizeSpecificationFromBlock(t *testing.T) {
	t.Parallel()

	block := DisplaySizeBlock{10, 20}

	specification, err := CreateDisplaySizeSpecificationFromBlock(block)

	assert.NoError(t, err)
	if assert.NotNil(t, specification) {
		if assert.NotNil(t, specification.Width) {
			assert.Equal(t, 10, *specification.Width)
		}

		if assert.NotNil(t, specification.Height) {
			assert.Equal(t, 20, *specification.Height)
		}
	}
}

func TestCreateDisplaySizeSpecificationFromBlockZeroValues(t *testing.T) {
	t.Parallel()

	block := DisplaySizeBlock{0, 0}

	specification, err := CreateDisplaySizeSpecificationFromBlock(block)

	assert.NoError(t, err)
	if assert.NotNil(t, specification) {
		assert.Nil(t, specification.Width)
		assert.Nil(t, specification.Height)
	}
}

func TestCreateDisplaySizeSpecificationFromBlockInvalidDimensions(t *testing.T) {
	t.Parallel()

	block := DisplaySizeBlock{0, 10}

	specification, err := CreateDisplaySizeSpecificationFromBlock(block)

	assert.Nil(t, specification)
	assert.ErrorContains(t, err, "Width")
}

func TestCreateDisplaySizeBlockFromSpecification(t *testing.T) {
	t.Parallel()

	specification := DisplaySizeSpecification{
		Width:  intPtr(30),
		Height: intPtr(40),
	}

	block, err := CreateDisplaySizeBlockFromSpecification(specification)

	assert.NoError(t, err)
	if assert.NotNil(t, block) {
		assert.Equal(t, DisplaySizeBlock{30, 40}, *block)
	}
}

func TestCreateDisplaySizeBlockFromSpecificationEmpty(t *testing.T) {
	t.Parallel()

	specification := DisplaySizeSpecification{}

	block, err := CreateDisplaySizeBlockFromSpecification(specification)

	assert.NoError(t, err)
	if assert.NotNil(t, block) {
		assert.Equal(t, DisplaySizeBlock{0, 0}, *block)
	}
}

func TestCreateDisplaySizeBlockFromSpecificationInvalid(t *testing.T) {
	t.Parallel()

	specification := DisplaySizeSpecification{
		Width: intPtr(30),
	}

	block, err := CreateDisplaySizeBlockFromSpecification(specification)

	assert.Nil(t, block)
	assert.ErrorContains(t, err, "Height")
}

func TestDisplaySizeSpecificationValidateMaxBound(t *testing.T) {
	t.Parallel()

	specification := &DisplaySizeSpecification{
		Width:  intPtr(256),
		Height: intPtr(40),
	}

	err := specification.Validate()

	assert.ErrorContains(t, err, "max")
}

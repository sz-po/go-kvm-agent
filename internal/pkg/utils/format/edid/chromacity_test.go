package edid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateChromacityBlockFromSpecification(t *testing.T) {
	t.Parallel()

	specification := ChromacitySpecification{
		RedX:   0.640,
		RedY:   0.330,
		GreenX: 0.300,
		GreenY: 0.600,
		BlueX:  0.150,
		BlueY:  0.060,
		WhiteX: 0.313,
		WhiteY: 0.329,
	}

	block, err := CreateChromacityBlockFromSpecification(specification)

	assert.NoError(t, err)
	if assert.NotNil(t, block) {
		expected := ChromacityBlock{0xEE, 0x95, 0xA3, 0x54, 0x4C, 0x99, 0x26, 0x0F, 0x50, 0x54}
		assert.Equal(t, expected, *block)
	}
}

func TestCreateChromacityBlockFromSpecificationInvalid(t *testing.T) {
	t.Parallel()

	specification := ChromacitySpecification{
		RedX:   -0.1,
		RedY:   0.5,
		GreenX: 0.5,
		GreenY: 0.5,
		BlueX:  0.5,
		BlueY:  0.5,
		WhiteX: 0.5,
		WhiteY: 0.5,
	}

	block, err := CreateChromacityBlockFromSpecification(specification)

	assert.Nil(t, block)
	assert.ErrorContains(t, err, "RedX")
}

func TestCreateChromacitySpecificationFromBlock(t *testing.T) {
	t.Parallel()

	block := ChromacityBlock{0xEE, 0x95, 0xA3, 0x54, 0x4C, 0x99, 0x26, 0x0F, 0x50, 0x54}

	specification, err := CreateChromacitySpecificationFromBlock(block)

	assert.NoError(t, err)
	if assert.NotNil(t, specification) {
		assert.InDelta(t, 0.6396, specification.RedX, 0.0006)
		assert.InDelta(t, 0.3300, specification.RedY, 0.0006)
		assert.InDelta(t, 0.2998, specification.GreenX, 0.0006)
		assert.InDelta(t, 0.5996, specification.GreenY, 0.0006)
		assert.InDelta(t, 0.1504, specification.BlueX, 0.0006)
		assert.InDelta(t, 0.0596, specification.BlueY, 0.0006)
		assert.InDelta(t, 0.3135, specification.WhiteX, 0.0006)
		assert.InDelta(t, 0.3291, specification.WhiteY, 0.0006)
	}
}

func TestCreateChromacityBlockFromSpecificationClampsUpperBound(t *testing.T) {
	t.Parallel()

	specification := ChromacitySpecification{
		RedX:   0.640,
		RedY:   0.330,
		GreenX: 0.300,
		GreenY: 0.600,
		BlueX:  0.150,
		BlueY:  0.060,
		WhiteX: 1,
		WhiteY: 1,
	}

	block, err := CreateChromacityBlockFromSpecification(specification)

	assert.NoError(t, err)
	if assert.NotNil(t, block) {
		expected := ChromacityBlock{0xEE, 0x9F, 0xA3, 0x54, 0x4C, 0x99, 0x26, 0x0F, 0xFF, 0xFF}
		assert.Equal(t, expected, *block)
	}
}

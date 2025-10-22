package edid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateStandardTimingsBlockFromSpecificationEmpty(t *testing.T) {
	block, err := CreateStandardTimingsBlockFromSpecification(StandardTimingsSpecification{})

	assert.NoError(t, err)
	if assert.NotNil(t, block) {
		expected := StandardTimingsBlock{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}
		assert.Equal(t, expected, *block)
	}
}

func TestCreateStandardTimingsBlockFromSpecification(t *testing.T) {
	specification := StandardTimingsSpecification{
		Entries: []StandardTimingEntrySpecification{
			{
				Width:       1280,
				Height:      1024,
				RefreshRate: 75,
				AspectRatio: StandardTimingEntryAspectRatio5x4,
			},
			{
				Width:       1024,
				Height:      768,
				RefreshRate: 70,
				AspectRatio: StandardTimingEntryAspectRatio4x3,
			},
		},
	}

	block, err := CreateStandardTimingsBlockFromSpecification(specification)

	assert.NoError(t, err)
	if assert.NotNil(t, block) {
		expected := StandardTimingsBlock{0x81, 0x8F, 0x61, 0x4A, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}
		assert.Equal(t, expected, *block)

		roundTripSpecification, roundTripErr := CreateStandardTimingsSpecificationFromBlock(*block)

		assert.NoError(t, roundTripErr)
		if assert.NotNil(t, roundTripSpecification) {
			assert.Equal(t, specification.Entries, roundTripSpecification.Entries)
		}
	}
}

func TestCreateStandardTimingsSpecificationFromBlockEmpty(t *testing.T) {
	testCases := []StandardTimingsBlock{
		{},
		func() StandardTimingsBlock {
			var block StandardTimingsBlock
			for index := range block {
				block[index] = 0x01
			}
			return block
		}(),
	}

	for _, block := range testCases {
		specification, err := CreateStandardTimingsSpecificationFromBlock(block)

		assert.NoError(t, err)
		if assert.NotNil(t, specification) {
			assert.Empty(t, specification.Entries)
		}
	}
}

func TestCreateStandardTimingsSpecificationFromBlockUnsupportedAspectRatioBits(t *testing.T) {
	original := standardTimingAspectRatioByBits
	standardTimingAspectRatioByBits = map[byte]StandardTimingEntryAspectRatio{
		0b00: StandardTimingEntryAspectRatio16x10,
		0b01: StandardTimingEntryAspectRatio4x3,
		0b10: StandardTimingEntryAspectRatio5x4,
	}
	t.Cleanup(func() {
		standardTimingAspectRatioByBits = original
	})

	block := StandardTimingsBlock{}
	block[0] = 0x20
	block[1] = 0b11000000

	specification, err := CreateStandardTimingsSpecificationFromBlock(block)

	assert.Nil(t, specification)
	assert.EqualError(t, err, "unsupported aspect ratio bits: 0b11")
}

func TestCreateStandardTimingsSpecificationFromBlockDeriveHeightError(t *testing.T) {
	original := standardTimingAspectRatioByBits
	standardTimingAspectRatioByBits = map[byte]StandardTimingEntryAspectRatio{
		0b00: StandardTimingEntryAspectRatio("unsupported"),
	}
	t.Cleanup(func() {
		standardTimingAspectRatioByBits = original
	})

	block := StandardTimingsBlock{}
	block[0] = 0x20
	block[1] = 0x00

	specification, err := CreateStandardTimingsSpecificationFromBlock(block)

	assert.Nil(t, specification)
	assert.EqualError(t, err, "derive height: unsupported aspect ratio: unsupported")
}

func TestCreateStandardTimingsSpecificationFromBlockEntryValidationErrorWidthOutOfRange(t *testing.T) {
	block := StandardTimingsBlock{}
	block[0] = 0x00
	block[1] = 0b01000000

	specification, err := CreateStandardTimingsSpecificationFromBlock(block)

	assert.Nil(t, specification)
	assert.EqualError(t, err, "validate entry: width out of range: 248")
}

func TestCreateStandardTimingsBlockFromSpecificationInvalidWidth(t *testing.T) {
	specification := StandardTimingsSpecification{
		Entries: []StandardTimingEntrySpecification{
			{
				Width:       1025,
				Height:      768,
				RefreshRate: 75,
				AspectRatio: StandardTimingEntryAspectRatio4x3,
			},
		},
	}

	block, err := CreateStandardTimingsBlockFromSpecification(specification)

	assert.Nil(t, block)
	assert.EqualError(t, err, "derive width byte: width must be divisible by 8")
}

func TestCreateStandardTimingsBlockFromSpecificationValidateSpecificationError(t *testing.T) {
	specification := StandardTimingsSpecification{
		Entries: []StandardTimingEntrySpecification{
			{
				Width:       1280,
				Height:      800,
				RefreshRate: 75,
				AspectRatio: StandardTimingEntryAspectRatio16x9,
			},
		},
	}

	block, err := CreateStandardTimingsBlockFromSpecification(specification)

	assert.Nil(t, block)
	assert.EqualError(t, err, "validate specification: entry 0 invalid: height 800 does not match aspect ratio 16:9")
}

func TestStandardTimingsSpecificationValidateNil(t *testing.T) {
	var specification *StandardTimingsSpecification

	err := specification.Validate()

	assert.EqualError(t, err, "nil specification")
}

func TestStandardTimingsSpecificationValidateTooManyEntries(t *testing.T) {
	specification := &StandardTimingsSpecification{}

	for index := 0; index < 9; index++ {
		specification.Entries = append(specification.Entries, StandardTimingEntrySpecification{
			Width:       256 + index*8,
			Height:      (256 + index*8) * 3 / 4,
			RefreshRate: 60,
			AspectRatio: StandardTimingEntryAspectRatio4x3,
		})
	}

	err := specification.Validate()

	assert.EqualError(t, err, "too many timing entries: 9")
}

func TestStandardTimingEntrySpecificationValidateAspectMismatch(t *testing.T) {
	specification := &StandardTimingEntrySpecification{
		Width:       1280,
		Height:      720,
		RefreshRate: 75,
		AspectRatio: StandardTimingEntryAspectRatio5x4,
	}

	err := specification.Validate()

	assert.EqualError(t, err, "height 720 does not match aspect ratio 5:4")
}

func TestStandardTimingEntrySpecificationValidateWidthOutOfRange(t *testing.T) {
	specification := &StandardTimingEntrySpecification{
		Width:       240,
		Height:      180,
		RefreshRate: 75,
		AspectRatio: StandardTimingEntryAspectRatio4x3,
	}

	err := specification.Validate()

	assert.EqualError(t, err, "width out of range: 240")
}

func TestStandardTimingEntrySpecificationValidateRefreshRateOutOfRange(t *testing.T) {
	specification := &StandardTimingEntrySpecification{
		Width:       256,
		Height:      192,
		RefreshRate: 50,
		AspectRatio: StandardTimingEntryAspectRatio4x3,
	}

	err := specification.Validate()

	assert.EqualError(t, err, "refresh rate out of range: 50")
}

func TestStandardTimingEntrySpecificationValidateHeightMustBePositive(t *testing.T) {
	specification := &StandardTimingEntrySpecification{
		Width:       256,
		Height:      0,
		RefreshRate: 75,
		AspectRatio: StandardTimingEntryAspectRatio4x3,
	}

	err := specification.Validate()

	assert.EqualError(t, err, "height must be positive")
}

func TestStandardTimingEntrySpecificationValidateUnsupportedAspectRatio(t *testing.T) {
	specification := &StandardTimingEntrySpecification{
		Width:       256,
		Height:      192,
		RefreshRate: 75,
		AspectRatio: StandardTimingEntryAspectRatio("21:9"),
	}

	err := specification.Validate()

	assert.EqualError(t, err, "unsupported aspect ratio: 21:9")
}

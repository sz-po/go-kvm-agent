package edid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateDisplayFeaturesSpecificationFromBlock(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		block    byte
		expected DisplayFeaturesSpecification
	}{
		{
			name:  "MonochromeOnly",
			block: 0x00,
			expected: DisplayFeaturesSpecification{
				IsMonochrome: true,
			},
		},
		{
			name:  "RgbAllFeatures",
			block: 0xEF,
			expected: DisplayFeaturesSpecification{
				SupportsStandby:            true,
				SupportsSuspend:            true,
				SupportsActiveOff:          true,
				IsRgbColor:                 true,
				UsesStandardSrgbColorSpace: true,
				HasPreferredTimingMode:     true,
				SupportsGeneralizedTiming:  true,
			},
		},
		{
			name:  "NonRgb",
			block: 0x14,
			expected: DisplayFeaturesSpecification{
				SupportsActiveOff: true,
				IsNonRgbColor:     true,
			},
		},
		{
			name:  "UndefinedColor",
			block: 0x98,
			expected: DisplayFeaturesSpecification{
				SupportsGeneralizedTiming: true,
				IsUndefinedColor:          true,
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			specification, err := CreateDisplayFeaturesSpecificationFromBlock(DisplayFeaturesBlock{testCase.block})

			assert.NoError(t, err)
			if assert.NotNil(t, specification) {
				assert.Equal(t, testCase.expected, *specification)
			}
		})
	}
}

func TestCreateDisplayFeaturesBlockFromSpecification(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		specification DisplayFeaturesSpecification
		expectedBlock DisplayFeaturesBlock
	}{
		{
			name: "MonochromeOnly",
			specification: DisplayFeaturesSpecification{
				IsMonochrome: true,
			},
			expectedBlock: DisplayFeaturesBlock{0x00},
		},
		{
			name: "RgbAllFeatures",
			specification: DisplayFeaturesSpecification{
				SupportsStandby:            true,
				SupportsSuspend:            true,
				SupportsActiveOff:          true,
				IsRgbColor:                 true,
				UsesStandardSrgbColorSpace: true,
				HasPreferredTimingMode:     true,
				SupportsGeneralizedTiming:  true,
			},
			expectedBlock: DisplayFeaturesBlock{0xEF},
		},
		{
			name: "NonRgb",
			specification: DisplayFeaturesSpecification{
				SupportsActiveOff: true,
				IsNonRgbColor:     true,
			},
			expectedBlock: DisplayFeaturesBlock{0x14},
		},
		{
			name: "UndefinedColor",
			specification: DisplayFeaturesSpecification{
				SupportsGeneralizedTiming: true,
				IsUndefinedColor:          true,
			},
			expectedBlock: DisplayFeaturesBlock{0x98},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			block, err := CreateDisplayFeaturesBlockFromSpecification(testCase.specification)

			assert.NoError(t, err)
			if assert.NotNil(t, block) {
				assert.Equal(t, testCase.expectedBlock, *block)
			}
		})
	}
}

func TestCreateDisplayFeaturesBlockFromSpecificationInvalid(t *testing.T) {
	t.Parallel()

	specification := DisplayFeaturesSpecification{
		IsMonochrome: true,
		IsRgbColor:   true,
	}

	block, err := CreateDisplayFeaturesBlockFromSpecification(specification)

	assert.Nil(t, block)
	assert.EqualError(t, err, "validate specification: color type must have exactly one flag set")
}

func TestDisplayFeaturesSpecificationValidateRequiresColorType(t *testing.T) {
	t.Parallel()

	specification := &DisplayFeaturesSpecification{}

	err := specification.Validate()

	assert.EqualError(t, err, "color type must have exactly one flag set")
}

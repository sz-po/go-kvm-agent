package edid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func float64Ptr(value float64) *float64 {
	return &value
}

func TestCreateDisplaySpecificationFromBlock(t *testing.T) {
	t.Parallel()

	block := DisplayBlock{0xA1, 30, 40, 120, 0xEF}

	specification, err := CreateDisplaySpecificationFromBlock(block)

	assert.NoError(t, err)
	if assert.NotNil(t, specification) {
		assert.NotNil(t, specification.Input.Digital)
		assert.Nil(t, specification.Input.Analog)

		if specification.Input.Digital != nil {
			if assert.NotNil(t, specification.Input.Digital.ColorBitDepth) {
				assert.Equal(t, uint8(8), *specification.Input.Digital.ColorBitDepth)
			}

			if assert.NotNil(t, specification.Input.Digital.Interface) {
				assert.Equal(t, string(DVIDigitalInterface), string(*specification.Input.Digital.Interface))
			}
		}

		if assert.NotNil(t, specification.Size.Width) {
			assert.Equal(t, 30, *specification.Size.Width)
		}

		if assert.NotNil(t, specification.Size.Height) {
			assert.Equal(t, 40, *specification.Size.Height)
		}

		if assert.NotNil(t, specification.Gamma) {
			assert.InEpsilon(t, 2.2, *specification.Gamma, 0.0001)
		}

		features := specification.Features
		assert.False(t, features.IsMonochrome)
		assert.True(t, features.IsRgbColor)
		assert.False(t, features.IsNonRgbColor)
		assert.False(t, features.IsUndefinedColor)
		assert.True(t, features.SupportsStandby)
		assert.True(t, features.SupportsSuspend)
		assert.True(t, features.SupportsActiveOff)
		assert.True(t, features.UsesStandardSrgbColorSpace)
		assert.True(t, features.HasPreferredTimingMode)
		assert.True(t, features.SupportsGeneralizedTiming)
	}
}

func TestCreateDisplaySpecificationFromBlockInputError(t *testing.T) {
	t.Parallel()

	block := DisplayBlock{}

	specification, err := CreateDisplaySpecificationFromBlock(block)

	assert.Nil(t, specification)
	assert.EqualError(t, err, "input: not supported")
}

func TestCreateDisplayBlockFromSpecification(t *testing.T) {
	t.Parallel()

	specification := DisplaySpecification{
		Input: DisplayInputSpecification{
			Digital: &DisplayDigitalInputSpecification{
				ColorBitDepth: uint8Ptr(8),
				Interface:     digitalInterfacePtr(DVIDigitalInterface),
			},
		},
		Size: DisplaySizeSpecification{
			Width:  intPtr(30),
			Height: intPtr(40),
		},
		Gamma: float64Ptr(2.2),
		Features: DisplayFeaturesSpecification{
			SupportsStandby:            true,
			SupportsSuspend:            true,
			SupportsActiveOff:          true,
			IsRgbColor:                 true,
			UsesStandardSrgbColorSpace: true,
			HasPreferredTimingMode:     true,
			SupportsGeneralizedTiming:  true,
		},
	}

	block, err := CreateDisplayBlockFromSpecification(specification)

	assert.NoError(t, err)
	if assert.NotNil(t, block) {
		assert.Equal(t, DisplayBlock{0xA1, 30, 40, 120, 0xEF}, *block)
	}
}

func TestCreateDisplayBlockFromSpecificationInvalidInput(t *testing.T) {
	t.Parallel()

	specification := DisplaySpecification{
		Input: DisplayInputSpecification{
			Analog: &DisplayAnalogInputSpecification{},
		},
		Size: DisplaySizeSpecification{
			Width:  intPtr(10),
			Height: intPtr(20),
		},
		Features: DisplayFeaturesSpecification{
			IsRgbColor: true,
		},
	}

	block, err := CreateDisplayBlockFromSpecification(specification)

	assert.Nil(t, block)
	assert.EqualError(t, err, "validate: invalid input: not supported")
}

func TestCreateDisplayBlockFromSpecificationMissingSize(t *testing.T) {
	t.Parallel()

	specification := DisplaySpecification{
		Input: DisplayInputSpecification{
			Digital: &DisplayDigitalInputSpecification{
				Interface: digitalInterfacePtr(DVIDigitalInterface),
			},
		},
		Size: DisplaySizeSpecification{
			Width: intPtr(10),
		},
		Features: DisplayFeaturesSpecification{
			IsRgbColor: true,
		},
	}

	block, err := CreateDisplayBlockFromSpecification(specification)

	assert.Nil(t, block)
	assert.ErrorContains(t, err, "validate: size display block")
}

func TestCreateDisplayBlockFromSpecificationFeaturesError(t *testing.T) {
	t.Parallel()

	specification := DisplaySpecification{
		Input: DisplayInputSpecification{
			Digital: &DisplayDigitalInputSpecification{
				Interface: digitalInterfacePtr(DVIDigitalInterface),
			},
		},
		Size: DisplaySizeSpecification{
			Width:  intPtr(10),
			Height: intPtr(20),
		},
		Features: DisplayFeaturesSpecification{
			IsMonochrome: true,
			IsRgbColor:   true,
		},
	}

	block, err := CreateDisplayBlockFromSpecification(specification)

	assert.Nil(t, block)
	assert.EqualError(t, err, "validate: invalid features: color type must have exactly one flag set")
}

func TestDisplaySpecificationValidate(t *testing.T) {
	t.Parallel()

	specification := &DisplaySpecification{
		Input: DisplayInputSpecification{
			Digital: &DisplayDigitalInputSpecification{
				Interface: digitalInterfacePtr(DVIDigitalInterface),
			},
		},
		Size: DisplaySizeSpecification{
			Width:  intPtr(10),
			Height: intPtr(20),
		},
		Features: DisplayFeaturesSpecification{
			IsRgbColor: true,
		},
	}

	assert.NoError(t, specification.Validate())

	specification.Input.Digital = nil
	assert.EqualError(t, specification.Validate(), "invalid input: missing digital/analog specification")

	specification.Input.Digital = &DisplayDigitalInputSpecification{}
	specification.Size.Height = nil
	assert.ErrorContains(t, specification.Validate(), "size display block")

	specification.Size.Height = intPtr(20)
	specification.Features.IsRgbColor = false
	specification.Features.IsMonochrome = true
	assert.NoError(t, specification.Validate())

	specification.Features.IsRgbColor = true
	assert.EqualError(t, specification.Validate(), "invalid features: color type must have exactly one flag set")
}

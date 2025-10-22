package edid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateEstablishedTimingsBlockFromSpecification(t *testing.T) {
	t.Parallel()

	specification := EstablishedTimingsSpecification{
		Supports720x400x70:   true,
		Supports720x400x88:   true,
		Supports640x480x60:   true,
		Supports640x480x67:   true,
		Supports640x480x72:   true,
		Supports640x480x75:   true,
		Supports800x600x56:   true,
		Supports800x600x60:   true,
		Supports800x600x72:   true,
		Supports800x600x75:   true,
		Supports832x624x75:   true,
		Supports1024x768x87i: true,
		Supports1024x768x60:  true,
		Supports1024x768x70:  true,
		Supports1024x768x75:  true,
		Supports1280x1024x75: true,
		Supports1152x870x75:  true,
	}

	block, err := CreateEstablishedTimingsBlockFromSpecification(specification)

	assert.NoError(t, err)
	if assert.NotNil(t, block) {
		assert.Equal(t, EstablishedTimingsBlock{0xFF, 0xFF, 0x80}, *block)

		roundTripSpecification, roundTripErr := CreateEstablishedTimingsSpecificationFromBlock(*block)

		assert.NoError(t, roundTripErr)
		if assert.NotNil(t, roundTripSpecification) {
			assert.Equal(t, specification, *roundTripSpecification)
		}
	}
}

func TestCreateEstablishedTimingsBlockFromSpecificationIndividualFlags(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                       string
		configureSpecification     func(*EstablishedTimingsSpecification)
		expectedEstablishedTimings EstablishedTimingsBlock
	}{
		{
			name: "Supports720x400x70",
			configureSpecification: func(specification *EstablishedTimingsSpecification) {
				specification.Supports720x400x70 = true
			},
			expectedEstablishedTimings: EstablishedTimingsBlock{0x80, 0x00, 0x00},
		},
		{
			name: "Supports720x400x88",
			configureSpecification: func(specification *EstablishedTimingsSpecification) {
				specification.Supports720x400x88 = true
			},
			expectedEstablishedTimings: EstablishedTimingsBlock{0x40, 0x00, 0x00},
		},
		{
			name: "Supports640x480x60",
			configureSpecification: func(specification *EstablishedTimingsSpecification) {
				specification.Supports640x480x60 = true
			},
			expectedEstablishedTimings: EstablishedTimingsBlock{0x20, 0x00, 0x00},
		},
		{
			name: "Supports640x480x67",
			configureSpecification: func(specification *EstablishedTimingsSpecification) {
				specification.Supports640x480x67 = true
			},
			expectedEstablishedTimings: EstablishedTimingsBlock{0x10, 0x00, 0x00},
		},
		{
			name: "Supports640x480x72",
			configureSpecification: func(specification *EstablishedTimingsSpecification) {
				specification.Supports640x480x72 = true
			},
			expectedEstablishedTimings: EstablishedTimingsBlock{0x08, 0x00, 0x00},
		},
		{
			name: "Supports640x480x75",
			configureSpecification: func(specification *EstablishedTimingsSpecification) {
				specification.Supports640x480x75 = true
			},
			expectedEstablishedTimings: EstablishedTimingsBlock{0x04, 0x00, 0x00},
		},
		{
			name: "Supports800x600x56",
			configureSpecification: func(specification *EstablishedTimingsSpecification) {
				specification.Supports800x600x56 = true
			},
			expectedEstablishedTimings: EstablishedTimingsBlock{0x02, 0x00, 0x00},
		},
		{
			name: "Supports800x600x60",
			configureSpecification: func(specification *EstablishedTimingsSpecification) {
				specification.Supports800x600x60 = true
			},
			expectedEstablishedTimings: EstablishedTimingsBlock{0x01, 0x00, 0x00},
		},
		{
			name: "Supports800x600x72",
			configureSpecification: func(specification *EstablishedTimingsSpecification) {
				specification.Supports800x600x72 = true
			},
			expectedEstablishedTimings: EstablishedTimingsBlock{0x00, 0x80, 0x00},
		},
		{
			name: "Supports800x600x75",
			configureSpecification: func(specification *EstablishedTimingsSpecification) {
				specification.Supports800x600x75 = true
			},
			expectedEstablishedTimings: EstablishedTimingsBlock{0x00, 0x40, 0x00},
		},
		{
			name: "Supports832x624x75",
			configureSpecification: func(specification *EstablishedTimingsSpecification) {
				specification.Supports832x624x75 = true
			},
			expectedEstablishedTimings: EstablishedTimingsBlock{0x00, 0x20, 0x00},
		},
		{
			name: "Supports1024x768x87i",
			configureSpecification: func(specification *EstablishedTimingsSpecification) {
				specification.Supports1024x768x87i = true
			},
			expectedEstablishedTimings: EstablishedTimingsBlock{0x00, 0x10, 0x00},
		},
		{
			name: "Supports1024x768x60",
			configureSpecification: func(specification *EstablishedTimingsSpecification) {
				specification.Supports1024x768x60 = true
			},
			expectedEstablishedTimings: EstablishedTimingsBlock{0x00, 0x08, 0x00},
		},
		{
			name: "Supports1024x768x70",
			configureSpecification: func(specification *EstablishedTimingsSpecification) {
				specification.Supports1024x768x70 = true
			},
			expectedEstablishedTimings: EstablishedTimingsBlock{0x00, 0x04, 0x00},
		},
		{
			name: "Supports1024x768x75",
			configureSpecification: func(specification *EstablishedTimingsSpecification) {
				specification.Supports1024x768x75 = true
			},
			expectedEstablishedTimings: EstablishedTimingsBlock{0x00, 0x02, 0x00},
		},
		{
			name: "Supports1280x1024x75",
			configureSpecification: func(specification *EstablishedTimingsSpecification) {
				specification.Supports1280x1024x75 = true
			},
			expectedEstablishedTimings: EstablishedTimingsBlock{0x00, 0x01, 0x00},
		},
		{
			name: "Supports1152x870x75",
			configureSpecification: func(specification *EstablishedTimingsSpecification) {
				specification.Supports1152x870x75 = true
			},
			expectedEstablishedTimings: EstablishedTimingsBlock{0x00, 0x00, 0x80},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			specification := EstablishedTimingsSpecification{}
			testCase.configureSpecification(&specification)

			block, err := CreateEstablishedTimingsBlockFromSpecification(specification)

			assert.NoError(t, err)
			if assert.NotNil(t, block) {
				assert.Equal(t, testCase.expectedEstablishedTimings, *block)

				decodedSpecification, decodeErr := CreateEstablishedTimingsSpecificationFromBlock(*block)

				assert.NoError(t, decodeErr)
				if assert.NotNil(t, decodedSpecification) {
					assert.Equal(t, specification, *decodedSpecification)
				}
			}
		})
	}
}

func TestCreateEstablishedTimingsSpecificationFromBlockAllFalse(t *testing.T) {
	t.Parallel()

	block := EstablishedTimingsBlock{0x00, 0x00, 0x00}

	specification, err := CreateEstablishedTimingsSpecificationFromBlock(block)

	assert.NoError(t, err)
	if assert.NotNil(t, specification) {
		assert.Equal(t, EstablishedTimingsSpecification{}, *specification)
	}
}

func TestCreateEstablishedTimingsSpecificationFromBlockReservedBits(t *testing.T) {
	t.Parallel()

	block := EstablishedTimingsBlock{0x00, 0x00, 0x01}

	specification, err := CreateEstablishedTimingsSpecificationFromBlock(block)

	assert.Nil(t, specification)
	assert.EqualError(t, err, "reserved bits set in established timings byte 2: 0x01")
}

func TestEstablishedTimingsSpecificationValidateNil(t *testing.T) {
	t.Parallel()

	var specification *EstablishedTimingsSpecification

	err := specification.Validate()

	assert.EqualError(t, err, "nil specification")
}

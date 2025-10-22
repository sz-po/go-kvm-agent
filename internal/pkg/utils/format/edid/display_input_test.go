package edid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func uint8Ptr(value uint8) *uint8 {
	return &value
}

func digitalInterfacePtr(value DigitalInterface) *DigitalInterface {
	return &value
}

func TestCreateDisplayDigitalInputBlockFromSpecificationColorBitDepthMapping(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		colorBitDepth      uint8
		expectedBlockValue byte
	}{
		{name: "Depth6Bits", colorBitDepth: 6, expectedBlockValue: 0x90},
		{name: "Depth8Bits", colorBitDepth: 8, expectedBlockValue: 0xA0},
		{name: "Depth10Bits", colorBitDepth: 10, expectedBlockValue: 0xB0},
		{name: "Depth12Bits", colorBitDepth: 12, expectedBlockValue: 0xC0},
		{name: "Depth14Bits", colorBitDepth: 14, expectedBlockValue: 0xD0},
		{name: "Depth16Bits", colorBitDepth: 16, expectedBlockValue: 0xE0},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			specification := DisplayDigitalInputSpecification{
				ColorBitDepth: uint8Ptr(testCase.colorBitDepth),
			}

			block, err := CreateDisplayDigitalInputBlockFromSpecification(specification)

			assert.NoError(t, err)
			if assert.NotNil(t, block) {
				assert.Equal(t, DisplayInputBlock{testCase.expectedBlockValue}, *block)
			}
		})
	}
}

func TestCreateDisplayDigitalInputBlockFromSpecificationInterfaceMapping(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		interfaceValue     DigitalInterface
		expectedBlockValue byte
	}{
		{name: "Undefined", interfaceValue: UndefinedDigitalInterface, expectedBlockValue: 0x80},
		{name: "Dvi", interfaceValue: DVIDigitalInterface, expectedBlockValue: 0x81},
		{name: "Mddi", interfaceValue: MDDIDigitalInterface, expectedBlockValue: 0x84},
		{name: "DisplayPort", interfaceValue: DisplayPortDigitalInterface, expectedBlockValue: 0x85},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			specification := DisplayDigitalInputSpecification{
				Interface: digitalInterfacePtr(testCase.interfaceValue),
			}

			block, err := CreateDisplayDigitalInputBlockFromSpecification(specification)

			assert.NoError(t, err)
			if assert.NotNil(t, block) {
				assert.Equal(t, DisplayInputBlock{testCase.expectedBlockValue}, *block)
			}
		})
	}
}

func TestCreateDisplayDigitalInputBlockFromSpecificationInvalidColorBitDepth(t *testing.T) {
	t.Parallel()

	colorBitDepth := uint8(7)

	block, err := CreateDisplayDigitalInputBlockFromSpecification(DisplayDigitalInputSpecification{
		ColorBitDepth: &colorBitDepth,
	})

	assert.Nil(t, block)
	assert.EqualError(t, err, "unsupported color depth: 7")
}

func TestCreateDisplayDigitalInputBlockFromSpecificationInvalidInterface(t *testing.T) {
	t.Parallel()

	invalidInterface := DigitalInterface("invalid")

	block, err := CreateDisplayDigitalInputBlockFromSpecification(DisplayDigitalInputSpecification{
		Interface: &invalidInterface,
	})

	assert.Nil(t, block)
	assert.EqualError(t, err, "unsupported digital interface: invalid")
}

func TestCreateDisplayDigitalInputSpecificationFromBlockColorBitDepthMapping(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		blockValue    byte
		expectedDepth uint8
	}{
		{name: "Depth6Bits", blockValue: 0x90, expectedDepth: 6},
		{name: "Depth8Bits", blockValue: 0xA0, expectedDepth: 8},
		{name: "Depth10Bits", blockValue: 0xB0, expectedDepth: 10},
		{name: "Depth12Bits", blockValue: 0xC0, expectedDepth: 12},
		{name: "Depth14Bits", blockValue: 0xD0, expectedDepth: 14},
		{name: "Depth16Bits", blockValue: 0xE0, expectedDepth: 16},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			specification, err := CreateDisplayDigitalInputSpecificationFromBlock(DisplayInputBlock{testCase.blockValue})

			assert.NoError(t, err)
			if assert.NotNil(t, specification) {
				if assert.NotNil(t, specification.ColorBitDepth) {
					assert.Equal(t, testCase.expectedDepth, *specification.ColorBitDepth)
				}
				assert.Nil(t, specification.Interface)
			}
		})
	}
}

func TestCreateDisplayDigitalInputSpecificationFromBlockInterfaceMapping(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		blockValue        byte
		expectedInterface DigitalInterface
		expectPointer     bool
	}{
		{name: "Undefined", blockValue: 0x80, expectedInterface: UndefinedDigitalInterface, expectPointer: false},
		{name: "Dvi", blockValue: 0x81, expectedInterface: DVIDigitalInterface, expectPointer: true},
		{name: "Mddi", blockValue: 0x84, expectedInterface: MDDIDigitalInterface, expectPointer: true},
		{name: "DisplayPort", blockValue: 0x85, expectedInterface: DisplayPortDigitalInterface, expectPointer: true},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			specification, err := CreateDisplayDigitalInputSpecificationFromBlock(DisplayInputBlock{testCase.blockValue})

			assert.NoError(t, err)
			if assert.NotNil(t, specification) {
				assert.Nil(t, specification.ColorBitDepth)
				if testCase.expectPointer {
					if assert.NotNil(t, specification.Interface) {
						assert.Equal(t, testCase.expectedInterface, *specification.Interface)
					}
				} else {
					assert.Nil(t, specification.Interface)
				}
			}
		})
	}
}

func TestCreateDisplayDigitalInputSpecificationFromBlockReservedInterface(t *testing.T) {
	t.Parallel()

	specification, err := CreateDisplayDigitalInputSpecificationFromBlock(DisplayInputBlock{0x82})

	assert.Nil(t, specification)
	assert.EqualError(t, err, "unsupported digital interface encoding: 2")
}

func TestCreateDisplayDigitalInputSpecificationFromBlockInvalidDigitalFlag(t *testing.T) {
	t.Parallel()

	block := DisplayInputBlock{0x22}

	specification, err := CreateDisplayDigitalInputSpecificationFromBlock(block)

	assert.Nil(t, specification)
	assert.EqualError(t, err, "not a digital input block")
}

func TestCreateDisplayDigitalInputSpecificationFromBlockInvalidColorEncoding(t *testing.T) {
	t.Parallel()

	block := DisplayInputBlock{0xF0}

	specification, err := CreateDisplayDigitalInputSpecificationFromBlock(block)

	assert.Nil(t, specification)
	assert.EqualError(t, err, "unsupported color bit depth encoding: 7")
}

func TestCreateDisplayDigitalInputSpecificationFromBlockInvalidInterfaceEncoding(t *testing.T) {
	t.Parallel()

	block := DisplayInputBlock{0x8F}

	specification, err := CreateDisplayDigitalInputSpecificationFromBlock(block)

	assert.Nil(t, specification)
	assert.EqualError(t, err, "unsupported digital interface encoding: 15")
}

func TestDisplayDigitalInputSpecificationValidate(t *testing.T) {
	t.Parallel()

	specification := &DisplayDigitalInputSpecification{}

	assert.NoError(t, specification.Validate())

	oddColorBitDepth := uint8(7)
	specification.ColorBitDepth = &oddColorBitDepth

	err := specification.Validate()
	assert.EqualError(t, err, "invalid color bit depth")

	specification.ColorBitDepth = nil
	invalidInterface := DigitalInterface("invalid")
	specification.Interface = &invalidInterface

	err = specification.Validate()
	assert.EqualError(t, err, "invalid interface")
}

func TestDisplayDigitalInputSpecificationValidateWithValidatorConstraints(t *testing.T) {
	t.Parallel()

	colorBitDepth := uint8(5)
	specification := &DisplayDigitalInputSpecification{
		ColorBitDepth: &colorBitDepth,
	}

	err := specification.Validate()
	assert.Error(t, err)
}

func TestCreateDisplayAnalogInputSpecificationFromBlockDigitalFlag(t *testing.T) {
	t.Parallel()

	block := DisplayInputBlock{0x80}

	specification, err := CreateDisplayAnalogInputSpecificationFromBlock(block)

	assert.Nil(t, specification)
	assert.EqualError(t, err, "not an analog input block")
}

func TestCreateDisplayAnalogInputSpecificationFromBlockNotSupported(t *testing.T) {
	t.Parallel()

	block := DisplayInputBlock{0x00}

	specification, err := CreateDisplayAnalogInputSpecificationFromBlock(block)

	assert.Nil(t, specification)
	assert.EqualError(t, err, "not supported")
}

func TestCreateDisplayAnalogInputBlockFromSpecificationNotSupported(t *testing.T) {
	t.Parallel()

	block, err := CreateDisplayAnalogInputBlockFromSpecification(DisplayAnalogInputSpecification{})

	assert.Nil(t, block)
	assert.EqualError(t, err, "not supported")
}

func TestDisplayAnalogInputSpecificationValidateNotSupported(t *testing.T) {
	t.Parallel()

	specification := &DisplayAnalogInputSpecification{}

	err := specification.Validate()
	assert.EqualError(t, err, "not supported")
}

func TestCreateDisplayInputBlockFromSpecification(t *testing.T) {
	t.Parallel()

	specification := DisplayInputSpecification{
		Digital: &DisplayDigitalInputSpecification{
			ColorBitDepth: uint8Ptr(8),
			Interface:     digitalInterfacePtr(DigitalInterface(DVIDigitalInterface)),
		},
	}

	block, err := CreateDisplayInputBlockFromSpecification(specification)

	assert.NoError(t, err)
	if assert.NotNil(t, block) {
		assert.Equal(t, DisplayInputBlock{0xA1}, *block)
	}
}

func TestCreateDisplayInputBlockFromSpecificationAnalogNotSupported(t *testing.T) {
	t.Parallel()

	specification := DisplayInputSpecification{
		Analog: &DisplayAnalogInputSpecification{},
	}

	block, err := CreateDisplayInputBlockFromSpecification(specification)

	assert.Nil(t, block)
	assert.EqualError(t, err, "specification validation: not supported")
}

func TestCreateDisplayInputBlockFromSpecificationMissingSpecification(t *testing.T) {
	t.Parallel()

	specification := DisplayInputSpecification{}

	block, err := CreateDisplayInputBlockFromSpecification(specification)

	assert.Nil(t, block)
	assert.EqualError(t, err, "specification validation: missing digital/analog specification")
}

func TestCreateDisplayInputSpecificationFromBlock(t *testing.T) {
	t.Parallel()

	block := DisplayInputBlock{0xA1}

	specification, err := CreateDisplayInputSpecificationFromBlock(block)

	assert.NoError(t, err)
	if assert.NotNil(t, specification) {
		assert.NotNil(t, specification.Digital)
		assert.Nil(t, specification.Analog)
	}
}

func TestCreateDisplayInputSpecificationFromBlockAnalogNotSupported(t *testing.T) {
	t.Parallel()

	block := DisplayInputBlock{0x00}

	specification, err := CreateDisplayInputSpecificationFromBlock(block)

	assert.Nil(t, specification)
	assert.EqualError(t, err, "not supported")
}

func TestDisplayInputSpecificationValidate(t *testing.T) {
	t.Parallel()

	specification := &DisplayInputSpecification{}

	err := specification.Validate()
	assert.EqualError(t, err, "missing digital/analog specification")

	specification.Digital = &DisplayDigitalInputSpecification{}
	specification.Analog = &DisplayAnalogInputSpecification{}

	err = specification.Validate()
	assert.EqualError(t, err, "multiple input specifications provided")

	specification.Analog = nil
	oddColorBitDepth := uint8(7)
	specification.Digital.ColorBitDepth = &oddColorBitDepth

	err = specification.Validate()
	assert.EqualError(t, err, "invalid color bit depth")
}

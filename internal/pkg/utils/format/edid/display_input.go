package edid

import (
	"fmt"
	"slices"

	"github.com/go-playground/validator/v10"
)

type DisplayInputBlock [1]byte

type DigitalInterface string

const (
	UndefinedDigitalInterface   DigitalInterface = ""
	DVIDigitalInterface                          = "dvi"
	MDDIDigitalInterface                         = "mddi"
	DisplayPortDigitalInterface                  = "display-port"
)

const digitalInputFlag byte = 1 << 7

type DisplayDigitalInputSpecification struct {
	ColorBitDepth *uint8            `json:"colorBitDepth" validate:"omitempty,min=6,max=16"`
	Interface     *DigitalInterface `json:"interface" validate:"omitempty"`
}

func CreateDisplayDigitalInputSpecificationFromBlock(block DisplayInputBlock) (*DisplayDigitalInputSpecification, error) {
	if block[0]&digitalInputFlag == 0 {
		return nil, fmt.Errorf("not a digital input block")
	}

	specification := &DisplayDigitalInputSpecification{}

	colorEncoding := (block[0] >> 4) & 0b111

	if colorEncoding != 0 {
		colorBitDepthLookup := map[byte]uint8{
			0b001: 6,
			0b010: 8,
			0b011: 10,
			0b100: 12,
			0b101: 14,
			0b110: 16,
		}

		colorBitDepth, colorEncodingSupported := colorBitDepthLookup[colorEncoding]

		if !colorEncodingSupported {
			return nil, fmt.Errorf("unsupported color bit depth encoding: %d", colorEncoding)
		}

		specification.ColorBitDepth = &colorBitDepth
	}

	digitalInterfaceEncoding := block[0] & 0b1111

	if digitalInterfaceEncoding != 0 {
		interfaceLookup := map[byte]DigitalInterface{
			0b0001: DVIDigitalInterface,
			0b0100: MDDIDigitalInterface,
			0b0101: DisplayPortDigitalInterface,
		}

		interfaceValue, interfaceEncodingSupported := interfaceLookup[digitalInterfaceEncoding]

		if !interfaceEncodingSupported {
			return nil, fmt.Errorf("unsupported digital interface encoding: %d", digitalInterfaceEncoding)
		}

		interfaceCopy := interfaceValue
		specification.Interface = &interfaceCopy
	}

	return specification, nil
}

func CreateDisplayDigitalInputBlockFromSpecification(specification DisplayDigitalInputSpecification) (*DisplayInputBlock, error) {
	block := &DisplayInputBlock{}

	// Digital
	block[0] |= digitalInputFlag

	if specification.ColorBitDepth != nil {
		var colorBitDepthByte byte
		switch *specification.ColorBitDepth {
		case 6:
			colorBitDepthByte = 0b001
		case 8:
			colorBitDepthByte = 0b010
		case 10:
			colorBitDepthByte = 0b011
		case 12:
			colorBitDepthByte = 0b100
		case 14:
			colorBitDepthByte = 0b101
		case 16:
			colorBitDepthByte = 0b110
		default:
			return nil, fmt.Errorf("unsupported color depth: %d", *specification.ColorBitDepth)
		}

		block[0] &^= 0b01110000
		block[0] |= colorBitDepthByte << 4
	}

	if specification.Interface != nil {
		var interfaceByte byte
		switch *specification.Interface {
		case UndefinedDigitalInterface:
			interfaceByte = 0b0000
		case DVIDigitalInterface:
			interfaceByte = 0b0001
		case MDDIDigitalInterface:
			interfaceByte = 0b0100
		case DisplayPortDigitalInterface:
			interfaceByte = 0b0101
		default:
			return nil, fmt.Errorf("unsupported digital interface: %s", *specification.Interface)
		}

		block[0] &^= 0b00001111
		block[0] |= interfaceByte
	}

	return block, nil
}

func (specification *DisplayDigitalInputSpecification) Validate() error {
	specificationValidator := validator.New(validator.WithRequiredStructEnabled())

	if err := specificationValidator.Struct(specification); err != nil {
		return err
	}

	if specification.ColorBitDepth != nil && *specification.ColorBitDepth%2 != 0 {
		return fmt.Errorf("invalid color bit depth")
	}

	if specification.Interface != nil {
		if !slices.Contains(
			[]DigitalInterface{
				UndefinedDigitalInterface,
				DVIDigitalInterface,
				MDDIDigitalInterface,
				DisplayPortDigitalInterface},
			*specification.Interface) {
			return fmt.Errorf("invalid interface")
		}
	}

	return nil
}

type DisplayAnalogInputSpecification struct {
}

func CreateDisplayAnalogInputSpecificationFromBlock(block DisplayInputBlock) (*DisplayAnalogInputSpecification, error) {
	if block[0]&digitalInputFlag != 0 {
		return nil, fmt.Errorf("not an analog input block")
	}

	return nil, fmt.Errorf("not supported")
}

func CreateDisplayAnalogInputBlockFromSpecification(specification DisplayAnalogInputSpecification) (*DisplayInputBlock, error) {
	return nil, fmt.Errorf("not supported")
}

func (specification *DisplayAnalogInputSpecification) Validate() error {
	return fmt.Errorf("not supported")
}

type DisplayInputSpecification struct {
	Digital *DisplayDigitalInputSpecification `json:"digital"`
	Analog  *DisplayAnalogInputSpecification  `json:"analog"`
}

func CreateDisplayInputSpecificationFromBlock(block DisplayInputBlock) (*DisplayInputSpecification, error) {
	specification := &DisplayInputSpecification{}

	isDigital := block[0]&digitalInputFlag != 0

	if isDigital {
		if digitalInputSpecification, err := CreateDisplayDigitalInputSpecificationFromBlock(block); err != nil {
			return nil, err
		} else {
			specification.Digital = digitalInputSpecification
		}
	} else {
		if analogInputSpecification, err := CreateDisplayAnalogInputSpecificationFromBlock(block); err != nil {
			return nil, err
		} else {
			specification.Analog = analogInputSpecification
		}
	}

	if err := specification.Validate(); err != nil {
		return nil, err
	}

	return specification, nil
}

func CreateDisplayInputBlockFromSpecification(specification DisplayInputSpecification) (*DisplayInputBlock, error) {
	if err := specification.Validate(); err != nil {
		return nil, fmt.Errorf("specification validation: %w", err)
	}

	if specification.Digital != nil {
		return CreateDisplayDigitalInputBlockFromSpecification(*specification.Digital)
	} else if specification.Analog != nil {
		return CreateDisplayAnalogInputBlockFromSpecification(*specification.Analog)
	} else {
		return nil, fmt.Errorf("missing digital/analog specification")
	}
}

func (specification *DisplayInputSpecification) Validate() error {
	digitalSpecificationProvided := specification.Digital != nil
	analogSpecificationProvided := specification.Analog != nil

	if digitalSpecificationProvided && analogSpecificationProvided {
		return fmt.Errorf("multiple input specifications provided")
	}

	if !digitalSpecificationProvided && !analogSpecificationProvided {
		return fmt.Errorf("missing digital/analog specification")
	}

	if digitalSpecificationProvided {
		return specification.Digital.Validate()
	}

	return specification.Analog.Validate()
}

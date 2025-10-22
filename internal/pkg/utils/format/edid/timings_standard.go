package edid

import "fmt"

type StandardTimingsBlock [16]byte

type StandardTimingEntryAspectRatio string

const (
	StandardTimingEntryAspectRatio16x9  StandardTimingEntryAspectRatio = "16:9"
	StandardTimingEntryAspectRatio4x3                                  = "4:3"
	StandardTimingEntryAspectRatio5x4                                  = "5:4"
	StandardTimingEntryAspectRatio16x10                                = "16:10"
)

type StandardTimingEntrySpecification struct {
	Width       int                            `json:"width" validate:"required,min=256,max=2288"`
	Height      int                            `json:"height" validate:"required,min=200,max=1920"`
	RefreshRate int                            `json:"refreshRate" validate:"required,min=60,max=123"`
	AspectRatio StandardTimingEntryAspectRatio `json:"aspectRatio" validate:"required"`
}

type StandardTimingsSpecification struct {
	Entries []StandardTimingEntrySpecification `json:"entries,omitempty" validate:"omitempty,dive"`
}

var (
	standardTimingAspectRatioByBits = map[byte]StandardTimingEntryAspectRatio{
		0b00: StandardTimingEntryAspectRatio16x10,
		0b01: StandardTimingEntryAspectRatio4x3,
		0b10: StandardTimingEntryAspectRatio5x4,
		0b11: StandardTimingEntryAspectRatio16x9,
	}

	standardTimingAspectRatioToBits = map[StandardTimingEntryAspectRatio]byte{
		StandardTimingEntryAspectRatio16x10: 0b00,
		StandardTimingEntryAspectRatio4x3:   0b01,
		StandardTimingEntryAspectRatio5x4:   0b10,
		StandardTimingEntryAspectRatio16x9:  0b11,
	}
)

func CreateStandardTimingsSpecificationFromBlock(block StandardTimingsBlock) (*StandardTimingsSpecification, error) {
	specification := &StandardTimingsSpecification{}

	for index := 0; index < len(block); index += 2 {
		widthByte := block[index]
		heightByte := block[index+1]

		if (widthByte == 0x00 && heightByte == 0x00) || (widthByte == 0x01 && heightByte == 0x01) {
			continue
		}

		width := (int(widthByte) + 31) * 8
		refreshRate := int(heightByte&0b00111111) + 60

		aspectBits := heightByte >> 6
		aspectRatio, aspectRatioSupported := standardTimingAspectRatioByBits[aspectBits]

		if !aspectRatioSupported {
			return nil, fmt.Errorf("unsupported aspect ratio bits: 0b%b", aspectBits)
		}

		height, err := deriveHeight(width, aspectRatio)
		if err != nil {
			return nil, fmt.Errorf("derive height: %w", err)
		}

		entry := StandardTimingEntrySpecification{
			Width:       width,
			Height:      height,
			RefreshRate: refreshRate,
			AspectRatio: aspectRatio,
		}

		if err := entry.Validate(); err != nil {
			return nil, fmt.Errorf("validate entry: %w", err)
		}

		specification.Entries = append(specification.Entries, entry)
	}

	if err := specification.Validate(); err != nil {
		return nil, fmt.Errorf("validate specification: %w", err)
	}

	return specification, nil
}

func CreateStandardTimingsBlockFromSpecification(specification StandardTimingsSpecification) (*StandardTimingsBlock, error) {
	if err := specification.Validate(); err != nil {
		return nil, fmt.Errorf("validate specification: %w", err)
	}

	if len(specification.Entries) > 8 {
		return nil, fmt.Errorf("too many timing entries: %d", len(specification.Entries))
	}

	block := &StandardTimingsBlock{}

	for index, entry := range specification.Entries {
		if err := entry.Validate(); err != nil {
			return nil, fmt.Errorf("validate entry: %w", err)
		}

		widthByte, err := deriveWidthByte(entry.Width)
		if err != nil {
			return nil, fmt.Errorf("derive width byte: %w", err)
		}

		bits, ratioSupported := standardTimingAspectRatioToBits[entry.AspectRatio]
		if !ratioSupported {
			return nil, fmt.Errorf("unsupported aspect ratio: %s", entry.AspectRatio)
		}

		entryHeight, heightErr := deriveHeight(entry.Width, entry.AspectRatio)
		if heightErr != nil {
			return nil, fmt.Errorf("derive height: %w", heightErr)
		}

		if entryHeight != entry.Height {
			return nil, fmt.Errorf("height %d does not match aspect ratio %s", entry.Height, entry.AspectRatio)
		}

		if entry.RefreshRate < 60 || entry.RefreshRate > 123 {
			return nil, fmt.Errorf("refresh rate out of range: %d", entry.RefreshRate)
		}

		refreshRate := byte(entry.RefreshRate - 60)

		block[2*index] = widthByte
		block[2*index+1] = (bits << 6) | refreshRate
	}

	for index := len(specification.Entries); index < 8; index++ {
		block[2*index] = 0x01
		block[2*index+1] = 0x01
	}

	return block, nil
}

func (specification *StandardTimingEntrySpecification) Validate() error {
	if specification.Width < 256 || specification.Width > 2288 {
		return fmt.Errorf("width out of range: %d", specification.Width)
	}

	if specification.RefreshRate < 60 || specification.RefreshRate > 123 {
		return fmt.Errorf("refresh rate out of range: %d", specification.RefreshRate)
	}

	if specification.Height <= 0 {
		return fmt.Errorf("height must be positive")
	}

	if _, supported := standardTimingAspectRatioToBits[specification.AspectRatio]; !supported {
		return fmt.Errorf("unsupported aspect ratio: %s", specification.AspectRatio)
	}

	derivedHeight, err := deriveHeight(specification.Width, specification.AspectRatio)
	if err != nil {
		return fmt.Errorf("derive height: %w", err)
	}

	if specification.Height != derivedHeight {
		return fmt.Errorf("height %d does not match aspect ratio %s", specification.Height, specification.AspectRatio)
	}

	return nil
}

func (specification *StandardTimingsSpecification) Validate() error {
	if specification == nil {
		return fmt.Errorf("nil specification")
	}

	if len(specification.Entries) == 0 {
		return nil
	}

	if len(specification.Entries) > 8 {
		return fmt.Errorf("too many timing entries: %d", len(specification.Entries))
	}

	for index := range specification.Entries {
		entry := specification.Entries[index]
		if err := entry.Validate(); err != nil {
			return fmt.Errorf("entry %d invalid: %w", index, err)
		}
	}

	return nil
}

func deriveWidthByte(width int) (byte, error) {
	if width%8 != 0 {
		return 0, fmt.Errorf("width must be divisible by 8")
	}

	widthValue := width/8 - 31

	if widthValue < 0 || widthValue > 255 {
		return 0, fmt.Errorf("width byte out of range")
	}

	return byte(widthValue), nil
}

func deriveHeight(width int, aspectRatio StandardTimingEntryAspectRatio) (int, error) {
	switch aspectRatio {
	case StandardTimingEntryAspectRatio16x9:
		return width * 9 / 16, nil
	case StandardTimingEntryAspectRatio4x3:
		return width * 3 / 4, nil
	case StandardTimingEntryAspectRatio5x4:
		return width * 4 / 5, nil
	case StandardTimingEntryAspectRatio16x10:
		return width * 10 / 16, nil
	default:
		return 0, fmt.Errorf("unsupported aspect ratio: %s", aspectRatio)
	}
}

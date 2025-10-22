package edid

import (
	"fmt"
	"math"

	"github.com/go-playground/validator/v10"
)

type ChromacityBlock [10]byte

type ChromacitySpecification struct {
	RedX   float64 `json:"redX" validate:"required,min=0,max=1"`
	RedY   float64 `json:"redY" validate:"required,min=0,max=1"`
	GreenX float64 `json:"greenX" validate:"required,min=0,max=1"`
	GreenY float64 `json:"greenY" validate:"required,min=0,max=1"`
	BlueX  float64 `json:"blueX" validate:"required,min=0,max=1"`
	BlueY  float64 `json:"blueY" validate:"required,min=0,max=1"`
	WhiteX float64 `json:"whiteX" validate:"required,min=0,max=1"`
	WhiteY float64 `json:"whiteY" validate:"required,min=0,max=1"`
}

const (
	chromacityResolution     = 1024.0
	chromacityCoordinateMask = 0x03
)

func CreateChromacityBlockFromSpecification(specification ChromacitySpecification) (*ChromacityBlock, error) {
	if err := specification.Validate(); err != nil {
		return nil, fmt.Errorf("validate specification: %w", err)
	}

	block := &ChromacityBlock{}

	coordinates := []struct {
		value    float64
		highIdx  int
		lowByte  int
		lowShift uint
	}{
		{specification.RedX, 2, 0, 6},
		{specification.RedY, 3, 0, 4},
		{specification.GreenX, 4, 0, 2},
		{specification.GreenY, 5, 0, 0},
		{specification.BlueX, 6, 1, 6},
		{specification.BlueY, 7, 1, 4},
		{specification.WhiteX, 8, 1, 2},
		{specification.WhiteY, 9, 1, 0},
	}

	for _, coordinate := range coordinates {
		quantized := encodeChromacityCoordinate(coordinate.value)
		block[coordinate.highIdx] = byte(quantized >> 2)
		lowBits := byte(quantized & chromacityCoordinateMask)
		block[coordinate.lowByte] |= lowBits << coordinate.lowShift
	}

	return block, nil
}

func CreateChromacitySpecificationFromBlock(block ChromacityBlock) (*ChromacitySpecification, error) {
	specification := &ChromacitySpecification{}

	coordinates := []struct {
		target   *float64
		highByte byte
		lowBits  byte
	}{
		{&specification.RedX, block[2], extractChromacityLowBits(block[0], 6)},
		{&specification.RedY, block[3], extractChromacityLowBits(block[0], 4)},
		{&specification.GreenX, block[4], extractChromacityLowBits(block[0], 2)},
		{&specification.GreenY, block[5], extractChromacityLowBits(block[0], 0)},
		{&specification.BlueX, block[6], extractChromacityLowBits(block[1], 6)},
		{&specification.BlueY, block[7], extractChromacityLowBits(block[1], 4)},
		{&specification.WhiteX, block[8], extractChromacityLowBits(block[1], 2)},
		{&specification.WhiteY, block[9], extractChromacityLowBits(block[1], 0)},
	}

	for _, coordinate := range coordinates {
		quantized := uint16(coordinate.highByte)<<2 | uint16(coordinate.lowBits)
		*coordinate.target = decodeChromacityCoordinate(quantized)
	}

	if err := specification.Validate(); err != nil {
		return nil, fmt.Errorf("validate specification: %w", err)
	}

	return specification, nil
}

func (specification *ChromacitySpecification) Validate() error {
	specificationValidator := validator.New(validator.WithPrivateFieldValidation())

	if err := specificationValidator.Struct(specification); err != nil {
		return err
	}

	return nil
}

func encodeChromacityCoordinate(value float64) uint16 {
	quantized := uint16(math.Round(value * chromacityResolution))

	if quantized > 0x3FF {
		return 0x3FF
	}

	return quantized
}

func decodeChromacityCoordinate(value uint16) float64 {
	return float64(value) / chromacityResolution
}

func extractChromacityLowBits(source byte, shift uint) byte {
	return (source >> shift) & chromacityCoordinateMask
}

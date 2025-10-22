package edid

import (
	"github.com/go-playground/validator/v10"
)

type DisplaySizeBlock [2]byte

type DisplaySizeSpecification struct {
	Width  *int `json:"width" validate:"required_with=Height,omitempty,min=1,max=255"`
	Height *int `json:"height" validate:"required_with=Width,omitempty,min=1,max=255"`
}

func CreateDisplaySizeSpecificationFromBlock(block DisplaySizeBlock) (*DisplaySizeSpecification, error) {
	specification := &DisplaySizeSpecification{}

	width := int(block[0])
	height := int(block[1])

	if width != 0 {
		specification.Width = &width
	}

	if height != 0 {
		specification.Height = &height
	}

	if err := specification.Validate(); err != nil {
		return nil, err
	}

	return specification, nil
}

func CreateDisplaySizeBlockFromSpecification(specification DisplaySizeSpecification) (*DisplaySizeBlock, error) {
	if err := specification.Validate(); err != nil {
		return nil, err
	}

	block := DisplaySizeBlock{}

	if specification.Width != nil {
		block[0] = byte(*specification.Width)
	}

	if specification.Height != nil {
		block[1] = byte(*specification.Height)
	}

	return &block, nil
}

func (specification *DisplaySizeSpecification) Validate() error {
	specificationValidator := validator.New(validator.WithRequiredStructEnabled())

	if err := specificationValidator.Struct(specification); err != nil {
		return err
	}

	return nil
}

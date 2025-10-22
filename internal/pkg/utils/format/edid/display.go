package edid

import "fmt"

type DisplayBlock [5]byte

type DisplaySpecification struct {
	Input    DisplayInputSpecification    `json:"input" validate:"required"`
	Size     DisplaySizeSpecification     `json:"size" validate:"required"`
	Gamma    *float64                     `json:"gamma" validate:"omitempty,min=1,max=3.55"`
	Features DisplayFeaturesSpecification `json:"features" validate:"required"`
}

func CreateDisplaySpecificationFromBlock(block DisplayBlock) (*DisplaySpecification, error) {
	specification := &DisplaySpecification{}

	if inputSpecification, err := CreateDisplayInputSpecificationFromBlock(DisplayInputBlock(block[0:1])); err != nil {
		return nil, fmt.Errorf("input: %w", err)
	} else {
		specification.Input = *inputSpecification
	}

	if sizeSpecification, err := CreateDisplaySizeSpecificationFromBlock(DisplaySizeBlock(block[1:3])); err != nil {
		return nil, fmt.Errorf("size: %w", err)
	} else {
		specification.Size = *sizeSpecification
	}

	if block[3] != 0 {
		gamma := float64(block[3]+100) / 100
		specification.Gamma = &gamma
	}

	if featuresSpecification, err := CreateDisplayFeaturesSpecificationFromBlock(DisplayFeaturesBlock(block[4:5])); err != nil {
		return nil, fmt.Errorf("features: %w", err)
	} else {
		specification.Features = *featuresSpecification
	}

	return specification, nil
}

func CreateDisplayBlockFromSpecification(specification DisplaySpecification) (*DisplayBlock, error) {
	if err := specification.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	block := &DisplayBlock{}

	if inputBlock, err := CreateDisplayInputBlockFromSpecification(specification.Input); err != nil {
		return nil, fmt.Errorf("input: %w", err)
	} else {
		copy(block[0:1], inputBlock[0:1])
	}

	if sizeBlock, err := CreateDisplaySizeBlockFromSpecification(specification.Size); err != nil {
		return nil, fmt.Errorf("size: %w", err)
	} else {
		copy(block[1:3], sizeBlock[0:2])
	}

	if specification.Gamma != nil {
		block[3] = byte(*specification.Gamma*100 - 100)
	}

	if featuresBlock, err := CreateDisplayFeaturesBlockFromSpecification(specification.Features); err != nil {
		return nil, fmt.Errorf("features: %w", err)
	} else {
		copy(block[4:5], featuresBlock[0:1])
	}

	return block, nil
}

func (specification *DisplaySpecification) Validate() error {
	if specification == nil {
		return fmt.Errorf("nil specification")
	}

	if specification.Gamma != nil {
		gamma := *specification.Gamma
		if gamma < 1 || gamma > 3.55 {
			return fmt.Errorf("invalid gamma value")
		}
	}

	if err := specification.Input.Validate(); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	if err := specification.Size.Validate(); err != nil {
		return fmt.Errorf("size display block: %w", err)
	}

	if err := specification.Features.Validate(); err != nil {
		return fmt.Errorf("invalid features: %w", err)
	}

	return nil
}

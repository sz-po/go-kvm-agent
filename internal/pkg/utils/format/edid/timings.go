package edid

import "fmt"

type TimingsBlock [91]byte

type TimingsSpecification struct {
	Established EstablishedTimingsSpecification `json:"established" validate:"required"`
	Standard    StandardTimingsSpecification    `json:"standard" validate:"required"`
	Detailed    DetailedTimingsSpecification    `json:"detailed" validate:"required"`
}

func CreateTimingsSpecificationFromBlock(block TimingsBlock) (*TimingsSpecification, error) {
	specification := &TimingsSpecification{}

	if establishedSpecification, err := CreateEstablishedTimingsSpecificationFromBlock(EstablishedTimingsBlock(block[0:3])); err != nil {
		return nil, fmt.Errorf("established timings: %w", err)
	} else {
		specification.Established = *establishedSpecification
	}

	if standardSpecification, err := CreateStandardTimingsSpecificationFromBlock(StandardTimingsBlock(block[3:19])); err != nil {
		return nil, fmt.Errorf("standard timings: %w", err)
	} else {
		specification.Standard = *standardSpecification
	}

	if detailedSpecification, err := CreateDetailedTimingsSpecificationFromBlock(DetailedTimingsBlock(block[19:91])); err != nil {
		return nil, fmt.Errorf("detailed timings: %w", err)
	} else {
		specification.Detailed = *detailedSpecification
	}

	return specification, nil
}

func CreateTimingsBlockFromSpecification(specification TimingsSpecification) (*TimingsBlock, error) {
	if err := specification.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	block := &TimingsBlock{}

	if establishedBlock, err := CreateEstablishedTimingsBlockFromSpecification(specification.Established); err != nil {
		return nil, fmt.Errorf("established timings: %w", err)
	} else {
		copy(block[0:3], establishedBlock[0:3])
	}

	if standardBlock, err := CreateStandardTimingsBlockFromSpecification(specification.Standard); err != nil {
		return nil, fmt.Errorf("standard timings: %w", err)
	} else {
		copy(block[3:19], standardBlock[0:16])
	}

	if detailedBlock, err := CreateDetailedTimingsBlockFromSpecification(specification.Detailed); err != nil {
		return nil, fmt.Errorf("detailed timings: %w", err)
	} else {
		copy(block[19:91], detailedBlock[0:72])
	}

	return block, nil
}

func (specification *TimingsSpecification) Validate() error {
	if specification == nil {
		return fmt.Errorf("nil specification")
	}

	if err := specification.Established.Validate(); err != nil {
		return fmt.Errorf("established timings: %w", err)
	}

	if err := specification.Standard.Validate(); err != nil {
		return fmt.Errorf("standard timings: %w", err)
	}

	if err := specification.Detailed.Validate(); err != nil {
		return fmt.Errorf("detailed timings: %w", err)
	}

	return nil
}

package edid

import "fmt"

type DisplayFeaturesBlock [1]byte

type DisplayFeaturesSpecification struct {
	SupportsStandby            bool `json:"supportsStandby"`
	SupportsSuspend            bool `json:"supportsSuspend"`
	SupportsActiveOff          bool `json:"supportsActiveOff"`
	IsMonochrome               bool `json:"isMonochrome"`
	IsRgbColor                 bool `json:"isRgbColor"`
	IsNonRgbColor              bool `json:"isNonRgbColor"`
	IsUndefinedColor           bool `json:"isUndefinedColor"`
	UsesStandardSrgbColorSpace bool `json:"usesStandardSrgbColorSpace"`
	HasPreferredTimingMode     bool `json:"hasPreferredTimingMode"`
	SupportsGeneralizedTiming  bool `json:"supportsGeneralizedTiming"`
}

func CreateDisplayFeaturesSpecificationFromBlock(block DisplayFeaturesBlock) (*DisplayFeaturesSpecification, error) {
	specification := &DisplayFeaturesSpecification{}

	featuresValue := block[0]

	specification.SupportsStandby = featuresValue&0b00000001 != 0
	specification.SupportsSuspend = featuresValue&0b00000010 != 0
	specification.SupportsActiveOff = featuresValue&0b00000100 != 0

	colorTypeBits := (featuresValue >> 3) & 0b11

	switch colorTypeBits {
	case 0b00:
		specification.IsMonochrome = true
	case 0b01:
		specification.IsRgbColor = true
	case 0b10:
		specification.IsNonRgbColor = true
	case 0b11:
		specification.IsUndefinedColor = true
	}

	specification.UsesStandardSrgbColorSpace = featuresValue&0b00100000 != 0
	specification.HasPreferredTimingMode = featuresValue&0b01000000 != 0
	specification.SupportsGeneralizedTiming = featuresValue&0b10000000 != 0

	if err := specification.Validate(); err != nil {
		return nil, fmt.Errorf("validate specification: %w", err)
	}

	return specification, nil
}

func CreateDisplayFeaturesBlockFromSpecification(specification DisplayFeaturesSpecification) (*DisplayFeaturesBlock, error) {
	if err := specification.Validate(); err != nil {
		return nil, fmt.Errorf("validate specification: %w", err)
	}

	var value byte

	if specification.SupportsStandby {
		value |= 0b00000001
	}

	if specification.SupportsSuspend {
		value |= 0b00000010
	}

	if specification.SupportsActiveOff {
		value |= 0b00000100
	}

	switch {
	case specification.IsMonochrome:
	case specification.IsRgbColor:
		value |= 0b00001000
	case specification.IsNonRgbColor:
		value |= 0b00010000
	case specification.IsUndefinedColor:
		value |= 0b00011000
	}

	if specification.UsesStandardSrgbColorSpace {
		value |= 0b00100000
	}

	if specification.HasPreferredTimingMode {
		value |= 0b01000000
	}

	if specification.SupportsGeneralizedTiming {
		value |= 0b10000000
	}

	block := &DisplayFeaturesBlock{}
	block[0] = value

	return block, nil
}

func (specification *DisplayFeaturesSpecification) Validate() error {
	colorTypeFlags := []bool{
		specification.IsMonochrome,
		specification.IsRgbColor,
		specification.IsNonRgbColor,
		specification.IsUndefinedColor,
	}

	colorTypeEnabledCount := 0

	for _, flag := range colorTypeFlags {
		if flag {
			colorTypeEnabledCount++
		}
	}

	if colorTypeEnabledCount != 1 {
		return fmt.Errorf("color type must have exactly one flag set")
	}

	return nil
}

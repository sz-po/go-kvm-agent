package edid

import "fmt"

type EstablishedTimingsBlock [3]byte

type EstablishedTimingsSpecification struct {
	Supports720x400x70   bool `json:"720x400x70,omitempty"`
	Supports720x400x88   bool `json:"720x400x88,omitempty"`
	Supports640x480x60   bool `json:"640x480x60,omitempty"`
	Supports640x480x67   bool `json:"640x480x67,omitempty"`
	Supports640x480x72   bool `json:"640x480x72,omitempty"`
	Supports640x480x75   bool `json:"640x480x75,omitempty"`
	Supports800x600x56   bool `json:"800x600x56,omitempty"`
	Supports800x600x60   bool `json:"800x600x60,omitempty"`
	Supports800x600x72   bool `json:"800x600x72,omitempty"`
	Supports800x600x75   bool `json:"800x600x75,omitempty"`
	Supports832x624x75   bool `json:"832x624x75,omitempty"`
	Supports1024x768x87i bool `json:"1024x768x87i,omitempty"`
	Supports1024x768x60  bool `json:"1024x768x60,omitempty"`
	Supports1024x768x70  bool `json:"1024x768x70,omitempty"`
	Supports1024x768x75  bool `json:"1024x768x75,omitempty"`
	Supports1280x1024x75 bool `json:"1280x1024x75,omitempty"`
	Supports1152x870x75  bool `json:"1152x870x75,omitempty"`
}

const (
	establishedTimingsByte0ReservedMask byte = 0x00
	establishedTimingsByte1ReservedMask byte = 0x00
	establishedTimingsByte2AppleMask    byte = 0b10000000
	establishedTimingsByte2ReservedMask byte = 0b01111111
)

func CreateEstablishedTimingsSpecificationFromBlock(block EstablishedTimingsBlock) (*EstablishedTimingsSpecification, error) {
	if reserved := block[0] & establishedTimingsByte0ReservedMask; reserved != 0 {
		return nil, fmt.Errorf("reserved bits set in established timings byte 0: 0x%02x", reserved)
	}

	if reserved := block[1] & establishedTimingsByte1ReservedMask; reserved != 0 {
		return nil, fmt.Errorf("reserved bits set in established timings byte 1: 0x%02x", reserved)
	}

	if reserved := block[2] & establishedTimingsByte2ReservedMask; reserved != 0 {
		return nil, fmt.Errorf("reserved bits set in established timings byte 2: 0x%02x", reserved)
	}

	specification := &EstablishedTimingsSpecification{}

	specification.Supports720x400x70 = block[0]&0b10000000 != 0
	specification.Supports720x400x88 = block[0]&0b01000000 != 0
	specification.Supports640x480x60 = block[0]&0b00100000 != 0
	specification.Supports640x480x67 = block[0]&0b00010000 != 0
	specification.Supports640x480x72 = block[0]&0b00001000 != 0
	specification.Supports640x480x75 = block[0]&0b00000100 != 0
	specification.Supports800x600x56 = block[0]&0b00000010 != 0
	specification.Supports800x600x60 = block[0]&0b00000001 != 0

	specification.Supports800x600x72 = block[1]&0b10000000 != 0
	specification.Supports800x600x75 = block[1]&0b01000000 != 0
	specification.Supports832x624x75 = block[1]&0b00100000 != 0
	specification.Supports1024x768x87i = block[1]&0b00010000 != 0
	specification.Supports1024x768x60 = block[1]&0b00001000 != 0
	specification.Supports1024x768x70 = block[1]&0b00000100 != 0
	specification.Supports1024x768x75 = block[1]&0b00000010 != 0
	specification.Supports1280x1024x75 = block[1]&0b00000001 != 0

	specification.Supports1152x870x75 = block[2]&establishedTimingsByte2AppleMask != 0

	if err := specification.Validate(); err != nil {
		return nil, fmt.Errorf("validate specification: %w", err)
	}

	return specification, nil
}

func CreateEstablishedTimingsBlockFromSpecification(specification EstablishedTimingsSpecification) (*EstablishedTimingsBlock, error) {
	if err := specification.Validate(); err != nil {
		return nil, fmt.Errorf("validate specification: %w", err)
	}

	block := &EstablishedTimingsBlock{}

	if specification.Supports720x400x70 {
		block[0] |= 0b10000000
	}

	if specification.Supports720x400x88 {
		block[0] |= 0b01000000
	}

	if specification.Supports640x480x60 {
		block[0] |= 0b00100000
	}

	if specification.Supports640x480x67 {
		block[0] |= 0b00010000
	}

	if specification.Supports640x480x72 {
		block[0] |= 0b00001000
	}

	if specification.Supports640x480x75 {
		block[0] |= 0b00000100
	}

	if specification.Supports800x600x56 {
		block[0] |= 0b00000010
	}

	if specification.Supports800x600x60 {
		block[0] |= 0b00000001
	}

	if specification.Supports800x600x72 {
		block[1] |= 0b10000000
	}

	if specification.Supports800x600x75 {
		block[1] |= 0b01000000
	}

	if specification.Supports832x624x75 {
		block[1] |= 0b00100000
	}

	if specification.Supports1024x768x87i {
		block[1] |= 0b00010000
	}

	if specification.Supports1024x768x60 {
		block[1] |= 0b00001000
	}

	if specification.Supports1024x768x70 {
		block[1] |= 0b00000100
	}

	if specification.Supports1024x768x75 {
		block[1] |= 0b00000010
	}

	if specification.Supports1280x1024x75 {
		block[1] |= 0b00000001
	}

	if specification.Supports1152x870x75 {
		block[2] |= establishedTimingsByte2AppleMask
	}

	return block, nil
}

func (specification *EstablishedTimingsSpecification) Validate() error {
	if specification == nil {
		return fmt.Errorf("nil specification")
	}

	return nil
}

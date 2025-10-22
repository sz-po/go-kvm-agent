package edid

import (
	"fmt"
)

const VersionBlockDefault = "\x01\x03"

type VersionBlock [2]byte

func NewVersionBlock() (*VersionBlock, error) {
	block := &VersionBlock{}
	copy(block[0:2], VersionBlockDefault)

	if err := block.Validate(); err != nil {
		return nil, fmt.Errorf("invalid version block: %w", err)
	}

	return block, nil
}

func (block *VersionBlock) Validate() error {
	if string(block[:]) != VersionBlockDefault {
		return fmt.Errorf("invalid version block")
	}

	return nil
}

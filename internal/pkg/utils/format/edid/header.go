package edid

import "fmt"

const HeaderBlockDefault = "\x00\xff\xff\xff\xff\xff\xff\x00"

type HeaderBlock [8]byte

func NewHeaderBlock() (*HeaderBlock, error) {
	block := &HeaderBlock{}
	copy(block[0:8], HeaderBlockDefault)

	if err := block.Validate(); err != nil {
		return nil, fmt.Errorf("invalid header block: %w", err)
	}

	return block, nil
}

func (block *HeaderBlock) Validate() error {
	if string(block[:]) != HeaderBlockDefault {
		return fmt.Errorf("invalid header block")
	}

	return nil
}

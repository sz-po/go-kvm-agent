package edid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHeaderBlock(t *testing.T) {
	t.Parallel()

	headerBlock, err := NewHeaderBlock()

	assert.NoError(t, err)
	assert.Equal(t, HeaderBlockDefault, string(headerBlock[:]))
}

func TestHeaderBlockValidate(t *testing.T) {
	t.Parallel()

	headerBlock := &HeaderBlock{}
	copy(headerBlock[:], HeaderBlockDefault)

	assert.NoError(t, headerBlock.Validate())

	headerBlock[0] = 0x01
	assert.Error(t, headerBlock.Validate())
}

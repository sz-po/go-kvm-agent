package edid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVersionBlock(t *testing.T) {
	t.Parallel()

	versionBlock, err := NewVersionBlock()

	assert.NoError(t, err)
	assert.NotNil(t, versionBlock)
	assert.Equal(t, VersionBlockDefault, string(versionBlock[:]))
}

func TestVersionBlockValidate(t *testing.T) {
	t.Parallel()

	validBlock := &VersionBlock{}
	copy(validBlock[:], VersionBlockDefault)

	assert.NoError(t, validBlock.Validate())

	invalidBlock := &VersionBlock{}
	invalidBlock[0] = 0x00

	assert.Error(t, invalidBlock.Validate())
}

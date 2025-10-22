package ffmpeg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInputStdin(t *testing.T) {
	t.Parallel()

	input := NewInputStdin()

	assert.NotNil(t, input)
}

func TestInputStdinParameters(t *testing.T) {
	t.Parallel()

	input := NewInputStdin()

	parameters := input.Parameters()

	expected := []string{"-i", "pipe:0"}
	assert.Equal(t, expected, parameters)
}

func TestInputStdinImplementsInputInterface(t *testing.T) {
	t.Parallel()

	var _ Input = (*InputStdin)(nil)
}

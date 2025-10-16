package ffmpeg

type Output interface {
	Parameters() []string
}

type RawOutput []string

func (output RawOutput) Parameters() []string {
	return output
}

package ffmpeg

type Input interface {
	Parameters() []string
}

type RawInput []string

func (input RawInput) Parameters() []string {
	return input
}

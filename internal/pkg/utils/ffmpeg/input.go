package ffmpeg

type RawInput []string

func (input RawInput) Parameters() []string {
	return input
}

package ffmpeg

type RawOutput []string

func (output RawOutput) Parameters() []string {
	return output
}

package ffmpeg

type RawConfiguration []string

func (configuration RawConfiguration) Parameters() []string {
	return configuration
}

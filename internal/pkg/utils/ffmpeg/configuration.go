package ffmpeg

type Configuration interface {
	Parameters() []string
}

type RawConfiguration []string

func (configuration RawConfiguration) Parameters() []string {
	return configuration
}

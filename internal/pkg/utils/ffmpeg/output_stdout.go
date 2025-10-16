package ffmpeg

type OutputStdout struct{}

func NewOutputStdout() *OutputStdout {
	return &OutputStdout{}
}

func (output *OutputStdout) Parameters() []string {
	return []string{
		"pipe:1",
	}
}

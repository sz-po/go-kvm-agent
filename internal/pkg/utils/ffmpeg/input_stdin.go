package ffmpeg

type InputStdin struct {
}

func NewInputStdin() *InputStdin {
	return &InputStdin{}
}

func (input *InputStdin) Parameters() []string {
	return []string{
		"-i",
		"pipe:0",
	}
}

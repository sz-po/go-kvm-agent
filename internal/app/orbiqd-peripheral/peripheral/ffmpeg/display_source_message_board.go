package ffmpeg

import (
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/ffmpeg"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type DisplaySourceMessageBoardInputConfig struct {
	DisplayMode peripheralSDK.DisplayMode `json:"displayMode"`
	CenterText  *string                   `json:"centerText"`
}

type DisplaySourceMessageBoardInput struct {
	messageBoard *ffmpeg.InputMessageBoard
}

func NewDisplaySourceMessageBoardInput(config DisplaySourceMessageBoardInputConfig) *DisplaySourceMessageBoardInput {
	centerText := utils.DefaultNil(config.CenterText, "example display source")

	return &DisplaySourceMessageBoardInput{
		messageBoard: ffmpeg.NewInputMessageBoard(centerText, "ffmpeg-display-source", config.DisplayMode),
	}
}

func (input *DisplaySourceMessageBoardInput) Parameters() []string {
	parameters := input.messageBoard.Parameters()

	parameters = append([]string{"-re"}, parameters...)

	return parameters
}

func (input *DisplaySourceMessageBoardInput) GetDisplayMode() (peripheralSDK.DisplayMode, error) {
	return input.messageBoard.GetDisplayMode(), nil
}

func (input *DisplaySourceMessageBoardInput) GetPixelFormat() peripheralSDK.DisplayPixelFormat {
	return peripheralSDK.DisplayPixelFormatRGB24
}

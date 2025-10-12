package ffmpeg

import (
	"fmt"

	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type DisplaySourceTestPatternInputConfig struct {
	DisplayMode peripheralSDK.DisplayMode `json:"displayMode"`
}

type DisplaySourceTestPatternInput struct {
	displayMode peripheralSDK.DisplayMode
}

func NewDisplaySourceTestPatternInput(config DisplaySourceTestPatternInputConfig) *DisplaySourceTestPatternInput {
	return &DisplaySourceTestPatternInput{
		displayMode: config.DisplayMode,
	}
}

func (input *DisplaySourceTestPatternInput) Parameters() []string {
	return []string{
		"-re",
		"-f",
		"lavfi",
		"-i",
		fmt.Sprintf("testsrc2=size=%dx%d:rate=%d", input.displayMode.Width, input.displayMode.Height, input.displayMode.RefreshRate),
		"-pix_fmt",
		"rgb24",
	}
}

func (input *DisplaySourceTestPatternInput) GetCurrentDisplayMode() peripheralSDK.DisplayMode {
	return input.displayMode
}

func (input *DisplaySourceTestPatternInput) GetCurrentPixelFormat() peripheralSDK.DisplayPixelFormat {
	return peripheralSDK.DisplayPixelFormatRGB24
}

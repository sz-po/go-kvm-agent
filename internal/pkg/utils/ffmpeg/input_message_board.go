package ffmpeg

import (
	"fmt"
	"math"
	"strings"

	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

const referenceHeight = 1080

type InputMessageBoard struct {
	centerText  string
	bottomText  string
	displayMode peripheralSDK.DisplayMode
}

func NewInputMessageBoard(centerText string, bottomText string, displayMode peripheralSDK.DisplayMode) *InputMessageBoard {
	return &InputMessageBoard{
		centerText:  centerText,
		bottomText:  bottomText,
		displayMode: displayMode,
	}
}

func (input *InputMessageBoard) Parameters() []string {
	centerFontSize := input.scaledValue(72)
	infoFontSize := input.scaledValue(36)
	clockFontSize := input.scaledValue(36)
	bottomFontSize := input.scaledValue(48)
	outlineThickness := input.scaledValue(5)
	boxBorderWidth := input.scaledValue(10)
	margin := input.scaledValue(20)
	bottomMargin := input.scaledValue(50)

	lavfiFilter := []string{
		fmt.Sprintf("smptehdbars=size=%dx%d:rate=%d", input.displayMode.Width, input.displayMode.Height, input.displayMode.RefreshRate),
		fmt.Sprintf("drawbox=x=0:y=0:w=iw:h=ih:color=white:t=%d", outlineThickness),
		fmt.Sprintf("drawtext=text='%s':fontcolor=white:fontsize=%d:box=1:boxcolor=black@0.5:boxborderw=%d:x=(w-text_w)/2:y=(h-text_h)/2", input.centerText, centerFontSize, boxBorderWidth),
		fmt.Sprintf("drawtext=text='%%{frame_num}':fontcolor=white:fontsize=%d:box=1:boxcolor=black@0.5:boxborderw=%d:x=w-text_w-%d:y=%d", infoFontSize, boxBorderWidth, margin, margin),
		fmt.Sprintf("drawtext=text='%%{localtime\\:%%H\\:%%M\\:%%S}':fontcolor=white:fontsize=%d:box=1:boxcolor=black@0.5:boxborderw=%d:x=%d:y=%d", clockFontSize, boxBorderWidth, margin, margin),
		fmt.Sprintf("drawtext=text='%s':fontcolor=white:fontsize=%d:box=1:boxcolor=black@0.8:boxborderw=%d:x=(w-text_w)/2:y=h-text_h-%d", input.bottomText, bottomFontSize, boxBorderWidth, bottomMargin),
	}

	return []string{
		"-f",
		"lavfi",
		"-i",
		strings.Join(lavfiFilter, ","),
	}
}

func (input *InputMessageBoard) GetDisplayMode() peripheralSDK.DisplayMode {
	return input.displayMode
}

func (input *InputMessageBoard) scaledValue(baseValue int) int {
	scaleFactor := input.scalingFactor()
	scaledValue := int(math.Round(float64(baseValue) * scaleFactor))
	if scaledValue < 1 {
		return 1
	}
	return scaledValue
}

func (input *InputMessageBoard) scalingFactor() float64 {
	if input.displayMode.Height <= 0 {
		return 1
	}
	return float64(input.displayMode.Height) / float64(referenceHeight)
}

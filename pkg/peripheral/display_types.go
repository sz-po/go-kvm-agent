package peripheral

import (
	"errors"
	"fmt"
)

// DisplayPixelFormat defines the pixel format for display frames.
type DisplayPixelFormat string

const (
	// DisplayPixelFormatUnknown represents an uninitialized or invalid pixel format.
	DisplayPixelFormatUnknown DisplayPixelFormat = ""
	// DisplayPixelFormatRGB24 represents 24-bit RGB pixel format (8 bits per channel).
	DisplayPixelFormatRGB24 DisplayPixelFormat = "rgb24"
)

// BytesPerPixel returns the number of bytes per pixel for the format.
func (pixelFormat DisplayPixelFormat) BytesPerPixel() int {
	switch pixelFormat {
	case DisplayPixelFormatRGB24:
		return 3
	default:
		return 0
	}
}

// String returns the string representation of the pixel format.
func (pixelFormat DisplayPixelFormat) String() string {
	return string(pixelFormat)
}

// DisplayMode represents a single display mode configuration.
type DisplayMode struct {
	Width       uint32 `json:"width"`
	Height      uint32 `json:"height"`
	RefreshRate uint32 `json:"refreshRate"`
}

func (displayMode DisplayMode) String() string {
	return fmt.Sprintf("%dx%d@%d", displayMode.Width, displayMode.Height, displayMode.RefreshRate)
}

type DisplayModeList []DisplayMode

func (displayModeList DisplayModeList) Supports(testedMode DisplayMode) bool {
	for _, supportedMode := range displayModeList {
		if supportedMode.Width == testedMode.Width &&
			supportedMode.Height == testedMode.Height &&
			supportedMode.RefreshRate == testedMode.RefreshRate {

			return true
		}
	}

	return false
}

// DisplayInfo contains information about a display device.
type DisplayInfo struct {
	Manufacturer   string
	Model          string
	SerialNumber   string
	SupportedModes []DisplayMode
	CurrentMode    *DisplayMode
}

var ErrUnsupportedDisplayMode = errors.New("unsupported display mode")

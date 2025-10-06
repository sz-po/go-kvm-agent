package peripherals

// DisplayPixelFormat defines the pixel format for display frames.
type DisplayPixelFormat int

const (
	DisplayPixelFormatUnknown DisplayPixelFormat = iota
	DisplayPixelFormatRGB24
)

// BytesPerPixel returns the number of bytes per pixel for the format.
func (pf DisplayPixelFormat) BytesPerPixel() int {
	switch pf {
	case DisplayPixelFormatRGB24:
		return 3
	default:
		return 0
	}
}

// String returns the string representation of the pixel format.
func (pf DisplayPixelFormat) String() string {
	switch pf {
	case DisplayPixelFormatRGB24:
		return "RGB24"
	case DisplayPixelFormatUnknown:
		return "Unknown"
	default:
		return "Unknown"
	}
}

// DisplayMode represents a single display mode configuration.
type DisplayMode struct {
	Width       uint32
	Height      uint32
	RefreshRate uint32
}

// DisplayInfo contains information about a display device.
type DisplayInfo struct {
	Manufacturer   string
	Model          string
	SerialNumber   string
	SupportedModes []DisplayMode
	CurrentMode    DisplayMode
}

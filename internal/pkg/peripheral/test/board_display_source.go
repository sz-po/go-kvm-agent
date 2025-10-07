package test

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fogleman/gg"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

// BoardDisplaySourceConfig contains configuration for the board display source.
type BoardDisplaySourceConfig struct {
	Mode        peripheral.DisplayMode        `json:"mode"`
	ChunkSize   int                           `json:"chunkSize"`
	PixelFormat peripheral.DisplayPixelFormat `json:"pixelFormat"`
}

const BoardDisplaySourceDriver = peripheral.PeripheralDriver("board-display-source")

// DefaultChunkSize is the default size of frame data chunks in bytes.
// Set to 57600 bytes (10 rows for 1920x1080 RGB24), which is a multiple of many common row sizes.
const DefaultChunkSize = 57600

// DefaultCornerSize is the default size of corner markers in pixels.
const DefaultCornerSize = 5

// BorderWidth is the static width of the border in pixels.
const BorderWidth = 1

var (
	// ErrInvalidMode is returned when the display mode is invalid.
	ErrInvalidMode = errors.New("invalid display mode")
	// ErrInvalidChunkSize is returned when chunk size is invalid.
	ErrInvalidChunkSize = errors.New("chunk size must be greater than 0")
	// ErrUnsupportedPixelFormat is returned when pixel format is not supported.
	ErrUnsupportedPixelFormat = errors.New("unsupported pixel format")
	// ErrAlreadyStarted is returned when Start is called on an already running source.
	ErrAlreadyStarted = errors.New("display source already started")
	// ErrNotStarted is returned when Stop is called on a non-running source.
	ErrNotStarted = errors.New("display source not started")
)

// NewBoardDisplaySourceConfig creates a new config with defaults.
func NewBoardDisplaySourceConfig(mode peripheral.DisplayMode) BoardDisplaySourceConfig {
	return BoardDisplaySourceConfig{
		Mode:        mode,
		ChunkSize:   DefaultChunkSize,
		PixelFormat: peripheral.DisplayPixelFormatRGB24,
	}
}

// boardDisplaySourceRenderer renders test pattern frames using the gg library.
type boardDisplaySourceRenderer struct {
	width       int
	height      int
	refreshRate uint32
	logger      *slog.Logger
}

// newBoardDisplaySourceRenderer creates a new board display source renderer.
func newBoardDisplaySourceRenderer(config BoardDisplaySourceConfig, logger *slog.Logger) (*boardDisplaySourceRenderer, error) {
	// Validate config
	if config.Mode.Width == 0 || config.Mode.Height == 0 || config.Mode.RefreshRate == 0 {
		return nil, ErrInvalidMode
	}
	if config.ChunkSize <= 0 {
		return nil, ErrInvalidChunkSize
	}
	if config.PixelFormat != peripheral.DisplayPixelFormatRGB24 {
		return nil, ErrUnsupportedPixelFormat
	}

	// Validate that chunk size is a multiple of row size
	bytesPerRow := int(config.Mode.Width) * config.PixelFormat.BytesPerPixel()
	if config.ChunkSize%bytesPerRow != 0 {
		return nil, fmt.Errorf("chunk size (%d) must be a multiple of row size (%d bytes): %w", config.ChunkSize, bytesPerRow, ErrInvalidChunkSize)
	}

	return &boardDisplaySourceRenderer{
		width:       int(config.Mode.Width),
		height:      int(config.Mode.Height),
		refreshRate: config.Mode.RefreshRate,
		logger:      logger,
	}, nil
}

// renderFrame renders a complete test pattern frame.
func (renderer *boardDisplaySourceRenderer) renderFrame(frameID uint64, lastRenderTime time.Duration) []byte {
	dc := gg.NewContext(renderer.width, renderer.height)

	// 1. Render background pattern
	renderer.renderBackground(dc)

	// 2. Render white border
	renderer.renderBorder(dc)

	// 3. Render colored corners
	renderer.renderCorners(dc)

	// 4. Render counter box
	renderer.renderCounter(dc, frameID, lastRenderTime, time.Now())

	// Convert to RGB24
	return renderer.imageToRGB24(dc.Image())
}

// renderBackground renders all test patterns in horizontal sections.
func (renderer *boardDisplaySourceRenderer) renderBackground(dc *gg.Context) {
	h := float64(renderer.height)
	w := float64(renderer.width)

	// Divide screen into 3 horizontal sections
	sectionHeight := h / 3.0

	// Top section: Horizontal gradient
	renderer.renderHorizontalGradient(dc, 0, 0, w, sectionHeight)

	// Middle section: Vertical bars
	renderer.renderVerticalBars(dc, 0, sectionHeight, w, sectionHeight)

	// Bottom section: Checkerboard
	renderer.renderCheckerboard(dc, 0, sectionHeight*2, w, sectionHeight)
}

// renderHorizontalGradient renders a smooth horizontal RGB gradient in the specified bounds.
func (renderer *boardDisplaySourceRenderer) renderHorizontalGradient(dc *gg.Context, x, y, width, height float64) {
	// Create a gradient that cycles through R -> G -> B -> R
	for ix := 0; ix < int(width); ix++ {
		t := float64(ix) / width
		var r, g, b float64

		if t < 0.33 {
			// R -> G
			p := t / 0.33
			r = 1.0 - p
			g = p
			b = 0.0
		} else if t < 0.67 {
			// G -> B
			p := (t - 0.33) / 0.34
			r = 0.0
			g = 1.0 - p
			b = p
		} else {
			// B -> R
			p := (t - 0.67) / 0.33
			r = p
			g = 0.0
			b = 1.0 - p
		}

		dc.SetRGB(r, g, b)
		dc.DrawRectangle(x+float64(ix), y, 1, height)
		dc.Fill()
	}
}

// renderVerticalBars renders 8 vertical color bars (classic test pattern) in the specified bounds.
func (renderer *boardDisplaySourceRenderer) renderVerticalBars(dc *gg.Context, x, y, width, height float64) {
	bars := []color.RGBA{
		{255, 255, 255, 255}, // White
		{255, 255, 0, 255},   // Yellow
		{0, 255, 255, 255},   // Cyan
		{0, 255, 0, 255},     // Green
		{255, 0, 255, 255},   // Magenta
		{255, 0, 0, 255},     // Red
		{0, 0, 255, 255},     // Blue
		{0, 0, 0, 255},       // Black
	}

	barWidth := width / float64(len(bars))

	for i, col := range bars {
		barX := x + float64(i)*barWidth
		dc.SetColor(col)
		dc.DrawRectangle(barX, y, barWidth, height)
		dc.Fill()
	}
}

// renderCheckerboard renders a checkerboard pattern in the specified bounds.
func (renderer *boardDisplaySourceRenderer) renderCheckerboard(dc *gg.Context, x, y, width, height float64) {
	squareSize := 32.0 // pixels
	cols := int((width + squareSize - 1) / squareSize)
	rows := int((height + squareSize - 1) / squareSize)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			// Alternate between black and white
			if (row+col)%2 == 0 {
				dc.SetRGB(1, 1, 1) // White
			} else {
				dc.SetRGB(0, 0, 0) // Black
			}

			squareX := x + float64(col)*squareSize
			squareY := y + float64(row)*squareSize
			squareW := squareSize
			squareH := squareSize

			// Clip to bounds
			if squareX+squareW > x+width {
				squareW = x + width - squareX
			}
			if squareY+squareH > y+height {
				squareH = y + height - squareY
			}

			dc.DrawRectangle(squareX, squareY, squareW, squareH)
			dc.Fill()
		}
	}
}

// renderBorder renders a 1px white border around the frame.
func (renderer *boardDisplaySourceRenderer) renderBorder(dc *gg.Context) {
	dc.SetRGB(1, 1, 1) // White
	dc.SetLineWidth(BorderWidth)

	// Draw rectangle outline
	offset := float64(BorderWidth) / 2.0
	dc.DrawRectangle(
		offset,
		offset,
		float64(renderer.width)-BorderWidth,
		float64(renderer.height)-BorderWidth,
	)
	dc.Stroke()
}

// renderCorners renders colored corner markers (R/G/B/Y).
func (renderer *boardDisplaySourceRenderer) renderCorners(dc *gg.Context) {
	size := float64(DefaultCornerSize)
	w := float64(renderer.width)
	h := float64(renderer.height)

	corners := []struct {
		x, y    float64
		r, g, b float64
	}{
		{0, 0, 1, 0, 0},               // Top-left: Red
		{w - size, 0, 0, 1, 0},        // Top-right: Green
		{0, h - size, 0, 0, 1},        // Bottom-left: Blue
		{w - size, h - size, 1, 1, 0}, // Bottom-right: Yellow
	}

	for _, corner := range corners {
		dc.SetRGB(corner.r, corner.g, corner.b)
		dc.DrawRectangle(corner.x, corner.y, size, size)
		dc.Fill()
	}
}

// renderCounter renders the central information box.
func (renderer *boardDisplaySourceRenderer) renderCounter(dc *gg.Context, frameID uint64, renderTime time.Duration, now time.Time) {
	// Counter box dimensions
	boxWidth := 300.0
	boxHeight := 100.0
	boxX := float64(renderer.width)/2.0 - boxWidth/2.0
	boxY := float64(renderer.height)/2.0 - boxHeight/2.0

	// Draw black background box
	dc.SetRGB(0, 0, 0)
	dc.DrawRectangle(boxX, boxY, boxWidth, boxHeight)
	dc.Fill()

	// Draw white border around box
	dc.SetRGB(1, 1, 1)
	dc.SetLineWidth(2)
	dc.DrawRectangle(boxX, boxY, boxWidth, boxHeight)
	dc.Stroke()

	// Prepare text
	dc.SetRGB(1, 1, 1) // White text
	fontSize := 14.0
	if err := dc.LoadFontFace("/System/Library/Fonts/Courier.dfont", fontSize); err != nil {
		// Fallback: try to load a common monospace font
		// If this fails, gg will use a default font
		renderer.logger.Warn("Failed to load font, using default.", slog.String("error", err.Error()))
	}

	// Format text lines
	lines := []string{
		fmt.Sprintf("Frame: %08d", frameID),
		fmt.Sprintf("Render: %.3fms", float64(renderTime.Microseconds())/1000.0),
		fmt.Sprintf("Time: %s", now.Format("15:04:05.000")),
		fmt.Sprintf("%dx%d @ %dHz", renderer.width, renderer.height, renderer.refreshRate),
	}

	// Draw lines centered in box
	lineHeight := 20.0
	startY := boxY + boxHeight/2.0 - float64(len(lines))*lineHeight/2.0

	for i, line := range lines {
		y := startY + float64(i)*lineHeight + lineHeight/2.0
		dc.DrawStringAnchored(line, float64(renderer.width)/2.0, y, 0.5, 0.5)
	}
}

// imageToRGB24 converts an image.Image to RGB24 byte array.
func (renderer *boardDisplaySourceRenderer) imageToRGB24(img image.Image) []byte {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	rgb := make([]byte, width*height*3)

	idx := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Convert from 16-bit to 8-bit color
			rgb[idx] = byte(r >> 8)
			rgb[idx+1] = byte(g >> 8)
			rgb[idx+2] = byte(b >> 8)
			idx += 3
		}
	}
	return rgb
}

// BoardDisplaySource is a test implementation of DisplaySource that generates test board patterns.
type BoardDisplaySource struct {
	id       peripheral.PeripheralID
	config   BoardDisplaySourceConfig
	renderer *boardDisplaySourceRenderer
	logger   *slog.Logger

	// Channels
	dataChannel    chan peripheral.DisplayEvent
	controlChannel chan peripheral.DisplayControlEvent

	// Lifecycle
	running atomic.Bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	// Metrics
	lastRenderTime time.Duration
	frameInterval  time.Duration
}

// NewBoardDisplaySource creates a new board display source.
func NewBoardDisplaySource(config BoardDisplaySourceConfig) (*BoardDisplaySource, error) {
	logger := slog.Default()

	renderer, err := newBoardDisplaySourceRenderer(config, logger)
	if err != nil {
		return nil, err
	}

	id, err := peripheral.NewPeripheralID(
		peripheral.PeripheralTypeDisplay,
		peripheral.PeripheralRoleSource,
		"test-pattern",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create peripheral ID: %w", err)
	}

	return &BoardDisplaySource{
		id:             id,
		config:         config,
		renderer:       renderer,
		logger:         logger.With(slog.String("peripheral", id.String())),
		dataChannel:    make(chan peripheral.DisplayEvent, 100),
		controlChannel: make(chan peripheral.DisplayControlEvent, 10),
		frameInterval:  time.Second / time.Duration(config.Mode.RefreshRate),
	}, nil
}

// ID returns the peripheral ID.
func (source *BoardDisplaySource) ID() peripheral.PeripheralID {
	return source.id
}

// DataChannel returns the data event channel.
func (source *BoardDisplaySource) DataChannel(ctx context.Context) <-chan peripheral.DisplayEvent {
	return source.dataChannel
}

// ControlChannel returns the control event channel.
func (source *BoardDisplaySource) ControlChannel(ctx context.Context) <-chan peripheral.DisplayControlEvent {
	return source.controlChannel
}

// GetCurrentDisplayMode returns the current display mode.
func (source *BoardDisplaySource) GetCurrentDisplayMode() (*peripheral.DisplayMode, error) {
	mode := source.config.Mode
	return &mode, nil
}

// Start starts the board display source.
func (source *BoardDisplaySource) Start(ctx context.Context, info peripheral.DisplayInfo) error {
	if source.running.Load() {
		return ErrAlreadyStarted
	}

	source.ctx, source.cancel = context.WithCancel(ctx)
	source.running.Store(true)

	source.logger.Info("Starting board display source.")

	// Start frame generation goroutine
	source.wg.Add(1)
	go func() {
		defer source.wg.Done()
		source.generateFrames()
	}()

	// Send started event
	source.controlChannel <- peripheral.NewDisplaySourceStartedEvent(time.Now())

	return nil
}

// Stop stops the board display source.
func (source *BoardDisplaySource) Stop(ctx context.Context) error {
	if !source.running.Load() {
		return ErrNotStarted
	}

	source.logger.Info("Stopping board display source.")

	source.cancel()
	source.wg.Wait()
	source.running.Store(false)

	// Send stopped event
	source.controlChannel <- peripheral.NewDisplaySourceStoppedEvent(time.Now())

	return nil
}

// generateFrames generates and sends test pattern frames at the configured rate.
func (source *BoardDisplaySource) generateFrames() {
	ticker := time.NewTicker(source.frameInterval)
	defer ticker.Stop()

	frameID := uint64(0)

	for {
		select {
		case <-source.ctx.Done():
			source.logger.Info("Frame generation stopped.")
			return
		case <-ticker.C:
			start := time.Now()

			// Render frame
			frameData := source.renderer.renderFrame(frameID, source.lastRenderTime)

			// Measure render time
			source.lastRenderTime = time.Since(start)

			// Send frame in chunks
			source.sendFrameInChunks(frameID, frameData)

			frameID++
		}
	}
}

// sendFrameInChunks sends a frame as a sequence of events.
func (source *BoardDisplaySource) sendFrameInChunks(frameID uint64, frameData []byte) {
	now := time.Now()

	// Send FrameStart event
	source.dataChannel <- peripheral.NewDisplayFrameStartEvent(
		frameID,
		source.config.Mode.Width,
		source.config.Mode.Height,
		source.config.PixelFormat,
		now,
	)

	// Send frame data in chunks
	chunkIndex := uint32(0)
	for i := 0; i < len(frameData); i += source.config.ChunkSize {
		end := i + source.config.ChunkSize
		if end > len(frameData) {
			end = len(frameData)
		}

		chunk := frameData[i:end]

		source.dataChannel <- peripheral.NewDisplayFrameChunkEvent(
			frameID,
			chunkIndex,
			chunk,
			time.Now(),
		)
		chunkIndex++
	}

	// Send FrameEnd event
	source.dataChannel <- peripheral.NewDisplayFrameEndEvent(
		frameID,
		chunkIndex,
		time.Now(),
	)
}

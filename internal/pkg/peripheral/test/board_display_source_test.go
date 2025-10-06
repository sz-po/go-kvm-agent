package test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

func TestNewBoardDisplaySourceConfig(t *testing.T) {
	mode := peripheral.DisplayMode{
		Width:       1920,
		Height:      1080,
		RefreshRate: 60,
	}

	config := NewBoardDisplaySourceConfig(mode)

	assert.Equal(t, mode, config.Mode)
	assert.Equal(t, DefaultChunkSize, config.ChunkSize)
	assert.Equal(t, peripheral.DisplayPixelFormatRGB24, config.PixelFormat)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      BoardDisplaySourceConfig
		expectError error
	}{
		{
			name: "valid config",
			config: BoardDisplaySourceConfig{
				Mode:        peripheral.DisplayMode{Width: 1920, Height: 1080, RefreshRate: 60},
				ChunkSize:   DefaultChunkSize,
				PixelFormat: peripheral.DisplayPixelFormatRGB24,
			},
			expectError: nil,
		},
		{
			name: "invalid mode - zero width",
			config: BoardDisplaySourceConfig{
				Mode:        peripheral.DisplayMode{Width: 0, Height: 1080, RefreshRate: 60},
				ChunkSize:   DefaultChunkSize,
				PixelFormat: peripheral.DisplayPixelFormatRGB24,
			},
			expectError: ErrInvalidMode,
		},
		{
			name: "invalid mode - zero height",
			config: BoardDisplaySourceConfig{
				Mode:        peripheral.DisplayMode{Width: 1920, Height: 0, RefreshRate: 60},
				ChunkSize:   DefaultChunkSize,
				PixelFormat: peripheral.DisplayPixelFormatRGB24,
			},
			expectError: ErrInvalidMode,
		},
		{
			name: "invalid mode - zero refresh rate",
			config: BoardDisplaySourceConfig{
				Mode:        peripheral.DisplayMode{Width: 1920, Height: 1080, RefreshRate: 0},
				ChunkSize:   DefaultChunkSize,
				PixelFormat: peripheral.DisplayPixelFormatRGB24,
			},
			expectError: ErrInvalidMode,
		},
		{
			name: "invalid chunk size",
			config: BoardDisplaySourceConfig{
				Mode:        peripheral.DisplayMode{Width: 1920, Height: 1080, RefreshRate: 60},
				ChunkSize:   0,
				PixelFormat: peripheral.DisplayPixelFormatRGB24,
			},
			expectError: ErrInvalidChunkSize,
		},
		{
			name: "unsupported pixel format",
			config: BoardDisplaySourceConfig{
				Mode:        peripheral.DisplayMode{Width: 1920, Height: 1080, RefreshRate: 60},
				ChunkSize:   DefaultChunkSize,
				PixelFormat: peripheral.DisplayPixelFormatUnknown,
			},
			expectError: ErrUnsupportedPixelFormat,
		},
		{
			name: "chunk size not multiple of row size",
			config: BoardDisplaySourceConfig{
				Mode:        peripheral.DisplayMode{Width: 1920, Height: 1080, RefreshRate: 60},
				ChunkSize:   5000, // Not a multiple of 5760 (1920*3)
				PixelFormat: peripheral.DisplayPixelFormatRGB24,
			},
			expectError: ErrInvalidChunkSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newBoardDisplaySourceRenderer(tt.config, slog.Default())
			if tt.expectError != nil {
				assert.ErrorIs(t, err, tt.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewBoardDisplaySource(t *testing.T) {
	mode := peripheral.DisplayMode{
		Width:       1920,
		Height:      1080,
		RefreshRate: 60,
	}
	config := NewBoardDisplaySourceConfig(mode)

	source, err := NewBoardDisplaySource(config)
	require.NoError(t, err)
	require.NotNil(t, source)

	assert.Equal(t, peripheral.PeripheralTypeDisplay, source.ID().Type())
	assert.Equal(t, peripheral.PeripheralRoleSource, source.ID().Role())
	assert.Equal(t, config, source.config)
}

func TestNewBoardDisplaySourceInvalidConfig(t *testing.T) {
	config := BoardDisplaySourceConfig{
		Mode:        peripheral.DisplayMode{Width: 0, Height: 0, RefreshRate: 0},
		ChunkSize:   DefaultChunkSize,
		PixelFormat: peripheral.DisplayPixelFormatRGB24,
	}

	source, err := NewBoardDisplaySource(config)
	assert.Error(t, err)
	assert.Nil(t, source)
}

func TestStartStop(t *testing.T) {
	mode := peripheral.DisplayMode{
		Width:       640,
		Height:      480,
		RefreshRate: 30,
	}
	config := NewBoardDisplaySourceConfig(mode)

	source, err := NewBoardDisplaySource(config)
	require.NoError(t, err)

	ctx := context.Background()
	info := peripheral.DisplayInfo{
		Manufacturer:   "Test",
		Model:          "TestSource",
		SerialNumber:   "12345",
		SupportedModes: []peripheral.DisplayMode{mode},
		CurrentMode:    mode,
	}

	// Start
	err = source.Start(ctx, info)
	assert.NoError(t, err)
	assert.True(t, source.running.Load())

	// Try starting again (should fail)
	err = source.Start(ctx, info)
	assert.ErrorIs(t, err, ErrAlreadyStarted)

	// Stop
	err = source.Stop(ctx)
	assert.NoError(t, err)
	assert.False(t, source.running.Load())

	// Try stopping again (should fail)
	err = source.Stop(ctx)
	assert.ErrorIs(t, err, ErrNotStarted)
}

func TestFrameGeneration(t *testing.T) {
	mode := peripheral.DisplayMode{
		Width:       320,
		Height:      240,
		RefreshRate: 10, // Low FPS for faster testing
	}
	config := NewBoardDisplaySourceConfig(mode)
	config.ChunkSize = 1920 // 2 rows for 320x240 RGB24 (320*3*2 = 1920)

	source, err := NewBoardDisplaySource(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	info := peripheral.DisplayInfo{CurrentMode: mode}

	// Start source
	err = source.Start(ctx, info)
	require.NoError(t, err)
	defer source.Stop(context.Background())

	// Collect events
	dataCtx := context.Background()
	dataChan := source.DataChannel(dataCtx)
	controlChan := source.ControlChannel(dataCtx)

	// Wait for source started event
	select {
	case event := <-controlChan:
		assert.Equal(t, peripheral.DisplayControlSourceStarted, event.Type())
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for source started event")
	}

	// Wait for frame events
	frameStartSeen := false
	frameChunksSeen := 0
	frameEndSeen := false

	timeout := time.After(200 * time.Millisecond)
	for !frameEndSeen {
		select {
		case event := <-dataChan:
			switch event.Type() {
			case peripheral.DisplayEventFrameStart:
				frameStartSeen = true
				startEvent := event.(peripheral.DisplayFrameStartEvent)
				assert.Equal(t, mode.Width, startEvent.Width)
				assert.Equal(t, mode.Height, startEvent.Height)
				assert.Equal(t, peripheral.DisplayPixelFormatRGB24, startEvent.Format)
			case peripheral.DisplayEventFrameChunk:
				frameChunksSeen++
				chunkEvent := event.(peripheral.DisplayFrameChunkEvent)
				assert.NotEmpty(t, chunkEvent.Data)
			case peripheral.DisplayEventFrameEnd:
				frameEndSeen = true
				endEvent := event.(peripheral.DisplayFrameEndEvent)
				assert.Equal(t, uint32(frameChunksSeen), endEvent.TotalChunks)
			}
		case <-timeout:
			t.Fatal("timeout waiting for frame events")
		}
	}

	assert.True(t, frameStartSeen, "frame start event not seen")
	assert.Greater(t, frameChunksSeen, 0, "no frame chunks seen")
	assert.True(t, frameEndSeen, "frame end event not seen")
}

func TestFrameChunking(t *testing.T) {
	mode := peripheral.DisplayMode{
		Width:       100,
		Height:      100,
		RefreshRate: 1,
	}
	config := NewBoardDisplaySourceConfig(mode)
	config.ChunkSize = 6000 // 20 rows for 100x100 RGB24 (100*3*20 = 6000)

	source, err := NewBoardDisplaySource(config)
	require.NoError(t, err)

	// Calculate expected frame size and chunks
	bytesPerPixel := config.PixelFormat.BytesPerPixel()
	frameSize := int(mode.Width) * int(mode.Height) * bytesPerPixel
	expectedChunks := (frameSize + config.ChunkSize - 1) / config.ChunkSize

	// Generate one frame manually
	frameData := source.renderer.renderFrame(0, 0)
	assert.Equal(t, frameSize, len(frameData))

	// Test chunking
	ctx := context.Background()
	info := peripheral.DisplayInfo{CurrentMode: mode}

	err = source.Start(ctx, info)
	require.NoError(t, err)
	defer source.Stop(ctx)

	// Skip control event
	<-source.ControlChannel(ctx)

	// Read frame events
	dataChan := source.DataChannel(ctx)

	// FrameStart
	event := <-dataChan
	assert.Equal(t, peripheral.DisplayEventFrameStart, event.Type())

	// Count chunks
	chunkCount := 0
	totalBytes := 0
	for {
		event := <-dataChan
		if event.Type() == peripheral.DisplayEventFrameEnd {
			break
		}
		assert.Equal(t, peripheral.DisplayEventFrameChunk, event.Type())
		chunk := event.(peripheral.DisplayFrameChunkEvent)
		chunkCount++
		totalBytes += len(chunk.Data)
	}

	assert.Equal(t, expectedChunks, chunkCount)
	assert.Equal(t, frameSize, totalBytes)
}

func TestRenderFrame(t *testing.T) {
	mode := peripheral.DisplayMode{
		Width:       320,
		Height:      240,
		RefreshRate: 1,
	}

	config := NewBoardDisplaySourceConfig(mode)
	source, err := NewBoardDisplaySource(config)
	require.NoError(t, err)
	require.NotNil(t, source)

	// Render one frame to ensure no panics
	frameData := source.renderer.renderFrame(0, 0)
	expectedSize := int(mode.Width) * int(mode.Height) * 3 // RGB24
	assert.Equal(t, expectedSize, len(frameData))
}

func TestGetCurrentDisplayMode(t *testing.T) {
	mode := peripheral.DisplayMode{
		Width:       1920,
		Height:      1080,
		RefreshRate: 60,
	}
	config := NewBoardDisplaySourceConfig(mode)

	source, err := NewBoardDisplaySource(config)
	require.NoError(t, err)

	currentMode, err := source.GetCurrentDisplayMode()
	assert.NoError(t, err)
	if assert.NotNil(t, currentMode) {
		assert.Equal(t, mode, *currentMode)
	}
}

func BenchmarkRenderFrame(b *testing.B) {
	mode := peripheral.DisplayMode{
		Width:       1920,
		Height:      1080,
		RefreshRate: 60,
	}
	config := NewBoardDisplaySourceConfig(mode)
	renderer, err := newBoardDisplaySourceRenderer(config, slog.Default())
	if err != nil {
		b.Fatalf("Failed to create renderer: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = renderer.renderFrame(uint64(i), time.Millisecond)
	}
}

func BenchmarkRenderFrame720p(b *testing.B) {
	mode := peripheral.DisplayMode{
		Width:       1280,
		Height:      720,
		RefreshRate: 60,
	}
	config := NewBoardDisplaySourceConfig(mode)
	renderer, err := newBoardDisplaySourceRenderer(config, slog.Default())
	if err != nil {
		b.Fatalf("Failed to create renderer: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = renderer.renderFrame(uint64(i), time.Millisecond)
	}
}

func BenchmarkRenderFrame480p(b *testing.B) {
	mode := peripheral.DisplayMode{
		Width:       640,
		Height:      480,
		RefreshRate: 60,
	}
	config := NewBoardDisplaySourceConfig(mode)
	renderer, err := newBoardDisplaySourceRenderer(config, slog.Default())
	if err != nil {
		b.Fatalf("Failed to create renderer: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = renderer.renderFrame(uint64(i), time.Millisecond)
	}
}

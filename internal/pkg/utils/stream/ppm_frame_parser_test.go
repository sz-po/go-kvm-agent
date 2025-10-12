package stream

import (
	"bytes"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

func TestPPMFrameParser_SingleFrameSingleChunk(t *testing.T) {
	width := 2
	height := 2
	pixelCount := width * height
	pixelBytes := pixelCount * 3

	header := fmt.Sprintf("P6\n%d %d\n255\n", width, height)
	pixels := bytes.Repeat([]byte{0x10, 0x20, 0x30}, pixelCount)
	ppmData := append([]byte(header), pixels...)

	parser, err := NewPPMFrameParser(1024)
	assert.NoError(t, err)

	events, err := parser.Ingest(ppmData)
	assert.NoError(t, err)

	assert.Len(t, events, 3)

	startEvent, ok := events[0].(peripheralSDK.DisplayFrameStartEvent)
	assert.True(t, ok)
	assert.Equal(t, uint32(width), startEvent.Width)
	assert.Equal(t, uint32(height), startEvent.Height)
	assert.Equal(t, peripheralSDK.DisplayPixelFormatRGB24, startEvent.Format)
	assert.NotZero(t, startEvent.Timestamp())

	chunkEvent, ok := events[1].(peripheralSDK.DisplayFrameChunkEvent)
	assert.True(t, ok)
	assert.Equal(t, uint64(1), chunkEvent.FrameID)
	assert.Equal(t, uint32(0), chunkEvent.ChunkIndex)
	assert.Len(t, chunkEvent.Data, pixelBytes)
	assert.Equal(t, pixels, chunkEvent.Data)
	assert.NotZero(t, chunkEvent.Timestamp())

	endEvent, ok := events[2].(peripheralSDK.DisplayFrameEndEvent)
	assert.True(t, ok)
	assert.Equal(t, uint64(1), endEvent.FrameID)
	assert.Equal(t, uint32(1), endEvent.TotalChunks)
	assert.NotZero(t, endEvent.Timestamp())
}

func TestPPMFrameParser_MultiChunkFrame(t *testing.T) {
	width := 3
	height := 2
	header := fmt.Sprintf("P6\n%d %d\n255\n", width, height)
	pixels := bytes.Repeat([]byte{0xAA}, width*height*3)
	ppmData := append([]byte(header), pixels...)

	parser, err := NewPPMFrameParser(5)
	assert.NoError(t, err)

	events, err := parser.Ingest(ppmData)
	assert.NoError(t, err)

	expectedChunks := (len(pixels) + 5 - 1) / 5
	assert.Len(t, events, expectedChunks+2)

	chunkCount := 0
	for _, event := range events {
		if chunkEvent, ok := event.(peripheralSDK.DisplayFrameChunkEvent); ok {
			assert.Equal(t, uint64(1), chunkEvent.FrameID)
			if chunkCount < expectedChunks-1 {
				assert.Len(t, chunkEvent.Data, 5)
			}
			chunkCount++
		}
	}
	assert.Equal(t, expectedChunks, chunkCount)

	endEvent := events[len(events)-1].(peripheralSDK.DisplayFrameEndEvent)
	assert.Equal(t, uint32(expectedChunks), endEvent.TotalChunks)
}

func TestPPMFrameParser_HeaderWithCommentsAndPartialInput(t *testing.T) {
	parser, err := NewPPMFrameParser(1024)
	assert.NoError(t, err)

	part1 := []byte("P6\n# comment line one\n4")
	part2 := []byte(" 4\n# another comment\n255\n")
	pixels := bytes.Repeat([]byte{0x01, 0x02, 0x03}, 16)

	events, err := parser.Ingest(part1)
	assert.NoError(t, err)
	assert.Empty(t, events)

	events, err = parser.Ingest(part2)
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	_, isStart := events[0].(peripheralSDK.DisplayFrameStartEvent)
	assert.True(t, isStart)

	events, err = parser.Ingest(pixels)
	assert.NoError(t, err)
	assert.Len(t, events, 2)
	_, isChunk := events[0].(peripheralSDK.DisplayFrameChunkEvent)
	assert.True(t, isChunk)
	_, isEnd := events[1].(peripheralSDK.DisplayFrameEndEvent)
	assert.True(t, isEnd)
}

func TestPPMFrameParser_ReadWidthWaitsForCompleteToken(t *testing.T) {
	parser, err := NewPPMFrameParser(1024)
	assert.NoError(t, err)

	events, err := parser.Ingest([]byte("P6\n"))
	assert.NoError(t, err)
	assert.Empty(t, events)

	events, err = parser.Ingest([]byte("1"))
	assert.NoError(t, err)
	assert.Empty(t, events)

	headerRemainder := []byte(" 1\n255\n")
	events, err = parser.Ingest(headerRemainder)
	assert.NoError(t, err)
	assert.Len(t, events, 1)

	pixels := []byte{0x01, 0x02, 0x03}
	events, err = parser.Ingest(pixels)
	assert.NoError(t, err)
	assert.Len(t, events, 2)
}

func TestPPMFrameParser_ReadWidthRejectsInvalidNumber(t *testing.T) {
	parser, err := NewPPMFrameParser(1024)
	assert.NoError(t, err)

	events, err := parser.Ingest([]byte("P6\n0 1\n"))
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPPMInvalidDimension)
	assert.Empty(t, events)
}

func TestPPMFrameParser_ReadHeightWaitsForCompleteToken(t *testing.T) {
	parser, err := NewPPMFrameParser(1024)
	assert.NoError(t, err)

	initialHeader := []byte("P6\n1 ")
	events, err := parser.Ingest(initialHeader)
	assert.NoError(t, err)
	assert.Empty(t, events)

	// Supply an incomplete height token so the parser must wait for more data.
	events, err = parser.Ingest([]byte("1"))
	assert.NoError(t, err)
	assert.Empty(t, events)

	pixels := []byte{0xAA, 0xBB, 0xCC}
	events, err = parser.Ingest(append([]byte("\n255\n"), pixels...))
	assert.NoError(t, err)
	assert.Len(t, events, 3)

	frameStart, ok := events[0].(peripheralSDK.DisplayFrameStartEvent)
	assert.True(t, ok)
	assert.Equal(t, uint32(1), frameStart.Width)
	assert.Equal(t, uint32(1), frameStart.Height)
}

func TestPPMFrameParser_ReadHeightRejectsInvalidNumber(t *testing.T) {
	parser, err := NewPPMFrameParser(1024)
	assert.NoError(t, err)

	events, err := parser.Ingest([]byte("P6\n2 0\n"))
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPPMInvalidDimension)
	assert.Empty(t, events)
}

func TestPPMFrameParser_ReadMaxValueWaitsForCompleteToken(t *testing.T) {
	parser, err := NewPPMFrameParser(1024)
	assert.NoError(t, err)

	events, err := parser.Ingest([]byte("P6\n1 1\n"))
	assert.NoError(t, err)
	assert.Empty(t, events)

	events, err = parser.Ingest([]byte("25"))
	assert.NoError(t, err)
	assert.Empty(t, events)

	events, err = parser.Ingest([]byte("5\n"))
	assert.NoError(t, err)
	assert.Len(t, events, 1)

	pixels := []byte{0xAA, 0xBB, 0xCC}
	events, err = parser.Ingest(pixels)
	assert.NoError(t, err)
	assert.Len(t, events, 2)
}

func TestPPMFrameParser_ReadMaxValueRejectsInvalidNumber(t *testing.T) {
	parser, err := NewPPMFrameParser(1024)
	assert.NoError(t, err)

	events, err := parser.Ingest([]byte("P6\n1 1\n254\n"))
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPPMInvalidMaxValue)
	assert.Empty(t, events)
}

func TestPPMFrameParser_ReadMaxValueDetectsPayloadOverflow(t *testing.T) {
	parser, err := NewPPMFrameParser(1024)
	assert.NoError(t, err)

	largeDimension := uint64(math.MaxUint32)
	header := fmt.Sprintf("P6\n%d %d\n255\n", largeDimension, largeDimension)
	events, err := parser.Ingest([]byte(header))
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPPMInvalidDimension)
	assert.Empty(t, events)
}

func TestPPMFrameParser_ReadMaxValueRejectsPayloadExceedingHostInt(t *testing.T) {
	parser, err := NewPPMFrameParser(1024)
	assert.NoError(t, err)

	originalHostMaxInt := hostMaxInt
	hostMaxInt = 5
	defer func() {
		hostMaxInt = originalHostMaxInt
	}()

	events, err := parser.Ingest([]byte("P6\n1 2\n255\n"))
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPPMInvalidDimension)
	assert.Empty(t, events)
}

func TestPPMFrameParser_DoesNotConsumeBeyondCurrentPayload(t *testing.T) {
	parser, err := NewPPMFrameParser(1024)
	assert.NoError(t, err)

	frameOneHeader := []byte("P6\n1 1\n255\n")
	frameOnePixels := []byte{0x01, 0x02, 0x03}
	frameTwoHeader := []byte("P6\n1 1\n255\n")
	frameTwoPixels := []byte{0x04, 0x05, 0x06}

	combined := append(append(frameOneHeader, frameOnePixels...), append(frameTwoHeader, frameTwoPixels...)...)

	events, err := parser.Ingest(combined)
	assert.NoError(t, err)
	assert.Len(t, events, 6)

	firstEnd, ok := events[2].(peripheralSDK.DisplayFrameEndEvent)
	assert.True(t, ok)
	assert.Equal(t, uint64(1), firstEnd.FrameID)
	assert.Equal(t, uint32(1), firstEnd.TotalChunks)

	secondStart, ok := events[3].(peripheralSDK.DisplayFrameStartEvent)
	assert.True(t, ok)
	assert.Equal(t, uint64(2), secondStart.FrameID)
}

func TestPPMFrameParser_ReadPayloadDetectsAccountingMismatch(t *testing.T) {
	parser, err := NewPPMFrameParser(1024)
	assert.NoError(t, err)

	parser.state = ppmParserStateReadPayload
	parser.expectedPayloadBytes = 1
	parser.payloadBytesRead = 2
	parser.pending = []byte{0x01}
	parser.offset = 0

	events, err := parser.Ingest(nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPPMInternalInvariant)
	assert.Empty(t, events)
}

func TestPPMFrameParser_ReadPayloadHandlesExactCompletion(t *testing.T) {
	parser, err := NewPPMFrameParser(1024)
	assert.NoError(t, err)

	parser.state = ppmParserStateReadPayload
	parser.expectedPayloadBytes = 3
	parser.payloadBytesRead = 3
	parser.pending = []byte("P6\n")
	parser.offset = 0
	parser.chunkBuffer = parser.chunkBuffer[:0]

	events, err := parser.Ingest(nil)
	assert.NoError(t, err)
	assert.Empty(t, events)
	assert.Equal(t, ppmParserStateReadWidth, parser.state)
}

func TestPPMFrameParser_SkipWhitespaceHandlesIncompleteComment(t *testing.T) {
	parser, err := NewPPMFrameParser(1024)
	assert.NoError(t, err)

	parser.state = ppmParserStateReadWidth
	commentBytes := []byte("# comment without newline")
	parser.pending = append([]byte(nil), commentBytes...)
	parser.offset = 0

	events, err := parser.Ingest(nil)
	assert.NoError(t, err)
	assert.Empty(t, events)
	assert.Equal(t, ppmParserStateReadWidth, parser.state)
	assert.Zero(t, parser.offset)
	assert.Equal(t, commentBytes, parser.pending)
}

func TestPPMFrameParser_ReadPayloadCompactsWhenIncomplete(t *testing.T) {
	parser, err := NewPPMFrameParser(6)
	assert.NoError(t, err)

	parser.state = ppmParserStateReadPayload
	parser.currentFrameID = 1
	parser.expectedPayloadBytes = 6
	parser.payloadBytesRead = 0
	parser.pending = []byte{0x01, 0x02, 0x03}
	parser.offset = 0
	parser.chunkBuffer = parser.chunkBuffer[:0]

	events, err := parser.Ingest(nil)
	assert.NoError(t, err)
	assert.Empty(t, events)
	assert.Equal(t, ppmParserStateReadPayload, parser.state)
	assert.Equal(t, 3, parser.payloadBytesRead)
	assert.Zero(t, parser.offset)
	assert.Len(t, parser.chunkBuffer, 3)
}

func TestPPMFrameParser_InvalidMagic(t *testing.T) {
	parser, err := NewPPMFrameParser(1024)
	assert.NoError(t, err)

	_, err = parser.Ingest([]byte("P3\n"))
	assert.Error(t, err)

	header := "P6\n1 1\n255\n"
	pixels := []byte{0x10, 0x20, 0x30}
	events, err := parser.Ingest(append([]byte(header), pixels...))
	assert.NoError(t, err)
	assert.Len(t, events, 3)
}

func TestPPMFrameParser_FlushDetectsIncompleteFrame(t *testing.T) {
	parser, err := NewPPMFrameParser(1024)
	assert.NoError(t, err)

	header := []byte("P6\n2 2\n255\n")

	events, err := parser.Ingest(header)
	assert.NoError(t, err)
	assert.Len(t, events, 1)

	_, err = parser.Flush()
	assert.Error(t, err)

	fullFrame := append([]byte("P6\n2 2\n255\n"), bytes.Repeat([]byte{0x01, 0x02, 0x03}, 4)...)
	events, err = parser.Ingest(fullFrame)
	assert.NoError(t, err)
	assert.Len(t, events, 3)

	flushEvents, err := parser.Flush()
	assert.NoError(t, err)
	assert.Empty(t, flushEvents)
}

func TestNewPPMFrameParser_InvalidChunkSize(t *testing.T) {
	parser, err := NewPPMFrameParser(0)
	assert.Nil(t, parser)
	assert.Error(t, err)
}

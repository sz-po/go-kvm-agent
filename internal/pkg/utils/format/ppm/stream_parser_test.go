package ppm

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	internalMemory "github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/memory"
	memorySDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/memory"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

func TestParsePPMStreamSingleFrame(t *testing.T) {
	width := 2
	height := 2
	payload := bytes.Repeat([]byte{0x01, 0x02, 0x03}, width*height)
	ppm := buildPPM(width, height, payload)

	pool := newTrackingPool(len(payload))
	var frames []*peripheralSDK.DisplayFrameBuffer
	handler := func(buffer *peripheralSDK.DisplayFrameBuffer) error {
		frames = append(frames, buffer)
		return nil
	}

	err := ParseStream(context.Background(), bytes.NewReader(ppm), handler, WithStreamParserMemoryBufferPool(pool))
	assertStreamTerminatedWithEOF(t, err)
	if assert.Len(t, frames, 1) {
		assert.Equal(t, len(payload), frames[0].GetSize())
		assertBufferContent(t, frames[0], payload)
		assert.False(t, pool.lastBuffer().released)
		assert.NoError(t, frames[0].Release())
		assert.True(t, pool.lastBuffer().released)
	}
}

func TestParsePPMStreamUsesDefaultMemoryPool(t *testing.T) {
	width := 1
	height := 1
	payload := []byte{0x7A, 0x7B, 0x7C}
	ppm := buildPPM(width, height, payload)

	ensureDefaultMemoryPool(t, len(payload))

	var frames []*peripheralSDK.DisplayFrameBuffer
	handler := func(buffer *peripheralSDK.DisplayFrameBuffer) error {
		frames = append(frames, buffer)
		return nil
	}

	err := ParseStream(context.Background(), bytes.NewReader(ppm), handler)
	assertStreamTerminatedWithEOF(t, err)
	if assert.Len(t, frames, 1) {
		assertBufferContent(t, frames[0], payload)
		assert.NoError(t, frames[0].Release())
	}
}

func TestParsePPMStreamMultipleFrames(t *testing.T) {
	frameOnePayload := bytes.Repeat([]byte{0x10, 0x20, 0x30}, 4)
	frameTwoPayload := bytes.Repeat([]byte{0xAA, 0xBB, 0xCC}, 2)

	stream := append(buildPPM(2, 2, frameOnePayload), buildPPM(1, 2, frameTwoPayload)...)

	pool := newTrackingPool(max(len(frameOnePayload), len(frameTwoPayload)))
	var sizes []int

	handler := func(buffer *peripheralSDK.DisplayFrameBuffer) error {
		sizes = append(sizes, buffer.GetSize())
		return buffer.Release()
	}

	err := ParseStream(context.Background(), bytes.NewReader(stream), handler, WithStreamParserMemoryBufferPool(pool))
	assertStreamTerminatedWithEOF(t, err)
	assert.Equal(t, []int{len(frameOnePayload), len(frameTwoPayload)}, sizes)
}

func TestParsePPMStreamCancelledContextStopsParsing(t *testing.T) {
	frameOnePayload := bytes.Repeat([]byte{0x10, 0x20, 0x30}, 4)
	frameTwoPayload := bytes.Repeat([]byte{0xAA, 0xBB, 0xCC}, 2)
	stream := append(buildPPM(2, 2, frameOnePayload), buildPPM(1, 2, frameTwoPayload)...)

	pool := newTrackingPool(max(len(frameOnePayload), len(frameTwoPayload)))
	frameCount := 0

	ctx, cancel := context.WithCancel(context.Background())

	handler := func(buffer *peripheralSDK.DisplayFrameBuffer) error {
		frameCount++
		if frameCount == 1 {
			cancel()
		}
		return buffer.Release()
	}

	err := ParseStream(ctx, bytes.NewReader(stream), handler, WithStreamParserMemoryBufferPool(pool))
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, frameCount)
}

func TestParsePPMStreamHeaderWithComments(t *testing.T) {
	width := 3
	height := 1
	payload := bytes.Repeat([]byte{0x00, 0x11, 0x22}, width*height)

	header := fmt.Sprintf("P6\n# initial comment\n%d # inline width comment\n# comment before height\n%d\n# comment before max\n255\n", width, height)
	ppm := append([]byte(header), payload...)

	pool := newTrackingPool(len(payload))
	handler := func(buffer *peripheralSDK.DisplayFrameBuffer) error {
		assertBufferContent(t, buffer, payload)
		return buffer.Release()
	}

	err := ParseStream(context.Background(), bytes.NewReader(ppm), handler, WithStreamParserMemoryBufferPool(pool))
	assertStreamTerminatedWithEOF(t, err)
}

func TestParsePPMStreamInvalidMagic(t *testing.T) {
	payload := bytes.Repeat([]byte{0xFF, 0x00, 0xFF}, 1)
	ppm := append([]byte("P3\n1 1\n255\n"), payload...)

	pool := newTrackingPool(len(payload))
	err := ParseStream(context.Background(), bytes.NewReader(ppm), func(_ *peripheralSDK.DisplayFrameBuffer) error { return nil }, WithStreamParserMemoryBufferPool(pool))

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidHeader)
	assert.ErrorIs(t, err, ErrInvalidMagic)
}

func TestParsePPMStreamInvalidMaxWidthConfiguration(t *testing.T) {
	pool := newTrackingPool(1)

	err := ParseStream(context.Background(), bytes.NewReader(nil), func(_ *peripheralSDK.DisplayFrameBuffer) error { return nil }, WithStreamParserMemoryBufferPool(pool), WithStreamParserMaxWidth(0))

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidConfiguration)
	assert.ErrorContains(t, err, "max width")
}

func TestParsePPMStreamInvalidMaxHeightConfiguration(t *testing.T) {
	pool := newTrackingPool(1)

	err := ParseStream(context.Background(), bytes.NewReader(nil), func(_ *peripheralSDK.DisplayFrameBuffer) error { return nil }, WithStreamParserMemoryBufferPool(pool), WithStreamParserMaxHeight(0))

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidConfiguration)
	assert.ErrorContains(t, err, "max height")
}

func TestParsePPMStreamWidthExceedsMaximum(t *testing.T) {
	width := 4
	height := 1
	payload := bytes.Repeat([]byte{0x01, 0x02, 0x03}, width*height)
	ppm := buildPPM(width, height, payload)

	pool := newTrackingPool(len(payload))
	err := ParseStream(context.Background(), bytes.NewReader(ppm), func(_ *peripheralSDK.DisplayFrameBuffer) error { return nil }, WithStreamParserMemoryBufferPool(pool), WithStreamParserMaxWidth(2))

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidHeader)
	assert.ErrorIs(t, err, ErrInvalidDimensions)
	assert.ErrorContains(t, err, "width exceeds 2")
}

func TestParsePPMStreamIncompleteFrame(t *testing.T) {
	width := 2
	height := 1
	payload := bytes.Repeat([]byte{0x01, 0x02, 0x03}, width*height)
	ppm := buildPPM(width, height, payload)

	pool := newIncompleteTrackingPool(len(payload), 1)
	err := ParseStream(context.Background(), bytes.NewReader(ppm), func(_ *peripheralSDK.DisplayFrameBuffer) error { return nil }, WithStreamParserMemoryBufferPool(pool))

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrIncompleteFrame)
	assert.ErrorContains(t, err, "read payload")
}

func TestParsePPMStreamMissingWidthToken(t *testing.T) {
	parser := newPPMStreamParserForTest([]byte("P6\n"))

	_, err := parser.readHeader()
	assert.Error(t, err)
	assert.ErrorContains(t, err, "read next token")
	assert.ErrorIs(t, err, io.EOF)
}

func TestParsePPMStreamNonNumericWidth(t *testing.T) {
	parser := newPPMStreamParserForTest([]byte("P6\nabc 1\n255\n"))

	_, err := parser.readHeader()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidDimensions)
	assert.ErrorContains(t, err, "ParseUint")
}

func TestParsePPMStreamInvalidDimension(t *testing.T) {
	ppm := []byte("P6\n0 1\n255\n")

	pool := newTrackingPool(1)
	err := ParseStream(context.Background(), bytes.NewReader(ppm), func(_ *peripheralSDK.DisplayFrameBuffer) error { return nil }, WithStreamParserMemoryBufferPool(pool))

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidHeader)
	assert.ErrorIs(t, err, ErrInvalidDimensions)
	assert.ErrorContains(t, err, "width equals zero")
}

func TestParsePPMStreamZeroHeight(t *testing.T) {
	parser := newPPMStreamParserForTest([]byte("P6\n1 0\n255\n"))

	_, err := parser.readHeader()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidDimensions)
	assert.ErrorContains(t, err, "height equals zero")
}

func TestParsePPMStreamInvalidMaxValue(t *testing.T) {
	ppm := []byte("P6\n1 1\n254\n")

	pool := newTrackingPool(3)
	err := ParseStream(context.Background(), bytes.NewReader(ppm), func(_ *peripheralSDK.DisplayFrameBuffer) error { return nil }, WithStreamParserMemoryBufferPool(pool))

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidHeader)
	assert.ErrorIs(t, err, ErrInvalidDepth)
	assert.ErrorContains(t, err, "max value not equals 255")
}

func TestParsePPMStreamNonNumericMaxValue(t *testing.T) {
	parser := newPPMStreamParserForTest([]byte("P6\n1 1\nabc\n"))

	_, err := parser.readHeader()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidDepth)
	assert.ErrorContains(t, err, "ParseUint")
}

func TestParsePPMStreamMissingHeightToken(t *testing.T) {
	parser := newPPMStreamParserForTest([]byte("P6\n1\n"))

	_, err := parser.readHeader()
	assert.Error(t, err)
	assert.ErrorContains(t, err, "read next token")
	assert.ErrorIs(t, err, io.EOF)
}

func TestParsePPMStreamNonNumericHeight(t *testing.T) {
	parser := newPPMStreamParserForTest([]byte("P6\n1 abc\n255\n"))

	_, err := parser.readHeader()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidDimensions)
	assert.ErrorContains(t, err, "ParseUint")
}

func TestParsePPMStreamMissingMaxValueToken(t *testing.T) {
	parser := newPPMStreamParserForTest([]byte("P6\n1 1\n"))

	_, err := parser.readHeader()
	assert.Error(t, err)
	assert.ErrorContains(t, err, "read next token")
	assert.ErrorIs(t, err, io.EOF)
}

func TestParsePPMStreamHandlerErrorReleasesBuffer(t *testing.T) {
	width := 1
	height := 1
	payload := []byte{0x01, 0x02, 0x03}
	ppm := buildPPM(width, height, payload)

	pool := newTrackingPool(len(payload))
	handlerErr := errors.New("handler failed")

	err := ParseStream(context.Background(), bytes.NewReader(ppm), func(buffer *peripheralSDK.DisplayFrameBuffer) error {
		return handlerErr
	}, WithStreamParserMemoryBufferPool(pool))

	assert.Error(t, err)
	assert.ErrorIs(t, err, handlerErr)
	if assert.NotNil(t, pool.lastBuffer()) {
		assert.True(t, pool.lastBuffer().released)
	}
}

func TestParsePPMStreamBufferCapacityTooLow(t *testing.T) {
	width := 2
	height := 1
	payload := bytes.Repeat([]byte{0x01, 0x02, 0x03}, width*height)
	ppm := buildPPM(width, height, payload)

	pool := newTrackingPool(len(payload) - 1)

	err := ParseStream(context.Background(), bytes.NewReader(ppm), func(_ *peripheralSDK.DisplayFrameBuffer) error { return nil }, WithStreamParserMemoryBufferPool(pool))

	assert.Error(t, err)
	assert.ErrorContains(t, err, "buffer size not supported")
	assert.ErrorIs(t, err, memorySDK.ErrBufferSizeNotSupported)
}

func TestParsePPMStreamChunkedReader(t *testing.T) {
	width := 2
	height := 2
	payload := bytes.Repeat([]byte{0x01, 0x02, 0x03}, width*height)
	ppm := buildPPM(width, height, payload)

	pool := newTrackingPool(len(payload))
	chunkReader := &fixedChunkReader{data: ppm, chunkSize: 3}

	captured := make([][]byte, 0)
	handler := func(buffer *peripheralSDK.DisplayFrameBuffer) error {
		captured = append(captured, readBuffer(t, buffer))
		return buffer.Release()
	}

	err := ParseStream(context.Background(), chunkReader, handler, WithStreamParserMemoryBufferPool(pool))
	assertStreamTerminatedWithEOF(t, err)
	if assert.Len(t, captured, 1) {
		assert.Equal(t, payload, captured[0])
	}
}

func assertStreamTerminatedWithEOF(t *testing.T, err error) {
	t.Helper()

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidHeader)
	assert.ErrorIs(t, err, io.EOF)
}

func newPPMStreamParserForTest(input []byte) *streamParser {
	return &streamParser{
		reader: bufio.NewReader(bytes.NewReader(input)),
		maxWidth:  hostMaxInt,
		maxHeight: hostMaxInt,
	}
}

func ensureDefaultMemoryPool(t *testing.T, capacity int) {
	t.Helper()

	pool, err := internalMemory.GetDefaultMemoryPool()
	if err == nil {
		if trackingPool, ok := pool.(*trackingPool); ok && trackingPool.maxCapacity < capacity {
			trackingPool.maxCapacity = capacity
		}
		return
	}

	trackingPool := newTrackingPool(capacity)
	setErr := internalMemory.SetDefaultMemoryPool(trackingPool)
	assert.NoError(t, setErr)
}

func buildPPM(width int, height int, payload []byte) []byte {
	header := fmt.Sprintf("P6\n%d %d\n255\n", width, height)
	ppm := make([]byte, len(header)+len(payload))
	copy(ppm, header)
	copy(ppm[len(header):], payload)
	return ppm
}

func assertBufferContent(t *testing.T, buffer *peripheralSDK.DisplayFrameBuffer, expected []byte) {
	t.Helper()

	result := readBuffer(t, buffer)
	assert.Equal(t, expected, result)
}

func readBuffer(t *testing.T, buffer *peripheralSDK.DisplayFrameBuffer) []byte {
	t.Helper()

	var out bytes.Buffer
	_, err := buffer.WriteTo(&out)
	assert.NoError(t, err)
	return out.Bytes()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type trackingPool struct {
	maxCapacity int
	buffers     []*testBuffer
}

func newTrackingPool(maxCapacity int) *trackingPool {
	return &trackingPool{maxCapacity: maxCapacity}
}

func (pool *trackingPool) Borrow(size int) (memorySDK.Buffer, error) {
	if size > pool.maxCapacity {
		return nil, memorySDK.ErrBufferSizeNotSupported
	}

	buffer := newTestBuffer(size)
	pool.buffers = append(pool.buffers, buffer)
	return buffer, nil
}

func (pool *trackingPool) lastBuffer() *testBuffer {
	if len(pool.buffers) == 0 {
		return nil
	}
	return pool.buffers[len(pool.buffers)-1]
}

type incompleteTrackingPool struct {
	maxCapacity int
	shortfall   int
}

func newIncompleteTrackingPool(maxCapacity int, shortfall int) *incompleteTrackingPool {
	return &incompleteTrackingPool{maxCapacity: maxCapacity, shortfall: shortfall}
}

func (pool *incompleteTrackingPool) Borrow(size int) (memorySDK.Buffer, error) {
	if size > pool.maxCapacity {
		return nil, memorySDK.ErrBufferSizeNotSupported
	}

	bytesToRead := size - pool.shortfall
	if bytesToRead < 0 {
		bytesToRead = 0
	}

	return newShortReadBuffer(size, bytesToRead), nil
}

type shortReadBuffer struct {
	*testBuffer
	bytesToRead int
}

func newShortReadBuffer(capacity int, bytesToRead int) *shortReadBuffer {
	if bytesToRead > capacity {
		bytesToRead = capacity
	}
	if bytesToRead < 0 {
		bytesToRead = 0
	}

	return &shortReadBuffer{
		testBuffer: newTestBuffer(capacity),
		bytesToRead: bytesToRead,
	}
}

func (buffer *shortReadBuffer) ReadFrom(reader io.Reader) (int64, error) {
	if buffer.bytesToRead == 0 {
		return 0, nil
	}

	temporary := make([]byte, buffer.bytesToRead)
	bytesRead, err := reader.Read(temporary)
	if bytesRead > 0 {
		buffer.data = append(buffer.data, temporary[:bytesRead]...)
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return int64(bytesRead), err
	}
	return int64(bytesRead), nil
}

type testBuffer struct {
	capacity int
	data     []byte
	released bool
}

func newTestBuffer(capacity int) *testBuffer {
	return &testBuffer{
		capacity: capacity,
		data:     make([]byte, 0, capacity),
	}
}

func (buffer *testBuffer) GetCapacity() int {
	return buffer.capacity
}

func (buffer *testBuffer) GetSize() int {
	return len(buffer.data)
}

func (buffer *testBuffer) WriteTo(writer io.Writer) (int64, error) {
	written, err := writer.Write(buffer.data)
	return int64(written), err
}

func (buffer *testBuffer) Write(input []byte) (int, error) {
	remaining := buffer.capacity - len(buffer.data)
	if len(input) > remaining {
		return 0, memorySDK.ErrBufferTooSmall
	}
	buffer.data = append(buffer.data, input...)
	return len(input), nil
}

func (buffer *testBuffer) ReadFrom(reader io.Reader) (int64, error) {
	total := int64(0)
	temp := make([]byte, buffer.capacity)

	for {
		remaining := buffer.capacity - len(buffer.data)
		if remaining == 0 {
			return total, nil
		}

		if len(temp) > remaining {
			temp = temp[:remaining]
		}

		bytesRead, err := reader.Read(temp)
		if bytesRead > 0 {
			buffer.data = append(buffer.data, temp[:bytesRead]...)
			total += int64(bytesRead)
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				if bytesRead == 0 {
					return total, io.EOF
				}
				return total, nil
			}
			return total, err
		}

		if bytesRead == 0 {
			continue
		}
	}
}

func (buffer *testBuffer) Retain() error {
	return nil
}

func (buffer *testBuffer) Release() error {
	buffer.released = true
	buffer.data = buffer.data[:0]
	return nil
}

type fixedChunkReader struct {
	data      []byte
	chunkSize int
	offset    int
}

func (reader *fixedChunkReader) Read(destination []byte) (int, error) {
	if reader.offset >= len(reader.data) {
		return 0, io.EOF
	}

	end := reader.offset + reader.chunkSize
	if end > len(reader.data) {
		end = len(reader.data)
	}

	bytesCopied := copy(destination, reader.data[reader.offset:end])
	reader.offset += bytesCopied

	return bytesCopied, nil
}

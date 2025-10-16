package ppm

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/memory"
	memorySDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/memory"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

// FrameHandler is called when a complete frame is ready in the buffer.
// The caller takes ownership of the buffer and must release it back to the pool.
type FrameHandler func(buffer *peripheralSDK.DisplayFrameBuffer) error

var (
	ErrInvalidConfiguration = errors.New("invalid configuration")
	ErrIncompleteFrame      = errors.New("incomplete frame")
	ErrInvalidHeader        = errors.New("invalid header")
	ErrInvalidMagic         = errors.New("invalid magic")
	ErrInvalidDimensions    = errors.New("invalid dimensions")
	ErrInvalidDepth         = errors.New("invalid depth")
)

var hostMaxInt = math.MaxInt

type StreamParserOpt func(*streamParser)

type streamParser struct {
	reader           *bufio.Reader
	handler          FrameHandler
	memoryBufferPool memorySDK.Pool
	context          context.Context

	maxWidth  int
	maxHeight int
}

func WithStreamParserMemoryBufferPool(pool memorySDK.Pool) StreamParserOpt {
	return func(parser *streamParser) {
		parser.memoryBufferPool = pool
	}
}

func WithStreamParserMaxWidth(maxWidth int) StreamParserOpt {
	return func(parser *streamParser) {
		parser.maxWidth = maxWidth
	}
}

func WithStreamParserMaxHeight(maxHeight int) StreamParserOpt {
	return func(parser *streamParser) {
		parser.maxHeight = maxHeight
	}
}

func ParseStream(ctx context.Context, reader io.Reader, handler FrameHandler, opts ...StreamParserOpt) error {
	bufferedReader := bufio.NewReader(reader)

	parser := &streamParser{
		reader:  bufferedReader,
		handler: handler,
		context: ctx,

		maxWidth:  1920,
		maxHeight: 1080,
	}

	for _, opt := range opts {
		opt(parser)
	}

	if parser.memoryBufferPool == nil {
		memoryBufferPool, err := memory.GetDefaultMemoryPool()
		if err != nil {
			return fmt.Errorf("get default memory pool: %w", err)
		}
		parser.memoryBufferPool = memoryBufferPool
	}

	if parser.maxWidth <= 0 {
		return fmt.Errorf("%w: max width must be greater than zero", ErrInvalidConfiguration)
	}

	if parser.maxHeight <= 0 {
		return fmt.Errorf("%w: max height must be greater than zero", ErrInvalidConfiguration)
	}

	return parser.parse()
}

func (parser *streamParser) parse() error {
	for {
		if err := parser.context.Err(); err != nil {
			return err
		}

		if err := parser.readFrame(); err != nil {
			return err
		}
	}
}

type ppmHeader struct {
	width        uint32
	height       uint32
	payloadBytes int
}

func (parser *streamParser) readFrame() error {
	header, err := parser.readHeader()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidHeader, err)
	}

	return parser.readPayload(header)
}

func (parser *streamParser) readHeader() (*ppmHeader, error) {
	magicToken, err := parser.readNextToken()
	if err != nil {
		return nil, fmt.Errorf("read next token: %w", err)
	}
	if magicToken != "P6" {
		return nil, ErrInvalidMagic
	}

	widthToken, err := parser.readNextToken()
	if err != nil {
		return nil, fmt.Errorf("read next token: %w", err)
	}
	widthValue, err := strconv.ParseUint(widthToken, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidDimensions, err)
	}
	if widthValue == 0 {
		return nil, fmt.Errorf("%w: width equals zero", ErrInvalidDimensions)
	}
	if widthValue > uint64(parser.maxWidth) {
		return nil, fmt.Errorf("%w: width exceeds %d", ErrInvalidDimensions, parser.maxWidth)
	}

	heightToken, err := parser.readNextToken()
	if err != nil {
		return nil, fmt.Errorf("read next token: %w", err)
	}
	heightValue, err := strconv.ParseUint(heightToken, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidDimensions, err)
	}
	if heightValue == 0 {
		return nil, fmt.Errorf("%w: height equals zero", ErrInvalidDimensions)
	}

	maxValueToken, err := parser.readNextToken()
	if err != nil {
		return nil, fmt.Errorf("read next token: %w", err)
	}
	maxValue, err := strconv.ParseUint(maxValueToken, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidDepth, err)
	}
	if maxValue != 255 {
		return nil, fmt.Errorf("%w: max value not equals 255", ErrInvalidDepth)
	}

	payloadBytes := uint64(widthValue) * uint64(heightValue) * 3

	return &ppmHeader{
		width:        uint32(widthValue),
		height:       uint32(heightValue),
		payloadBytes: int(payloadBytes),
	}, nil
}

func (parser *streamParser) readPayload(header *ppmHeader) error {
	buffer, err := parser.memoryBufferPool.Borrow(header.payloadBytes)
	if err != nil {
		return fmt.Errorf("borrow buffer from pool: %w", err)
	}

	limitReader := io.LimitReader(parser.reader, int64(header.payloadBytes))
	bytesRead, err := buffer.ReadFrom(limitReader)
	if err != nil {
		_ = buffer.Release()
		return fmt.Errorf("read payload: %w", err)
	}
	if bytesRead != int64(header.payloadBytes) {
		_ = buffer.Release()
		return fmt.Errorf("read payload: %w", ErrIncompleteFrame)
	}

	frameBuffer := peripheralSDK.NewDisplayFrameBuffer(buffer)
	if handlerErr := parser.handler(frameBuffer); handlerErr != nil {
		_ = buffer.Release()
		return fmt.Errorf("frame handler: %w", handlerErr)
	}

	return nil
}

func (parser *streamParser) readNextToken() (string, error) {
	var tokenBuilder strings.Builder
	tokenBuilder.Grow(16)

	for {
		byteValue, err := parser.reader.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if tokenBuilder.Len() > 0 {
					return tokenBuilder.String(), nil
				}
				return "", io.EOF
			}
			return "", fmt.Errorf("read byte: %w", err)
		}

		if tokenBuilder.Len() == 0 {
			if isWhitespace(byteValue) {
				continue
			}
			if byteValue == '#' {
				if err := parser.discardComment(); err != nil {
					return "", err
				}
				continue
			}
			tokenBuilder.WriteByte(byteValue)
			continue
		}

		if isWhitespace(byteValue) {
			return tokenBuilder.String(), nil
		}

		if byteValue == '#' {
			if err := parser.discardComment(); err != nil {
				return "", err
			}
			return tokenBuilder.String(), nil
		}

		tokenBuilder.WriteByte(byteValue)
	}
}

func (parser *streamParser) discardComment() error {
	for {
		byteValue, err := parser.reader.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("skip comment: %w", err)
		}
		if byteValue == '\n' || byteValue == '\r' {
			return nil
		}
	}
}

func isWhitespace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\n' || value == '\r'
}

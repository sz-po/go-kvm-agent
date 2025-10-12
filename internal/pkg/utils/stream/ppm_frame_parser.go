package stream

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

var (
	ErrPPMInvalidMagic      = errors.New("ppm parser invalid magic")
	ErrPPMInvalidDimension  = errors.New("ppm parser invalid dimension")
	ErrPPMInvalidMaxValue   = errors.New("ppm parser invalid max value")
	ErrPPMIncompleteFrame   = errors.New("ppm parser incomplete frame")
	ErrPPMChunkSizeInvalid  = errors.New("ppm parser chunk size invalid")
	ErrPPMInternalInvariant = errors.New("ppm parser internal invariant failure")
)

var hostMaxInt = math.MaxInt

type ppmParserState int

const (
	ppmParserStateReadMagic ppmParserState = iota
	ppmParserStateReadWidth
	ppmParserStateReadHeight
	ppmParserStateReadMaxValue
	ppmParserStateAwaitPayloadStart
	ppmParserStateReadPayload
)

type PPMFrameParser struct {
	chunkSize int

	state ppmParserState

	pending []byte
	offset  int

	nextFrameID uint64

	currentFrameID       uint64
	currentWidth         uint32
	currentHeight        uint32
	expectedPayloadBytes int
	payloadBytesRead     int
	chunkBuffer          []byte
	nextChunkIndex       uint32
	totalChunksEmitted   uint32
}

func NewPPMFrameParser(chunkSize int) (*PPMFrameParser, error) {
	if chunkSize <= 0 {
		return nil, fmt.Errorf("new ppm frame parser: %w", ErrPPMChunkSizeInvalid)
	}

	parser := &PPMFrameParser{
		chunkSize:          chunkSize,
		state:              ppmParserStateReadMagic,
		pending:            make([]byte, 0),
		offset:             0,
		nextFrameID:        1,
		chunkBuffer:        make([]byte, 0, chunkSize),
		nextChunkIndex:     0,
		totalChunksEmitted: 0,
	}

	return parser, nil
}

func (parser *PPMFrameParser) Ingest(data []byte) ([]peripheralSDK.DisplayDataEvent, error) {
	if len(data) > 0 {
		parser.pending = append(parser.pending, data...)
	}

	events := make([]peripheralSDK.DisplayDataEvent, 0)

	for {
		switch parser.state {
		case ppmParserStateReadMagic:
			token, complete, err := parser.readToken()
			if err != nil {
				parser.Reset()
				return events, err
			}
			if !complete {
				return events, nil
			}
			if token != "P6" {
				parser.Reset()
				return events, fmt.Errorf("ppm parser: expected magic P6 but received %q: %w", token, ErrPPMInvalidMagic)
			}
			parser.state = ppmParserStateReadWidth
			parser.compact()
		case ppmParserStateReadWidth:
			token, complete, err := parser.readToken()
			if err != nil {
				parser.Reset()
				return events, err
			}
			if !complete {
				return events, nil
			}
			widthValue, parseErr := strconv.ParseUint(token, 10, 32)
			if parseErr != nil || widthValue == 0 {
				parser.Reset()
				return events, fmt.Errorf("ppm parser: parse width: %w", ErrPPMInvalidDimension)
			}
			parser.currentWidth = uint32(widthValue)
			parser.state = ppmParserStateReadHeight
			parser.compact()
		case ppmParserStateReadHeight:
			token, complete, err := parser.readToken()
			if err != nil {
				parser.Reset()
				return events, err
			}
			if !complete {
				return events, nil
			}
			heightValue, parseErr := strconv.ParseUint(token, 10, 32)
			if parseErr != nil || heightValue == 0 {
				parser.Reset()
				return events, fmt.Errorf("ppm parser: parse height: %w", ErrPPMInvalidDimension)
			}
			parser.currentHeight = uint32(heightValue)
			parser.state = ppmParserStateReadMaxValue
			parser.compact()
		case ppmParserStateReadMaxValue:
			token, complete, err := parser.readToken()
			if err != nil {
				parser.Reset()
				return events, err
			}
			if !complete {
				return events, nil
			}
			maxValue, parseErr := strconv.ParseUint(token, 10, 16)
			if parseErr != nil || maxValue != 255 {
				parser.Reset()
				return events, fmt.Errorf("ppm parser: parse max value: %w", ErrPPMInvalidMaxValue)
			}

			payloadBytes := uint64(parser.currentWidth) * uint64(parser.currentHeight) * 3
			if payloadBytes == 0 || payloadBytes > math.MaxInt64 {
				parser.Reset()
				return events, fmt.Errorf("ppm parser: compute payload size: %w", ErrPPMInvalidDimension)
			}
			if payloadBytes > uint64(hostMaxInt) {
				parser.Reset()
				return events, fmt.Errorf("ppm parser: payload exceeds host capacity: %w", ErrPPMInvalidDimension)
			}

			parser.expectedPayloadBytes = int(payloadBytes)
			parser.payloadBytesRead = 0
			parser.chunkBuffer = parser.chunkBuffer[:0]
			parser.nextChunkIndex = 0
			parser.totalChunksEmitted = 0
			parser.currentFrameID = parser.nextFrameID
			parser.nextFrameID++

			parser.state = ppmParserStateAwaitPayloadStart
			parser.compact()

			frameStart := peripheralSDK.NewDisplayFrameStartEvent(
				parser.currentFrameID,
				parser.currentWidth,
				parser.currentHeight,
				peripheralSDK.DisplayPixelFormatRGB24,
				time.Now(),
			)
			events = append(events, frameStart)
		case ppmParserStateAwaitPayloadStart:
			hasData, err := parser.skipWhitespaceAndComments()
			if err != nil {
				parser.Reset()
				return events, err
			}
			if !hasData {
				return events, nil
			}
			parser.state = ppmParserStateReadPayload
		case ppmParserStateReadPayload:
			availableBytes := len(parser.pending) - parser.offset
			if availableBytes <= 0 {
				parser.compact()
				return events, nil
			}

			bytesRemaining := parser.expectedPayloadBytes - parser.payloadBytesRead
			if bytesRemaining < 0 {
				parser.Reset()
				return events, fmt.Errorf("ppm parser: payload accounting mismatch: %w", ErrPPMInternalInvariant)
			}
			if bytesRemaining == 0 {
				parser.state = ppmParserStateReadMagic
				parser.compact()
				continue
			}

			bytesToProcess := availableBytes
			if bytesToProcess > bytesRemaining {
				bytesToProcess = bytesRemaining
			}

			segment := parser.pending[parser.offset : parser.offset+bytesToProcess]
			parser.offset += bytesToProcess
			parser.payloadBytesRead += bytesToProcess

			segmentOffset := 0
			for segmentOffset < len(segment) {
				requiredForChunk := parser.chunkSize - len(parser.chunkBuffer)
				if requiredForChunk > len(segment)-segmentOffset {
					requiredForChunk = len(segment) - segmentOffset
				}
				parser.chunkBuffer = append(parser.chunkBuffer, segment[segmentOffset:segmentOffset+requiredForChunk]...)
				segmentOffset += requiredForChunk

				if len(parser.chunkBuffer) == parser.chunkSize {
					chunkData := make([]byte, len(parser.chunkBuffer))
					copy(chunkData, parser.chunkBuffer)
					chunkEvent := peripheralSDK.NewDisplayFrameChunkEvent(
						parser.currentFrameID,
						parser.nextChunkIndex,
						chunkData,
						time.Now(),
					)
					events = append(events, chunkEvent)
					parser.nextChunkIndex++
					parser.totalChunksEmitted++
					parser.chunkBuffer = parser.chunkBuffer[:0]
				}
			}

			if parser.payloadBytesRead == parser.expectedPayloadBytes {
				if len(parser.chunkBuffer) > 0 {
					chunkData := make([]byte, len(parser.chunkBuffer))
					copy(chunkData, parser.chunkBuffer)
					chunkEvent := peripheralSDK.NewDisplayFrameChunkEvent(
						parser.currentFrameID,
						parser.nextChunkIndex,
						chunkData,
						time.Now(),
					)
					events = append(events, chunkEvent)
					parser.nextChunkIndex++
					parser.totalChunksEmitted++
					parser.chunkBuffer = parser.chunkBuffer[:0]
				}

				frameEnd := peripheralSDK.NewDisplayFrameEndEvent(
					parser.currentFrameID,
					parser.totalChunksEmitted,
					time.Now(),
				)
				events = append(events, frameEnd)

				parser.state = ppmParserStateReadMagic
				parser.compact()
			} else {
				parser.compact()
			}
		default:
			parser.Reset()
			return events, fmt.Errorf("ppm parser: unknown state: %w", ErrPPMInternalInvariant)
		}
	}
}

func (parser *PPMFrameParser) Flush() ([]peripheralSDK.DisplayDataEvent, error) {
	if parser.state != ppmParserStateReadMagic {
		parser.Reset()
		return nil, fmt.Errorf("ppm parser: flush with incomplete frame: %w", ErrPPMIncompleteFrame)
	}
	return nil, nil
}

func (parser *PPMFrameParser) Reset() {
	parser.state = ppmParserStateReadMagic
	parser.pending = parser.pending[:0]
	parser.offset = 0
	parser.currentFrameID = 0
	parser.currentWidth = 0
	parser.currentHeight = 0
	parser.expectedPayloadBytes = 0
	parser.payloadBytesRead = 0
	parser.chunkBuffer = parser.chunkBuffer[:0]
	parser.nextChunkIndex = 0
	parser.totalChunksEmitted = 0
}

func (parser *PPMFrameParser) readToken() (string, bool, error) {
	hasData, err := parser.skipWhitespaceAndComments()
	if err != nil {
		return "", false, err
	}
	if !hasData {
		return "", false, nil
	}

	startIndex := parser.offset
	index := startIndex

	for index < len(parser.pending) {
		currentByte := parser.pending[index]
		if isWhitespace(currentByte) || currentByte == '#' {
			break
		}
		index++
	}

	if index == len(parser.pending) {
		return "", false, nil
	}

	token := string(parser.pending[startIndex:index])
	if len(token) == 0 {
		return "", false, nil
	}

	parser.offset = index
	return token, true, nil
}

func (parser *PPMFrameParser) skipWhitespaceAndComments() (bool, error) {
	for {
		for parser.offset < len(parser.pending) {
			currentByte := parser.pending[parser.offset]
			if isWhitespace(currentByte) {
				parser.offset++
				continue
			}
			if currentByte == '#' {
				newlineIndex := indexByte(parser.pending[parser.offset:], '\n')
				if newlineIndex == -1 {
					parser.compact()
					return false, nil
				}
				parser.offset += newlineIndex + 1
				continue
			}
			return true, nil
		}

		parser.compact()
		if len(parser.pending) == 0 {
			parser.offset = 0
			return false, nil
		}
	}
}

func (parser *PPMFrameParser) compact() {
	if parser.offset == 0 {
		return
	}
	if parser.offset >= len(parser.pending) {
		parser.pending = parser.pending[:0]
		parser.offset = 0
		return
	}

	remaining := len(parser.pending) - parser.offset
	copy(parser.pending[:remaining], parser.pending[parser.offset:])
	parser.pending = parser.pending[:remaining]
	parser.offset = 0
}

func isWhitespace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\n' || value == '\r'
}

func indexByte(buffer []byte, target byte) int {
	for index := 0; index < len(buffer); index++ {
		if buffer[index] == target {
			return index
		}
	}
	return -1
}

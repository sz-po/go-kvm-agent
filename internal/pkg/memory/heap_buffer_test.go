package memory

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	memorypkg "github.com/szymonpodeszwa/go-kvm-agent/pkg/memory"
)

func TestNewHeapBufferRejectsNonPositiveCapacity(t *testing.T) {
	heapBuffer, err := newHeapBuffer(nil, 0)
	assert.Nil(t, heapBuffer)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "buffer capacity must be greater than zero")
}

func TestHeapBufferGetCapacityAndSize(t *testing.T) {
	heapBuffer, err := newHeapBuffer(nil, 8)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, 8, heapBuffer.GetCapacity())
	assert.Equal(t, 0, heapBuffer.GetSize())

	bytesRead, readErr := heapBuffer.ReadFrom(bytes.NewReader([]byte("four")))
	assert.NoError(t, readErr)
	assert.Equal(t, int64(4), bytesRead)
	assert.Equal(t, 8, heapBuffer.GetCapacity())
	assert.Equal(t, 4, heapBuffer.GetSize())
}

func TestHeapBufferReadAndWriteWithinCapacity(t *testing.T) {
	heapBuffer, err := newHeapBuffer(nil, 16)
	if !assert.NoError(t, err) {
		return
	}

	bytesRead, readErr := heapBuffer.ReadFrom(bytes.NewReader([]byte("hello")))
	assert.NoError(t, readErr)
	assert.Equal(t, int64(5), bytesRead)
	assert.Equal(t, 5, heapBuffer.GetSize())
	assert.Equal(t, []byte("hello"), heapBuffer.data)

	var destination bytes.Buffer
	writtenBytes, writeErr := heapBuffer.WriteTo(&destination)
	assert.NoError(t, writeErr)
	assert.Equal(t, int64(5), writtenBytes)
	assert.Equal(t, "hello", destination.String())
}

func TestHeapBufferWriteAppendsDataWithinCapacity(t *testing.T) {
	heapBuffer, err := newHeapBuffer(nil, 8)
	if !assert.NoError(t, err) {
		return
	}

	bytesWritten, writeErr := heapBuffer.Write([]byte("he"))
	assert.NoError(t, writeErr)
	assert.Equal(t, 2, bytesWritten)
	assert.Equal(t, 2, heapBuffer.GetSize())
	assert.Equal(t, 2, heapBuffer.writePos)
	assert.Equal(t, []byte("he"), heapBuffer.data)

	bytesWritten, writeErr = heapBuffer.Write([]byte("llo"))
	assert.NoError(t, writeErr)
	assert.Equal(t, 3, bytesWritten)
	assert.Equal(t, 5, heapBuffer.GetSize())
	assert.Equal(t, 5, heapBuffer.writePos)
	assert.Equal(t, []byte("hello"), heapBuffer.data)
}

func TestHeapBufferWriteAppendsAfterReadFrom(t *testing.T) {
	heapBuffer, err := newHeapBuffer(nil, 8)
	if !assert.NoError(t, err) {
		return
	}

	bytesRead, readErr := heapBuffer.ReadFrom(bytes.NewReader([]byte("abcd")))
	assert.NoError(t, readErr)
	assert.Equal(t, int64(4), bytesRead)
	assert.Equal(t, 4, heapBuffer.GetSize())
	assert.Equal(t, 4, heapBuffer.writePos)
	assert.Equal(t, []byte("abcd"), heapBuffer.data)

	bytesWritten, writeErr := heapBuffer.Write([]byte("ef"))
	assert.NoError(t, writeErr)
	assert.Equal(t, 2, bytesWritten)
	assert.Equal(t, 6, heapBuffer.GetSize())
	assert.Equal(t, 6, heapBuffer.writePos)
	assert.Equal(t, []byte("abcdef"), heapBuffer.data)
}

func TestHeapBufferReadFromAppendsAfterWrite(t *testing.T) {
	heapBuffer, err := newHeapBuffer(nil, 8)
	if !assert.NoError(t, err) {
		return
	}

	bytesWritten, writeErr := heapBuffer.Write([]byte("ab"))
	assert.NoError(t, writeErr)
	assert.Equal(t, 2, bytesWritten)
	assert.Equal(t, 2, heapBuffer.GetSize())
	assert.Equal(t, 2, heapBuffer.writePos)
	assert.Equal(t, []byte("ab"), heapBuffer.data)

	bytesRead, readErr := heapBuffer.ReadFrom(bytes.NewReader([]byte("cd")))
	assert.NoError(t, readErr)
	assert.Equal(t, int64(2), bytesRead)
	assert.Equal(t, 4, heapBuffer.GetSize())
	assert.Equal(t, 4, heapBuffer.writePos)
	assert.Equal(t, []byte("abcd"), heapBuffer.data)
}

func TestHeapBufferReadFromRestoresSliceLengthWhenShrunk(t *testing.T) {
	heapBuffer, err := newHeapBuffer(nil, 4)
	if !assert.NoError(t, err) {
		return
	}

	bytesWritten, writeErr := heapBuffer.Write([]byte("abc"))
	assert.NoError(t, writeErr)
	assert.Equal(t, 3, bytesWritten)
	assert.Equal(t, 3, heapBuffer.GetSize())
	assert.Equal(t, 3, heapBuffer.writePos)
	assert.Equal(t, []byte("abc"), heapBuffer.data)

	heapBuffer.data = heapBuffer.data[:1]
	assert.Equal(t, 1, len(heapBuffer.data))
	assert.Equal(t, 3, heapBuffer.writePos)

	bytesRead, readErr := heapBuffer.ReadFrom(bytes.NewReader([]byte("d")))
	assert.NoError(t, readErr)
	assert.Equal(t, int64(1), bytesRead)
	assert.Equal(t, 4, heapBuffer.GetSize())
	assert.Equal(t, 4, heapBuffer.writePos)
	assert.Equal(t, []byte("abcd"), heapBuffer.data)
}

func TestHeapBufferReadFromErrorsWhenWritePosExceeded(t *testing.T) {
	heapBuffer, err := newHeapBuffer(nil, 2)
	if !assert.NoError(t, err) {
		return
	}

	heapBuffer.writePos = 3
	heapBuffer.data = heapBuffer.data[:2]

	bytesRead, readErr := heapBuffer.ReadFrom(bytes.NewReader([]byte("x")))
	assert.Equal(t, int64(0), bytesRead)
	assert.Error(t, readErr)
	assert.ErrorContains(t, readErr, "heap buffer write position exceeds capacity")
}

func TestHeapBufferWriteRestoresSliceLengthWhenShrunk(t *testing.T) {
	heapBuffer, err := newHeapBuffer(nil, 8)
	if !assert.NoError(t, err) {
		return
	}

	bytesWritten, writeErr := heapBuffer.Write([]byte("abcd"))
	assert.NoError(t, writeErr)
	assert.Equal(t, 4, bytesWritten)
	assert.Equal(t, 4, heapBuffer.GetSize())
	assert.Equal(t, 4, heapBuffer.writePos)
	assert.Equal(t, []byte("abcd"), heapBuffer.data)

	heapBuffer.data = heapBuffer.data[:2]
	assert.Equal(t, 2, len(heapBuffer.data))
	assert.Equal(t, 4, heapBuffer.writePos)

	bytesWritten, writeErr = heapBuffer.Write([]byte("ef"))
	assert.NoError(t, writeErr)
	assert.Equal(t, 2, bytesWritten)
	assert.Equal(t, 6, heapBuffer.GetSize())
	assert.Equal(t, 6, heapBuffer.writePos)
	assert.Equal(t, []byte("abcdef"), heapBuffer.data)
}

func TestHeapBufferWriteReturnsErrorWhenExceedingCapacity(t *testing.T) {
	heapBuffer, err := newHeapBuffer(nil, 4)
	if !assert.NoError(t, err) {
		return
	}

	bytesWritten, writeErr := heapBuffer.Write([]byte("data"))
	assert.NoError(t, writeErr)
	assert.Equal(t, 4, bytesWritten)
	assert.Equal(t, 4, heapBuffer.GetSize())
	assert.Equal(t, 4, heapBuffer.writePos)
	assert.Equal(t, []byte("data"), heapBuffer.data)

	bytesWritten, writeErr = heapBuffer.Write([]byte("z"))
	assert.Error(t, writeErr)
	assert.ErrorIs(t, writeErr, memorypkg.ErrBufferTooSmall)
	assert.Equal(t, 0, bytesWritten)
	assert.Equal(t, 4, heapBuffer.GetSize())
	assert.Equal(t, 4, heapBuffer.writePos)
	assert.Equal(t, []byte("data"), heapBuffer.data)
}

func TestHeapBufferReadAllFromTruncatesToCapacity(t *testing.T) {
	heapBuffer, err := newHeapBuffer(nil, 4)
	if !assert.NoError(t, err) {
		return
	}

	bytesRead, readErr := heapBuffer.ReadFrom(bytes.NewReader([]byte("hello")))
	assert.NoError(t, readErr)
	assert.Equal(t, int64(4), bytesRead)
	assert.Equal(t, 4, heapBuffer.GetSize())
	assert.Equal(t, []byte("hell"), heapBuffer.data)

	var destination bytes.Buffer
	writtenBytes, writeErr := heapBuffer.WriteTo(&destination)
	assert.NoError(t, writeErr)
	assert.Equal(t, int64(4), writtenBytes)
	assert.Equal(t, "hell", destination.String())
}

func TestHeapBufferReadAllFromStopsOnZeroProgress(t *testing.T) {
	heapBuffer, err := newHeapBuffer(nil, 4)
	if !assert.NoError(t, err) {
		return
	}

	bytesRead, readErr := heapBuffer.ReadFrom(&zeroProgressReader{})
	assert.NoError(t, readErr)
	assert.Equal(t, int64(0), bytesRead)
	assert.Equal(t, 0, heapBuffer.GetSize())
}

func TestHeapBufferReadAllFromPropagatesErrors(t *testing.T) {
	partialData := []byte("abcd")
	readError := errors.New("boom")

	heapBuffer, err := newHeapBuffer(nil, 8)
	if !assert.NoError(t, err) {
		return
	}

	bytesRead, readErr := heapBuffer.ReadFrom(&failingReader{
		payload:   partialData,
		targetErr: readError,
	})
	assert.Equal(t, int64(len(partialData)), bytesRead)
	assert.Error(t, readErr)
	assert.ErrorContains(t, readErr, "read into heap buffer")
	assert.ErrorIs(t, readErr, readError)
	assert.Equal(t, partialData, heapBuffer.data)
}

func TestHeapBufferWriteAllToPropagatesErrors(t *testing.T) {
	writeError := errors.New("cannot write")
	heapBuffer, err := newHeapBuffer(nil, 4)
	if !assert.NoError(t, err) {
		return
	}

	bytesRead, readErr := heapBuffer.ReadFrom(bytes.NewReader([]byte("data")))
	assert.NoError(t, readErr)
	assert.Equal(t, int64(4), bytesRead)

	writtenBytes, writeErr := heapBuffer.WriteTo(&failingWriter{targetErr: writeError})
	assert.Equal(t, int64(0), writtenBytes)
	assert.Error(t, writeErr)
	assert.ErrorContains(t, writeErr, "write heap buffer to writer")
	assert.ErrorIs(t, writeErr, writeError)
}

func TestHeapBufferRetainDelegatesToPool(t *testing.T) {
	poolMock := newHeapPoolMock(t)
	heapBuffer, err := newHeapBuffer(poolMock, 2)
	if !assert.NoError(t, err) {
		return
	}

	retainErr := heapBuffer.Retain()
	assert.NoError(t, retainErr)
	assert.Equal(t, heapBuffer, poolMock.lastRetained)
	assert.Equal(t, 1, poolMock.retainCalls)
}

func TestHeapBufferReleaseDelegatesToPool(t *testing.T) {
	poolMock := newHeapPoolMock(t)
	heapBuffer, err := newHeapBuffer(poolMock, 2)
	if !assert.NoError(t, err) {
		return
	}

	releaseErr := heapBuffer.Release()
	assert.NoError(t, releaseErr)
	assert.Equal(t, heapBuffer, poolMock.lastReleased)
	assert.Equal(t, 1, poolMock.releaseCalls)
}

func TestHeapBufferRetainAndReleasePropagateErrors(t *testing.T) {
	poolMock := newHeapPoolMock(t)
	poolMock.retainErr = errors.New("retain failed")
	poolMock.releaseErr = errors.New("release failed")

	heapBuffer, err := newHeapBuffer(poolMock, 2)
	if !assert.NoError(t, err) {
		return
	}

	retainErr := heapBuffer.Retain()
	assert.Error(t, retainErr)
	assert.ErrorContains(t, retainErr, "retain failed")

	releaseErr := heapBuffer.Release()
	assert.Error(t, releaseErr)
	assert.ErrorContains(t, releaseErr, "release failed")
}

type zeroProgressReader struct{}

func (reader *zeroProgressReader) Read(destination []byte) (int, error) {
	if len(destination) == 0 {
		return 0, io.EOF
	}

	return 0, nil
}

type failingReader struct {
	payload   []byte
	position  int
	targetErr error
}

func (reader *failingReader) Read(destination []byte) (int, error) {
	if reader.position >= len(reader.payload) {
		return 0, reader.targetErr
	}

	bytesCopied := copy(destination, reader.payload[reader.position:])
	reader.position += bytesCopied

	return bytesCopied, reader.targetErr
}

type failingWriter struct {
	targetErr error
}

func (writer *failingWriter) Write(data []byte) (int, error) {
	return 0, writer.targetErr
}

type heapPoolMock struct {
	*memorypkg.PoolMock
	retainCalls  int
	releaseCalls int
	retainErr    error
	releaseErr   error
	lastRetained memorypkg.Buffer
	lastReleased memorypkg.Buffer
}

func newHeapPoolMock(t *testing.T) *heapPoolMock {
	return &heapPoolMock{PoolMock: memorypkg.NewPoolMock(t)}
}

func (mock *heapPoolMock) retain(buffer memorypkg.Buffer) error {
	mock.retainCalls++
	mock.lastRetained = buffer
	return mock.retainErr
}

func (mock *heapPoolMock) release(buffer memorypkg.Buffer) error {
	mock.releaseCalls++
	mock.lastReleased = buffer
	return mock.releaseErr
}

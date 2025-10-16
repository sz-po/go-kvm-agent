package memory

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	memorySDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/memory"
)

type heapBufferPool interface {
	retain(memorySDK.Buffer) error
	release(memorySDK.Buffer) error
}

type HeapBuffer struct {
	pool heapBufferPool

	data     []byte
	dataLock *sync.RWMutex

	writePos int
}

var _ memorySDK.Buffer = (*HeapBuffer)(nil)

func newHeapBuffer(pool heapBufferPool, capacity int) (*HeapBuffer, error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("buffer capacity must be greater than zero")
	}

	heapBuffer := &HeapBuffer{
		pool: pool,

		data:     make([]byte, capacity),
		dataLock: &sync.RWMutex{},
	}

	heapBuffer.reset()

	return heapBuffer, nil
}

func (buffer *HeapBuffer) GetCapacity() int {
	buffer.dataLock.RLock()
	defer buffer.dataLock.RUnlock()

	return cap(buffer.data)
}

func (buffer *HeapBuffer) GetSize() int {
	buffer.dataLock.RLock()
	defer buffer.dataLock.RUnlock()

	return len(buffer.data)
}

func (buffer *HeapBuffer) WriteTo(writer io.Writer) (int64, error) {
	buffer.dataLock.RLock()
	defer buffer.dataLock.RUnlock()

	reader := bytes.NewReader(buffer.data)

	writtenBytes, err := reader.WriteTo(writer)
	if err != nil {
		return writtenBytes, fmt.Errorf("write heap buffer to writer: %w", err)
	}

	return writtenBytes, nil
}

func (buffer *HeapBuffer) Write(bytes []byte) (n int, err error) {
	buffer.dataLock.Lock()
	defer buffer.dataLock.Unlock()

	requiredSize := buffer.writePos + len(bytes)
	if requiredSize > cap(buffer.data) {
		return 0, memorySDK.ErrBufferTooSmall
	}

	if buffer.writePos > len(buffer.data) {
		buffer.data = buffer.data[:buffer.writePos]
	}

	if requiredSize > len(buffer.data) {
		buffer.data = buffer.data[:requiredSize]
	}

	bytesWritten := copy(buffer.data[buffer.writePos:], bytes)

	buffer.writePos += bytesWritten

	return bytesWritten, nil
}

func (buffer *HeapBuffer) ReadFrom(reader io.Reader) (int64, error) {
	buffer.dataLock.Lock()
	defer buffer.dataLock.Unlock()

	capacity := cap(buffer.data)
	if buffer.writePos > capacity {
		return 0, fmt.Errorf("heap buffer write position exceeds capacity")
	}

	if buffer.writePos > len(buffer.data) {
		buffer.data = buffer.data[:buffer.writePos]
	}

	buffer.data = buffer.data[:capacity]

	initialWritePos := buffer.writePos
	totalReadBytes := buffer.writePos
	for totalReadBytes < capacity {
		bytesRead, err := reader.Read(buffer.data[totalReadBytes:])
		if bytesRead > 0 {
			totalReadBytes += bytesRead
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			buffer.data = buffer.data[:totalReadBytes]
			buffer.writePos = totalReadBytes
			return int64(totalReadBytes - initialWritePos), fmt.Errorf("read into heap buffer: %w", err)
		}
		if bytesRead == 0 {
			break
		}
	}

	buffer.data = buffer.data[:totalReadBytes]
	buffer.writePos = totalReadBytes

	return int64(totalReadBytes - initialWritePos), nil
}

func (buffer *HeapBuffer) Retain() error {
	return buffer.pool.retain(buffer)
}

func (buffer *HeapBuffer) Release() error {
	return buffer.pool.release(buffer)
}

func (buffer *HeapBuffer) reset() {
	buffer.dataLock.Lock()
	defer buffer.dataLock.Unlock()

	buffer.writePos = 0
	buffer.data = buffer.data[:0]
}

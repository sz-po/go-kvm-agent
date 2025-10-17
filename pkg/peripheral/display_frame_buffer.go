package peripheral

import (
	"context"
	"errors"
	"io"

	memorySDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/memory"
)

// DisplayFrameBuffer holds buffer with raw display frame data. It wraps memory buffer and holds frame metadata.
type DisplayFrameBuffer struct {
	buffer memorySDK.Buffer
}

func NewDisplayFrameBuffer(buffer memorySDK.Buffer) *DisplayFrameBuffer {
	return &DisplayFrameBuffer{
		buffer: buffer,
	}
}

func (frameBuffer *DisplayFrameBuffer) GetCapacity() int {
	return frameBuffer.buffer.GetCapacity()
}

func (frameBuffer *DisplayFrameBuffer) GetSize() int {
	return frameBuffer.buffer.GetSize()
}

func (frameBuffer *DisplayFrameBuffer) WriteTo(w io.Writer) (n int64, err error) {
	return frameBuffer.buffer.WriteTo(w)
}

func (frameBuffer *DisplayFrameBuffer) Retain() error {
	return frameBuffer.buffer.Retain()
}

func (frameBuffer *DisplayFrameBuffer) Release() error {
	return frameBuffer.buffer.Release()
}

// DisplayFrameBufferProvider provides methods for reading frames.
type DisplayFrameBufferProvider interface {
	// GetDisplayFrameBuffer returns buffer with current frame. Caller is responsible for
	// releasing buffer to the pool. It may return ErrDisplayFrameBufferNotReady if no frame buffer
	// is available.
	GetDisplayFrameBuffer(ctx context.Context) (*DisplayFrameBuffer, error)

	// GetDisplayMode returns current display mode, or error if for some reason is not possible
	// to read current display mode.
	GetDisplayMode(ctx context.Context) (*DisplayMode, error)

	// GetDisplayPixelFormat returns pixel format used in frame buffer.
	GetDisplayPixelFormat(ctx context.Context) DisplayPixelFormat
}

var ErrDisplayFrameBufferNotReady = errors.New("display frame buffer not ready")

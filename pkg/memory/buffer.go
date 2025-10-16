package memory

import (
	"errors"
	"io"
)

// TODO: refine comments

// Buffer holds raw data and provides method for filling it and writing data from it.
type Buffer interface {
	io.ReaderFrom
	io.WriterTo
	io.Writer

	// GetCapacity returns buffer capacity.
	GetCapacity() int

	// GetSize return size of data in buffer.
	GetSize() int

	// Retain increments reference counter. If buffer will be used by more than one reader, it should be
	// retained before passing it.
	Retain() error

	// Release decrements internal reference counter. If counter drop to zero, buffer will be returned to the pool and
	// must not be used again. It also empties buffer data.
	Release() error
}

var ErrBufferAlreadyReleased = errors.New("buffer already released")
var ErrBufferTooSmall = errors.New("buffer too small")

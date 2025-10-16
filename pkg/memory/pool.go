package memory

import "errors"

type Pool interface {
	// Borrow returns empty buffer with reference counter set to 1. Buffer should be released after use.
	Borrow(size int) (Buffer, error)
}

var ErrNoFreeBuffers = errors.New("no free buffers")
var ErrBufferSizeNotSupported = errors.New("buffer size not supported")
var ErrRetainAfterReleaseToPool = errors.New("retain after release to pool")

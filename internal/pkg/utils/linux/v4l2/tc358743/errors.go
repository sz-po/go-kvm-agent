package tc358743

import "errors"

var (
	ErrVideoCaptureNotSupported = errors.New("video capture not supported")
	ErrStreamingNotSupported    = errors.New("streaming not supported")
)

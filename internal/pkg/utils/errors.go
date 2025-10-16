package utils

import (
	"errors"
	"io"
	"net"
	"os"
	"syscall"
)

// IsConnectionClosedError checks if the error indicates a closed network connection.
// It detects common connection closure errors: EOF, ErrClosed, EPIPE, and ECONNRESET.
func IsConnectionClosedError(err error) bool {
	if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
		return true
	}

	var syscallErr *os.SyscallError
	if errors.As(err, &syscallErr) {
		return syscallErr.Err == syscall.EPIPE || syscallErr.Err == syscall.ECONNRESET
	}

	return false
}

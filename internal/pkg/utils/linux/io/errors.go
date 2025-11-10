//go:build linux

package io

import (
	"errors"
	"fmt"
	sys "syscall"
)

var (
	ErrSystem                         = errors.New("system error")
	ErrBadArgument                    = errors.New("bad argument error")
	ErrTemporary                      = errors.New("temporary error")
	ErrTimeout                        = errors.New("timeout error")
	ErrInterrupted                    = errors.New("interrupted")
	ErrDeviceOrResourceBusy           = errors.New("device or resource busy")
	ErrNoEntity                       = errors.New("no entity")
	ErrNoLink                         = errors.New("no link")
	ErrNoBuffersAllocated             = errors.New("no buffers allocated")
	ErrResourceTemporarilyUnavailable = errors.New("resource temporarily unavailable")
)

func parseErrorType(errno sys.Errno) error {
	switch errno {
	case sys.EBADF, sys.ENOMEM, sys.ENODEV, sys.EIO, sys.ENXIO, sys.EFAULT: // structural, terminal
		return fmt.Errorf("%w: %w", ErrSystem, errno)
	case sys.EINTR:
		return ErrInterrupted
	case sys.EINVAL: // bad argument
		return ErrBadArgument
	case sys.EBUSY:
		return ErrDeviceOrResourceBusy
	case sys.ENOENT:
		return ErrNoEntity
	case sys.ENOLINK:
		return ErrNoLink
	default:
		if errno.Timeout() {
			return fmt.Errorf("%w: %w", ErrTimeout, errno)
		}
		if errno.Temporary() {
			return fmt.Errorf("%w: %w", ErrTemporary, errno)
		}
		return fmt.Errorf("%w: %d", errno, errno)
	}
}

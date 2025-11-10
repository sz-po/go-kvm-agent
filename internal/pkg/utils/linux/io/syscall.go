//go:build linux

package io

import (
	"context"
	"errors"

	sys "golang.org/x/sys/unix"
)

// ioctl is a wrapper for Syscall(SYS_IOCTL)
func ioctl(fd, req, arg uintptr) (err sys.Errno) {
	for {
		_, _, errno := sys.Syscall(sys.SYS_IOCTL, fd, req, arg)
		switch errno {
		case 0:
			return 0
		case sys.EINTR:
			continue // retry
		default:
			return errno
		}
	}
}

func SendCtl(fd DeviceDescriptor, req, arg uintptr) error {
	errno := ioctl(uintptr(fd), req, arg)
	if errno == 0 {
		return nil
	}

	return parseErrorType(errno)
}

func WaitForRead(ctx context.Context, descriptor DeviceDescriptor) <-chan struct{} {
	sigChan := make(chan struct{})

	go func(fd uintptr) {
		defer close(sigChan)
		var fdsRead sys.FdSet
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			fdsRead.Zero()
			fdsRead.Set(int(fd))
			// Use shorter timeout for more responsive shutdown
			tv := sys.Timeval{Sec: 0, Usec: 100000} // 100ms
			n, errno := sys.Select(int(fd+1), &fdsRead, nil, nil, &tv)
			if errors.Is(errno, sys.EINTR) {
				continue
			}

			if n == 0 {
				// timeout, no data available
				continue
			}

			select {
			case sigChan <- struct{}{}:
			case <-ctx.Done():
				return
			}
		}
	}(uintptr(descriptor))

	return sigChan
}

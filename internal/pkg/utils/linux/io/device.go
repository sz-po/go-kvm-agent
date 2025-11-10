//go:build linux

package io

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"golang.org/x/sys/unix"
)

type DeviceDescriptor uintptr

var EmptyDeviceDescriptor = DeviceDescriptor(0)

func OpenDevice(devicePath string, flags int, mode uint32) (DeviceDescriptor, error) {
	fstat, err := os.Stat(devicePath)
	if err != nil {
		return 0, fmt.Errorf("open device: %w", err)
	}

	if (fstat.Mode() | fs.ModeCharDevice) == 0 {
		return 0, fmt.Errorf("device open: %s: not character device", devicePath)
	}

	var fd int

	for {
		fd, err = unix.Openat(unix.AT_FDCWD, devicePath, flags, mode)
		if err == nil {
			break
		}

		if errors.Is(err, ErrInterrupted) {
			continue
		}

		return 0, &os.PathError{Op: "open", Path: devicePath, Err: err}
	}
	return DeviceDescriptor(fd), nil
}

func Close(fileDescriptor DeviceDescriptor) error {
	return unix.Close(int(fileDescriptor))
}

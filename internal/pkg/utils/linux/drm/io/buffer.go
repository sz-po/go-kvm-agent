//go:build linux

package io

import (
	linuxio "github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/linux/io"
)

// CreateDumbBuffer will allocate a GEM buffer via DRM_IOCTL_MODE_CREATE_DUMB.
func CreateDumbBuffer(descriptor linuxio.DeviceDescriptor, spec DumbBufferSpec) (DumbBuffer, error) {
	return EmptyDumbBuffer, ErrNotImplemented
}

// DestroyDumbBuffer will free the GEM buffer via DRM_IOCTL_MODE_DESTROY_DUMB.
func DestroyDumbBuffer(descriptor linuxio.DeviceDescriptor, handle GemHandle) error {
	return ErrNotImplemented
}

// MapDumbBuffer will call DRM_IOCTL_MODE_MAP_DUMB and mmap the result.
func MapDumbBuffer(descriptor linuxio.DeviceDescriptor, buffer DumbBuffer) (MappedDumbBuffer, error) {
	return EmptyMappedDumbBuffer, ErrNotImplemented
}

// UnmapDumbBuffer will unmap the buffer memory when the mmap implementation exists.
func UnmapDumbBuffer(mapped MappedDumbBuffer) error {
	return ErrNotImplemented
}

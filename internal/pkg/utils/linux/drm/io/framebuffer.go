//go:build linux

package io

import (
	linuxio "github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/linux/io"
)

// AddFramebuffer will wrap DRM_IOCTL_MODE_ADDFB2 in the future.
func AddFramebuffer(descriptor linuxio.DeviceDescriptor, config FramebufferConfig) (FramebufferID, error) {
	return InvalidFramebufferID, ErrNotImplemented
}

// RemoveFramebuffer will wrap DRM_IOCTL_MODE_RMFB in the future.
func RemoveFramebuffer(descriptor linuxio.DeviceDescriptor, framebufferID FramebufferID) error {
	return ErrNotImplemented
}

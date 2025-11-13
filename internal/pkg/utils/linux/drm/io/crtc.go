//go:build linux

package io

import (
	linuxio "github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/linux/io"
)

// SetCrtc will eventually wrap DRM_IOCTL_MODE_SETCRTC to configure scanout.
func SetCrtc(descriptor linuxio.DeviceDescriptor, config CrtcConfig) error {
	return ErrNotImplemented
}

// PageFlip will issue DRM_IOCTL_MODE_PAGE_FLIP once implemented.
func PageFlip(descriptor linuxio.DeviceDescriptor, request PageFlipRequest) error {
	return ErrNotImplemented
}

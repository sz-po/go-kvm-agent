//go:build linux

package io

import (
	linuxio "github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/linux/io"
)

// ReadEvent will parse data returned by read(2) on the DRM device once implemented.
func ReadEvent(descriptor linuxio.DeviceDescriptor) (Event, error) {
	return EmptyEvent, ErrNotImplemented
}

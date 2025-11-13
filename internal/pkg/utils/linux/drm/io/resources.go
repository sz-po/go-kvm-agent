//go:build linux

package io

import (
	linuxio "github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/linux/io"
)

// GetCardResources will call DRM_IOCTL_MODE_GETRESOURCES once implemented.
// The stub allows higher-level code to be wired before the actual ioctl usage exists.
func GetCardResources(descriptor linuxio.DeviceDescriptor) (CardResources, error) {
	return EmptyCardResources, ErrNotImplemented
}

// GetConnector is expected to wrap DRM_IOCTL_MODE_GETCONNECTOR.
func GetConnector(descriptor linuxio.DeviceDescriptor, connectorID ConnectorID) (Connector, error) {
	return EmptyConnector, ErrNotImplemented
}

// GetEncoder is expected to wrap DRM_IOCTL_MODE_GETENCODER.
func GetEncoder(descriptor linuxio.DeviceDescriptor, encoderID EncoderID) (Encoder, error) {
	return EmptyEncoder, ErrNotImplemented
}

// GetCrtc is expected to wrap DRM_IOCTL_MODE_GETCRTC.
func GetCrtc(descriptor linuxio.DeviceDescriptor, crtcID CrtcID) (Crtc, error) {
	return EmptyCrtc, ErrNotImplemented
}

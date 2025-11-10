//go:build linux

package io

/*
#cgo linux CFLAGS: -I ${SRCDIR}/../include/
#include <linux/videodev2.h>
*/
import "C"

import (
	"unsafe"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/linux/io"
)

const (
	CapVideoCapture uint32 = C.V4L2_CAP_VIDEO_CAPTURE
	CapStreaming    uint32 = C.V4L2_CAP_STREAMING
)

type CapabilityFeatures struct {
	VideoCapture bool `json:"videoCapture"`
	Streaming    bool `json:"streaming"`
}
type Capability struct {
	Driver  string `json:"driver"`
	Card    string `json:"card"`
	BusInfo string `json:"busInfo"`
	Version uint32 `json:"version"`

	Features CapabilityFeatures `json:"features"`
}

var EmptyCapability = Capability{}

// QueryCapabilities queries the capabilities of the device.
// Uses VIDIOC_QUERYCAP ioctl.
func QueryCapabilities(descriptor io.DeviceDescriptor) (Capability, error) {
	var rawCapability C.struct_v4l2_capability

	if err := io.SendCtl(descriptor, C.VIDIOC_QUERYCAP, uintptr(unsafe.Pointer(&rawCapability))); err != nil {
		return EmptyCapability, err
	}

	return Capability{
		Driver:  C.GoString((*C.char)(unsafe.Pointer(&rawCapability.driver[0]))),
		Card:    C.GoString((*C.char)(unsafe.Pointer(&rawCapability.card[0]))),
		BusInfo: C.GoString((*C.char)(unsafe.Pointer(&rawCapability.bus_info[0]))),
		Version: uint32(rawCapability.version),

		Features: CapabilityFeatures{
			VideoCapture: uint32(rawCapability.capabilities)&CapVideoCapture != 0,
			Streaming:    uint32(rawCapability.capabilities)&CapStreaming != 0,
		},
	}, nil
}

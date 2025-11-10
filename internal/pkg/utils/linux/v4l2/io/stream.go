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

// StartStream initiates capture or output operations for the specified buffer type.
// The device begins filling buffers after this call.
// Uses VIDIOC_STREAMON ioctl.
func StartStream(descriptor io.DeviceDescriptor, bufferType BufferType) error {
	rawBufferType := C.__u32(bufferType)
	return io.SendCtl(descriptor, C.VIDIOC_STREAMON, uintptr(unsafe.Pointer(&rawBufferType)))
}

// StopStream halts capture or output operations for the specified buffer type.
// All buffers are removed from incoming and outgoing queues, and unprocessed frames are lost.
// Uses VIDIOC_STREAMOFF ioctl.
func StopStream(descriptor io.DeviceDescriptor, bufferType BufferType) error {
	rawBufferType := C.__u32(bufferType)
	return io.SendCtl(descriptor, C.VIDIOC_STREAMOFF, uintptr(unsafe.Pointer(&rawBufferType)))
}

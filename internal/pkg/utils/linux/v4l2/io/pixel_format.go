//go:build linux

package io

/*
#cgo linux CFLAGS: -I ${SRCDIR}/../include/
#include <linux/videodev2.h>
*/
import "C"

import (
	"encoding/binary"
	"errors"
	"unsafe"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/linux/io"
)

type PixelFormatIndex uint32

type PixelFormatCode uint32

func (code PixelFormatCode) String() string {
	return string(binary.LittleEndian.AppendUint32(nil, uint32(code)))
}

type PixelFormat struct {
	Code        PixelFormatCode `json:"code"`
	Description string          `json:"description"`
}

var EmptyPixelFormat = PixelFormat{}

type PixelFormatList map[PixelFormatIndex]PixelFormat

func (list PixelFormatList) GetIndexByCode(code PixelFormatCode) (*PixelFormatIndex, error) {
	for index, format := range list {
		if format.Code == code {
			return &index, nil
		}
	}
	return nil, errors.New("pixel format not found")
}

var EmptyPixelFormatList = PixelFormatList{}

// ListPixelFormats queries the supported pixel formats of the device.
// Uses VIDIOC_ENUM_FMT ioctl.
func ListPixelFormats(descriptor io.DeviceDescriptor, bufferType BufferType) (PixelFormatList, error) {
	var err error

	result := PixelFormatList{}
	index := PixelFormatIndex(0)
	for {
		var rawPixelFormat C.struct_v4l2_fmtdesc
		rawPixelFormat.index = C.uint(index)
		rawPixelFormat._type = C.uint(bufferType)

		if err = io.SendCtl(descriptor, C.VIDIOC_ENUM_FMT, uintptr(unsafe.Pointer(&rawPixelFormat))); err != nil {
			if errors.Is(err, io.ErrBadArgument) && len(result) > 0 {
				break
			}
			return result, err
		}

		result[index] = PixelFormat{
			Code:        PixelFormatCode(rawPixelFormat.pixelformat),
			Description: C.GoString((*C.char)(unsafe.Pointer(&rawPixelFormat.description[0]))),
		}

		index++
	}
	return result, nil
}

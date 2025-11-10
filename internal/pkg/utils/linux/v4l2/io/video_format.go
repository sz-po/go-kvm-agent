//go:build linux

package io

/*
#cgo linux CFLAGS: -I ${SRCDIR}/../include/
#include <linux/videodev2.h>

static inline struct v4l2_pix_format* format_pix(struct v4l2_format* f) { return &f->fmt.pix; }
*/
import "C"

import (
	"unsafe"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/linux/io"
)

type VideoFormatField uint32

const (
	VideoFormatFieldNone       VideoFormatField = C.V4L2_FIELD_NONE
	VideoFormatFieldInterlaced VideoFormatField = C.V4L2_FIELD_INTERLACED
	VideoFormatFieldSeqTB      VideoFormatField = C.V4L2_FIELD_SEQ_TB
	VideoFormatFieldSeqBT      VideoFormatField = C.V4L2_FIELD_SEQ_BT
)

type VideoFormatColorspace uint32

const (
	VideoFormatColorspaceSRGB = C.V4L2_COLORSPACE_SRGB
)

type VideoFormatQuantization uint32

const (
	VideoFormatQuantizationFullRange = C.V4L2_QUANTIZATION_FULL_RANGE
)

type VideoFormatTransferFunction uint32

const (
	VideoFormatTransferFunctionSRGB = C.V4L2_XFER_FUNC_SRGB
)

type VideoFormatFlag uint32

type VideoFormat struct {
	Width        uint32                      `json:"width"`
	Height       uint32                      `json:"height"`
	PixelFormat  PixelFormatCode             `json:"pixelFormat"`
	Field        VideoFormatField            `json:"field"`
	BytesPerLine uint32                      `json:"bytesPerLine"`
	SizeImage    uint32                      `json:"sizeImage"`
	Colorspace   VideoFormatColorspace       `json:"colorspace"`
	Priv         uint32                      `json:"priv"`
	Flags        VideoFormatFlag             `json:"flags"`
	Quantization VideoFormatQuantization     `json:"quantization"`
	TransferFunc VideoFormatTransferFunction `json:"transferFunc"`
}

var EmptyVideoFormat = VideoFormat{}

// SetVideoFormat sets the video format for the specified buffer type.
// Uses VIDIOC_S_FMT ioctl.
func SetVideoFormat(descriptor io.DeviceDescriptor, bufferType BufferType, format VideoFormat) error {
	var rawFormat C.struct_v4l2_format
	rawFormat._type = C.__u32(bufferType)

	rawFormatPix := C.format_pix(&rawFormat)
	rawFormatPix.width = C.__u32(format.Width)
	rawFormatPix.height = C.__u32(format.Height)
	rawFormatPix.pixelformat = C.__u32(format.PixelFormat)
	rawFormatPix.field = C.__u32(format.Field)
	rawFormatPix.bytesperline = C.__u32(format.BytesPerLine)
	rawFormatPix.sizeimage = C.__u32(format.SizeImage)
	rawFormatPix.colorspace = C.__u32(format.Colorspace)
	rawFormatPix.priv = C.__u32(format.Priv)
	rawFormatPix.flags = C.__u32(format.Flags)
	rawFormatPix.quantization = C.__u32(format.Quantization)
	rawFormatPix.xfer_func = C.__u32(format.TransferFunc)

	return io.SendCtl(descriptor, C.VIDIOC_S_FMT, uintptr(unsafe.Pointer(&rawFormat)))
}

// TryVideoFormat tests if the video format is supported and returns the adjusted format.
// Uses VIDIOC_TRY_FMT ioctl.
func TryVideoFormat(descriptor io.DeviceDescriptor, bufferType BufferType, format VideoFormat) (VideoFormat, error) {
	var rawFormat C.struct_v4l2_format
	rawFormat._type = C.__u32(bufferType)

	rawFormatPix := C.format_pix(&rawFormat)
	rawFormatPix.width = C.__u32(format.Width)
	rawFormatPix.height = C.__u32(format.Height)
	rawFormatPix.pixelformat = C.__u32(format.PixelFormat)
	rawFormatPix.field = C.__u32(format.Field)
	rawFormatPix.bytesperline = C.__u32(format.BytesPerLine)
	rawFormatPix.sizeimage = C.__u32(format.SizeImage)
	rawFormatPix.colorspace = C.__u32(format.Colorspace)
	rawFormatPix.priv = C.__u32(format.Priv)
	rawFormatPix.flags = C.__u32(format.Flags)
	rawFormatPix.quantization = C.__u32(format.Quantization)
	rawFormatPix.xfer_func = C.__u32(format.TransferFunc)

	err := io.SendCtl(descriptor, C.VIDIOC_TRY_FMT, uintptr(unsafe.Pointer(&rawFormat)))
	if err != nil {
		return EmptyVideoFormat, err
	}

	return VideoFormat{
		Width:        uint32(rawFormatPix.width),
		Height:       uint32(rawFormatPix.height),
		PixelFormat:  PixelFormatCode(rawFormatPix.pixelformat),
		Field:        VideoFormatField(rawFormatPix.field),
		BytesPerLine: uint32(rawFormatPix.bytesperline),
		SizeImage:    uint32(rawFormatPix.sizeimage),
		Colorspace:   VideoFormatColorspace(rawFormatPix.colorspace),
		Priv:         uint32(rawFormatPix.priv),
		Flags:        VideoFormatFlag(rawFormatPix.flags),
		Quantization: VideoFormatQuantization(rawFormatPix.quantization),
		TransferFunc: VideoFormatTransferFunction(rawFormatPix.xfer_func),
	}, nil
}

// GetVideoFormat retrieves the current video format for the specified buffer type.
// Uses VIDIOC_G_FMT ioctl.
func GetVideoFormat(descriptor io.DeviceDescriptor, bufferType BufferType) (VideoFormat, error) {
	var rawFormat C.struct_v4l2_format
	rawFormat._type = C.__u32(bufferType)

	err := io.SendCtl(descriptor, C.VIDIOC_G_FMT, uintptr(unsafe.Pointer(&rawFormat)))
	if err != nil {
		return EmptyVideoFormat, err
	}

	rawFormatPix := C.format_pix(&rawFormat)

	return VideoFormat{
		Width:        uint32(rawFormatPix.width),
		Height:       uint32(rawFormatPix.height),
		PixelFormat:  PixelFormatCode(rawFormatPix.pixelformat),
		Field:        VideoFormatField(rawFormatPix.field),
		BytesPerLine: uint32(rawFormatPix.bytesperline),
		SizeImage:    uint32(rawFormatPix.sizeimage),
		Colorspace:   VideoFormatColorspace(rawFormatPix.colorspace),
		Priv:         uint32(rawFormatPix.priv),
		Flags:        VideoFormatFlag(rawFormatPix.flags),
		Quantization: VideoFormatQuantization(rawFormatPix.quantization),
		TransferFunc: VideoFormatTransferFunction(rawFormatPix.xfer_func),
	}, nil
}

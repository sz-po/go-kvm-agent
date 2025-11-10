//go:build linux

package io

/*
#cgo linux CFLAGS: -I ${SRCDIR}/../include/
#include <linux/videodev2.h>

static inline struct v4l2_bt_timings* dv_timings_bt(struct v4l2_dv_timings* t) { return &t->bt; }
*/
import "C"

import (
	"errors"
	"unsafe"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/linux/io"
)

type DigitalVideoTimingType = uint32

const (
	DigitalVideoTimingTypeBT DigitalVideoTimingType = C.V4L2_DV_BT_656_1120
)

type DigitalVideoBTTimingsFlag uint64

type DigitalVideoBTPolaritiesFlag uint32

type DigitalVideoBTStandardsFlag uint32

type DigitalVideoBTTimings struct {
	Width        uint32 `json:"width"`
	Height       uint32 `json:"height"`
	Interlaced   bool   `json:"interlaced"`
	PixelClockHz uint64 `json:"pixelClockHz"`

	HorizontalFrontPorch uint32 `json:"horizontalFrontPorch"`
	HorizontalSync       uint32 `json:"horizontalSync"`
	HorizontalBackPorch  uint32 `json:"horizontalBackPorch"`

	VerticalFrontPorch uint32 `json:"verticalFrontPorch"`
	VerticalSync       uint32 `json:"verticalSync"`
	VerticalBackPorch  uint32 `json:"verticalBackPorch"`

	InterlacedFrontPorch uint32 `json:"interlacedFrontPorch"`
	InterlacedSync       uint32 `json:"interlacedSync"`
	InterlacedBackPorch  uint32 `json:"interlacedBackPorch"`

	Polarities DigitalVideoBTPolaritiesFlag `json:"polarities"`
	Standards  DigitalVideoBTStandardsFlag  `json:"standards"`
	Flags      DigitalVideoBTTimingsFlag    `json:"flags"`
}

var EmptyDigitalVideoBTTimings = DigitalVideoBTTimings{}

func (timings DigitalVideoBTTimings) GetFrameRate() float64 {
	horizontalTotal := uint64(timings.Width) +
		uint64(timings.HorizontalFrontPorch) +
		uint64(timings.HorizontalSync) +
		uint64(timings.HorizontalBackPorch)

	verticalBase := uint64(timings.Height) +
		uint64(timings.VerticalFrontPorch) +
		uint64(timings.VerticalSync) +
		uint64(timings.VerticalBackPorch)

	verticalTotal := verticalBase
	if timings.Interlaced {
		verticalTotal += uint64(
			timings.InterlacedFrontPorch +
				timings.InterlacedSync +
				timings.InterlacedBackPorch,
		)
	}

	if horizontalTotal == 0 || verticalTotal == 0 || timings.PixelClockHz == 0 {
		return 0
	}
	return float64(timings.PixelClockHz) / float64(horizontalTotal*verticalTotal)
}

func SetDigitalVideoBTTimings(descriptor io.DeviceDescriptor, timings DigitalVideoBTTimings) error {
	var rawTimings C.struct_v4l2_dv_timings
	rawTimings._type = C.__u32(DigitalVideoTimingTypeBT)

	rawTimingsBt := C.dv_timings_bt(&rawTimings)
	rawTimingsBt.width = C.__u32(timings.Width)
	rawTimingsBt.height = C.__u32(timings.Height)
	rawTimingsBt.pixelclock = C.__u64(timings.PixelClockHz)

	if timings.Interlaced {
		rawTimingsBt.interlaced = 1
	} else {
		rawTimingsBt.interlaced = 0
	}

	rawTimingsBt.hfrontporch = C.__u32(timings.HorizontalFrontPorch)
	rawTimingsBt.hsync = C.__u32(timings.HorizontalSync)
	rawTimingsBt.hbackporch = C.__u32(timings.HorizontalBackPorch)

	rawTimingsBt.vfrontporch = C.__u32(timings.VerticalFrontPorch)
	rawTimingsBt.vsync = C.__u32(timings.VerticalSync)
	rawTimingsBt.vbackporch = C.__u32(timings.VerticalBackPorch)

	rawTimingsBt.il_vfrontporch = C.__u32(timings.InterlacedFrontPorch)
	rawTimingsBt.il_vsync = C.__u32(timings.InterlacedSync)
	rawTimingsBt.il_vbackporch = C.__u32(timings.InterlacedBackPorch)

	rawTimingsBt.polarities = C.__u32(timings.Polarities)
	rawTimingsBt.flags = C.__u32(timings.Flags)

	return io.SendCtl(descriptor, C.VIDIOC_S_DV_TIMINGS, uintptr(unsafe.Pointer(&rawTimings)))
}

func QueryDigitalVideoBTTimings(descriptor io.DeviceDescriptor) (DigitalVideoBTTimings, error) {
	var rawTimings C.struct_v4l2_dv_timings

	err := io.SendCtl(descriptor, C.VIDIOC_QUERY_DV_TIMINGS, uintptr(unsafe.Pointer(&rawTimings)))
	if err != nil {
		return EmptyDigitalVideoBTTimings, err
	}

	if DigitalVideoTimingType(rawTimings._type) != DigitalVideoTimingTypeBT {
		return EmptyDigitalVideoBTTimings, ErrDigitalVideoTimingsAreNotBT
	}

	rawTimingsBt := C.dv_timings_bt(&rawTimings)

	return DigitalVideoBTTimings{
		Width:        uint32(rawTimingsBt.width),
		Height:       uint32(rawTimingsBt.height),
		Interlaced:   rawTimingsBt.interlaced != 0,
		PixelClockHz: uint64(rawTimingsBt.pixelclock),

		HorizontalFrontPorch: uint32(rawTimingsBt.hfrontporch),
		HorizontalSync:       uint32(rawTimingsBt.hsync),
		HorizontalBackPorch:  uint32(rawTimingsBt.hbackporch),

		VerticalFrontPorch: uint32(rawTimingsBt.vfrontporch),
		VerticalSync:       uint32(rawTimingsBt.vsync),
		VerticalBackPorch:  uint32(rawTimingsBt.vbackporch),

		InterlacedFrontPorch: uint32(rawTimingsBt.il_vfrontporch),
		InterlacedSync:       uint32(rawTimingsBt.il_vsync),
		InterlacedBackPorch:  uint32(rawTimingsBt.il_vbackporch),

		Polarities: DigitalVideoBTPolaritiesFlag(rawTimingsBt.polarities),
		Standards:  DigitalVideoBTStandardsFlag(rawTimingsBt.standards),
		Flags:      DigitalVideoBTTimingsFlag(rawTimingsBt.flags),
	}, nil
}

var ErrDigitalVideoTimingsAreNotBT = errors.New("digital video timings are not BT")

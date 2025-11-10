//go:build linux

package io

/*
#cgo linux CFLAGS: -I ${SRCDIR}/../include/
#include <linux/videodev2.h>

static inline __u32 buffer_mmap_offset(struct v4l2_buffer* b) { return b->m.offset; }
*/
import "C"

import (
	"unsafe"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/linux/io"
	"golang.org/x/sys/unix"
)

type BufferIndex = uint32

type MemoryType = uint32

const (
	MemoryTypeMmap MemoryType = C.V4L2_MEMORY_MMAP
)

type BufferType = uint32

const (
	BufferTypeVideoCapture BufferType = C.V4L2_BUF_TYPE_VIDEO_CAPTURE
)

type BufferFlag uint32

const (
	BufferFlagMapped   BufferFlag = C.V4L2_BUF_FLAG_MAPPED
	BufferFlagQueued   BufferFlag = C.V4L2_BUF_FLAG_QUEUED
	BufferFlagDone     BufferFlag = C.V4L2_BUF_FLAG_DONE
	BufferFlagError    BufferFlag = C.V4L2_BUF_FLAG_ERROR
	BufferFlagPrepared BufferFlag = C.V4L2_BUF_FLAG_PREPARED
)

// BufferDescriptor contains buffer metadata filled by the kernel during dequeue operations.
type BufferDescriptor struct {
	Type      BufferType       `json:"type"`
	Memory    MemoryType       `json:"memory"`
	Index     BufferIndex      `json:"index"`
	BytesUsed uint32           `json:"bytesUsed"`
	Flags     BufferFlag       `json:"flags"`
	Sequence  uint32           `json:"sequence"`
	Field     VideoFormatField `json:"field"`
}

var EmptyBufferDescriptor = BufferDescriptor{}

type Buffer struct {
	BufferDescriptor
	Length uint32 `json:"length"`
}

type MmapBuffer struct {
	Buffer
	Offset uint32 `json:"offset"`
}

var EmptyMmapBuffer = MmapBuffer{}

// BoundMmapBuffer represents an MMAP buffer that has been mapped into process memory.
// The Data field provides direct access to the buffer memory for reading and writing.
// After calling UnbindMmapBuffer, the Data field should not be accessed.
type BoundMmapBuffer struct {
	MmapBuffer
	Data       []byte `json:"-"`
	descriptor io.DeviceDescriptor
}

var EmptyBoundMmapBuffer = BoundMmapBuffer{}

type BoundMmapBuffers map[BufferIndex]BoundMmapBuffer

var EmptyBoundMmapBuffers = BoundMmapBuffers{}

// RequestBuffers requests buffer allocation from the kernel for memory-mapped streaming I/O.
// Returns the actual number of buffers allocated by the kernel, which may be less than requested.
// Uses VIDIOC_REQBUFS ioctl.
func RequestBuffers(descriptor io.DeviceDescriptor, bufferType BufferType, memoryType MemoryType, count uint32) (uint32, error) {
	if count == 0 {
		return 0, io.ErrBadArgument
	}

	var rawRequestBuffers C.struct_v4l2_requestbuffers
	rawRequestBuffers._type = C.__u32(bufferType)
	rawRequestBuffers.memory = C.__u32(memoryType)
	rawRequestBuffers.count = C.__u32(count)

	if err := io.SendCtl(descriptor, C.VIDIOC_REQBUFS, uintptr(unsafe.Pointer(&rawRequestBuffers))); err != nil {
		return 0, err
	}

	allocatedCount := uint32(rawRequestBuffers.count)
	if allocatedCount == 0 {
		return 0, io.ErrNoBuffersAllocated
	}

	return allocatedCount, nil
}

func ReleaseBuffers(descriptor io.DeviceDescriptor, bufferType BufferType, memoryType MemoryType) error {
	_, err := RequestBuffers(descriptor, bufferType, memoryType, 0)
	return err
}

// QueryMmapBuffer retrieves information about an MMAP buffer.
// Returns buffer metadata including memory offset, length, and status flags.
// Uses VIDIOC_QUERYBUF ioctl with V4L2_MEMORY_MMAP.
func QueryMmapBuffer(descriptor io.DeviceDescriptor, bufferType BufferType, index BufferIndex) (MmapBuffer, error) {
	var rawBuffer C.struct_v4l2_buffer
	rawBuffer._type = C.__u32(bufferType)
	rawBuffer.memory = C.__u32(MemoryTypeMmap)
	rawBuffer.index = C.__u32(index)

	if err := io.SendCtl(descriptor, C.VIDIOC_QUERYBUF, uintptr(unsafe.Pointer(&rawBuffer))); err != nil {
		return EmptyMmapBuffer, err
	}

	return MmapBuffer{
		Buffer: Buffer{
			BufferDescriptor: BufferDescriptor{
				Index:     BufferIndex(rawBuffer.index),
				BytesUsed: uint32(rawBuffer.bytesused),
				Flags:     BufferFlag(rawBuffer.flags),
				Sequence:  uint32(rawBuffer.sequence),
				Field:     VideoFormatField(rawBuffer.field),
				Type:      BufferType(rawBuffer._type),
				Memory:    MemoryType(rawBuffer.memory),
			},

			Length: uint32(rawBuffer.length),
		},
		Offset: uint32(C.buffer_mmap_offset(&rawBuffer)),
	}, nil
}

// BindMmapBuffer maps the MMAP buffer into process memory for direct access.
// Returns a BoundMmapBuffer with a Data slice that can be used to read/write buffer contents.
// The buffer must be unmapped using UnbindMmapBuffer when done.
// Uses mmap with PROT_READ|PROT_WRITE and MAP_SHARED flags.
func BindMmapBuffer(descriptor io.DeviceDescriptor, buffer MmapBuffer) (BoundMmapBuffer, error) {
	data, err := unix.Mmap(
		int(descriptor),
		int64(buffer.Offset),
		int(buffer.Length),
		unix.PROT_READ|unix.PROT_WRITE,
		unix.MAP_SHARED,
	)
	if err != nil {
		return EmptyBoundMmapBuffer, err
	}

	return BoundMmapBuffer{
		MmapBuffer: buffer,
		Data:       data,
		descriptor: descriptor,
	}, nil
}

// UnbindMmapBuffer unmaps the buffer from process memory.
// After calling this function, the Data field must not be accessed.
// Returns an error if the buffer was already unbound or never bound.
func UnbindMmapBuffer(buffer BoundMmapBuffer) error {
	if err := unix.Munmap(buffer.Data); err != nil {
		return err
	}

	buffer.Data = nil
	return nil
}

// QueueBuffer enqueues a buffer to the kernel for capture.
// The buffer will be filled by the hardware and can be retrieved using DequeueMmapBuffer.
// Uses VIDIOC_QBUF ioctl.
func QueueBuffer(descriptor io.DeviceDescriptor, bufferDescriptor BufferDescriptor) error {
	var rawBuffer C.struct_v4l2_buffer
	rawBuffer._type = C.__u32(bufferDescriptor.Type)
	rawBuffer.memory = C.__u32(bufferDescriptor.Memory)
	rawBuffer.index = C.__u32(bufferDescriptor.Index)

	return io.SendCtl(descriptor, C.VIDIOC_QBUF, uintptr(unsafe.Pointer(&rawBuffer)))
}

// DequeueMmapBuffer dequeues an MMAP buffer from the kernel after it has been filled by hardware.
// Returns a BufferDescriptor containing buffer metadata including which buffer was filled (Index),
// how much data is valid (BytesUsed), buffer status (Flags), and frame sequence number.
// Uses VIDIOC_DQBUF ioctl with V4L2_MEMORY_MMAP.
func DequeueMmapBuffer(descriptor io.DeviceDescriptor, bufferType BufferType) (BufferDescriptor, error) {
	var rawBuffer C.struct_v4l2_buffer
	rawBuffer._type = C.__u32(bufferType)
	rawBuffer.memory = C.__u32(MemoryTypeMmap)

	if err := io.SendCtl(descriptor, C.VIDIOC_DQBUF, uintptr(unsafe.Pointer(&rawBuffer))); err != nil {
		return EmptyBufferDescriptor, err
	}

	return BufferDescriptor{
		Type:      BufferType(rawBuffer._type),
		Memory:    MemoryType(rawBuffer.memory),
		Index:     BufferIndex(rawBuffer.index),
		BytesUsed: uint32(rawBuffer.bytesused),
		Flags:     BufferFlag(rawBuffer.flags),
		Sequence:  uint32(rawBuffer.sequence),
		Field:     VideoFormatField(rawBuffer.field),
	}, nil
}

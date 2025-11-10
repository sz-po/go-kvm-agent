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

type EventType uint32

const (
	EventTypeAll          EventType = C.V4L2_EVENT_ALL
	EventTypeSourceChange           = C.V4L2_EVENT_SOURCE_CHANGE
)

type Event struct {
	Type EventType
}

var EmptyEvent = Event{}

func SubscribeEvent(descriptor io.DeviceDescriptor, eventType EventType) error {
	var rawEventSubscription C.struct_v4l2_event_subscription

	rawEventSubscription._type = C.__u32(eventType)

	err := io.SendCtl(descriptor, C.VIDIOC_SUBSCRIBE_EVENT, uintptr(unsafe.Pointer(&rawEventSubscription)))
	if err != nil {
		return err
	}

	return nil
}

func UnsubscribeEvent(descriptor io.DeviceDescriptor, eventType EventType) error {
	var rawEventSubscription C.struct_v4l2_event_subscription

	rawEventSubscription._type = C.__u32(eventType)

	err := io.SendCtl(descriptor, C.VIDIOC_SUBSCRIBE_EVENT, uintptr(unsafe.Pointer(&rawEventSubscription)))
	if err != nil {
		return err
	}

	return nil
}

func DequeueEvent(descriptor io.DeviceDescriptor) (Event, error) {
	var rawEvent C.struct_v4l2_event

	err := io.SendCtl(descriptor, C.VIDIOC_DQEVENT, uintptr(unsafe.Pointer(&rawEvent)))
	if err != nil {
		return EmptyEvent, err
	}

	return Event{
		Type: EventType(rawEvent._type),
	}, nil
}

//go:build linux

package io

import (
	"context"

	"golang.org/x/sys/unix"
)

type PollEvent int16

const (
	PollEventInput    PollEvent = unix.POLLIN
	PollEventPriority           = unix.POLLPRI
)

func Poll(ctx context.Context, descriptor DeviceDescriptor, events ...PollEvent) <-chan PollEvent {
	done := ctx.Done()

	output := make(chan PollEvent, 64)

	var pollDescriptorEvents int16
	for _, event := range events {
		pollDescriptorEvents = pollDescriptorEvents | int16(event)
	}

	go func() {
		defer func() {
			close(output)
		}()
		for {
			select {
			case <-done:
				return
			default:
				pollDescriptor := []unix.PollFd{
					{
						Fd:     int32(descriptor),
						Events: pollDescriptorEvents,
					},
				}

				eventCount, err := unix.Poll(pollDescriptor, 100)
				if err != nil {
					continue
				}
				if eventCount > 0 {
					event := pollDescriptor[0].Revents
					if event&int16(PollEventInput) != 0 {
						output <- PollEventInput
					}
					if event&int16(PollEventPriority) != 0 {
						output <- PollEventPriority
					}
				}
			}
		}
	}()

	return output
}

//go:build linux

package io

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"
)

func TestPoll_InputEvent(t *testing.T) {
	descriptors := make([]int, 2)
	err := unix.Pipe(descriptors)
	assert.NoError(t, err)
	readDescriptor := descriptors[0]
	writeDescriptor := descriptors[1]
	defer unix.Close(readDescriptor)
	defer unix.Close(writeDescriptor)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventChannel := Poll(ctx, DeviceDescriptor(readDescriptor), PollEventInput)

	_, err = unix.Write(writeDescriptor, []byte("test data"))
	assert.NoError(t, err)

	select {
	case event := <-eventChannel:
		assert.Equal(t, PollEventInput, event)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for PollEventInput")
	}
}

func TestPoll_ContextCancellation(t *testing.T) {
	descriptors := make([]int, 2)
	err := unix.Pipe(descriptors)
	assert.NoError(t, err)
	readDescriptor := descriptors[0]
	writeDescriptor := descriptors[1]
	defer unix.Close(readDescriptor)
	defer unix.Close(writeDescriptor)

	ctx, cancel := context.WithCancel(context.Background())

	eventChannel := Poll(ctx, DeviceDescriptor(readDescriptor), PollEventInput)

	cancel()

	select {
	case _, ok := <-eventChannel:
		assert.False(t, ok, "Channel should be closed after context cancellation")
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for channel to close")
	}
}

func TestPoll_MultipleEvents(t *testing.T) {
	descriptors := make([]int, 2)
	err := unix.Pipe(descriptors)
	assert.NoError(t, err)
	readDescriptor := descriptors[0]
	writeDescriptor := descriptors[1]
	defer unix.Close(readDescriptor)
	defer unix.Close(writeDescriptor)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventChannel := Poll(ctx, DeviceDescriptor(readDescriptor), PollEventInput)

	_, err = unix.Write(writeDescriptor, []byte("first"))
	assert.NoError(t, err)

	event := <-eventChannel
	assert.Equal(t, PollEventInput, event)

	buffer := make([]byte, 5)
	_, err = unix.Read(readDescriptor, buffer)
	assert.NoError(t, err)

	_, err = unix.Write(writeDescriptor, []byte("second"))
	assert.NoError(t, err)

	event = <-eventChannel
	assert.Equal(t, PollEventInput, event)
}

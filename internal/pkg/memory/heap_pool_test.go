package memory

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	memorypkg "github.com/szymonpodeszwa/go-kvm-agent/pkg/memory"
)

func TestNewHeapPoolValidatesCapacities(t *testing.T) {
	_, err := NewHeapPool(16, 0)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "pool capacity must be greater than zero")

	_, err = NewHeapPool(0, 1)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "buffer capacity must be greater than zero")
}

func TestNewHeapPoolInitializesReferences(t *testing.T) {
	memoryPool, err := NewHeapPool(8, 3)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, 3, len(memoryPool.references))

	for _, references := range memoryPool.references {
		assert.Equal(t, 0, references)
	}
}

func TestHeapPoolBorrowHonorsCapacityAndLimits(t *testing.T) {
	memoryPool, err := NewHeapPool(4, 2)
	if !assert.NoError(t, err) {
		return
	}

	firstBuffer, borrowErr := memoryPool.Borrow(3)
	if !assert.NoError(t, borrowErr) {
		return
	}

	typedFirstBuffer, isHeapBuffer := firstBuffer.(*HeapBuffer)
	if !assert.True(t, isHeapBuffer) {
		return
	}
	assert.Equal(t, 1, memoryPool.references[typedFirstBuffer])

	secondBuffer, secondBorrowErr := memoryPool.Borrow(1)
	if !assert.NoError(t, secondBorrowErr) {
		return
	}

	typedSecondBuffer, isHeapBuffer := secondBuffer.(*HeapBuffer)
	if !assert.True(t, isHeapBuffer) {
		return
	}
	assert.Equal(t, 1, memoryPool.references[typedSecondBuffer])

	_, thirdBorrowErr := memoryPool.Borrow(1)
	assert.ErrorIs(t, thirdBorrowErr, memorypkg.ErrNoFreeBuffers)
}

func TestHeapPoolBorrowRejectsOversizedRequest(t *testing.T) {
	memoryPool, err := NewHeapPool(4, 1)
	if !assert.NoError(t, err) {
		return
	}

	buffer, borrowErr := memoryPool.Borrow(8)
	assert.Nil(t, buffer)
	assert.ErrorIs(t, borrowErr, memorypkg.ErrBufferSizeNotSupported)

	for _, references := range memoryPool.references {
		assert.Equal(t, 0, references)
	}
}

func TestHeapPoolRetainAndReleaseLifecycle(t *testing.T) {
	memoryPool, err := NewHeapPool(4, 1)
	if !assert.NoError(t, err) {
		return
	}

	buffer, borrowErr := memoryPool.Borrow(4)
	if !assert.NoError(t, borrowErr) {
		return
	}

	typedBuffer, isHeapBuffer := buffer.(*HeapBuffer)
	if !assert.True(t, isHeapBuffer) {
		return
	}

	_, readErr := buffer.ReadFrom(bytes.NewReader([]byte("data")))
	assert.NoError(t, readErr)
	assert.Equal(t, 4, typedBuffer.GetSize())

	retainErr := buffer.Retain()
	assert.NoError(t, retainErr)
	assert.Equal(t, 2, memoryPool.references[typedBuffer])

	releaseErr := buffer.Release()
	assert.NoError(t, releaseErr)
	assert.Equal(t, 1, memoryPool.references[typedBuffer])

	releaseErr = buffer.Release()
	assert.NoError(t, releaseErr)
	assert.Equal(t, 0, memoryPool.references[typedBuffer])
	assert.Equal(t, 0, typedBuffer.GetSize())

	extraReleaseErr := buffer.Release()
	assert.ErrorIs(t, extraReleaseErr, memorypkg.ErrBufferAlreadyReleased)
	assert.Equal(t, 0, memoryPool.references[typedBuffer])

	retainAfterReleaseErr := buffer.Retain()
	assert.ErrorIs(t, retainAfterReleaseErr, memorypkg.ErrRetainAfterReleaseToPool)
	assert.Equal(t, 0, memoryPool.references[typedBuffer])
}

func TestHeapPoolRetainRejectsForeignBuffer(t *testing.T) {
	memoryPool, err := NewHeapPool(4, 1)
	if !assert.NoError(t, err) {
		return
	}

	foreignBuffer := memorypkg.NewBufferMock(t)

	retainErr := memoryPool.retain(foreignBuffer)
	assert.ErrorIs(t, retainErr, ErrNotHeapBuffer)
}

func TestHeapPoolRetainRejectsUnknownHeapBuffer(t *testing.T) {
	memoryPool, err := NewHeapPool(4, 1)
	if !assert.NoError(t, err) {
		return
	}

	unmanagedBuffer, bufferErr := newHeapBuffer(memoryPool, 4)
	if !assert.NoError(t, bufferErr) {
		return
	}

	retainErr := memoryPool.retain(unmanagedBuffer)
	assert.ErrorIs(t, retainErr, ErrBufferNotManagedByPool)
}

func TestHeapPoolReleaseRejectsForeignBuffer(t *testing.T) {
	memoryPool, err := NewHeapPool(4, 1)
	if !assert.NoError(t, err) {
		return
	}

	foreignBuffer := memorypkg.NewBufferMock(t)

	releaseErr := memoryPool.release(foreignBuffer)
	assert.ErrorIs(t, releaseErr, ErrNotHeapBuffer)
}

func TestHeapPoolReleaseRejectsUnknownHeapBuffer(t *testing.T) {
	memoryPool, err := NewHeapPool(4, 1)
	if !assert.NoError(t, err) {
		return
	}

	unmanagedBuffer, bufferErr := newHeapBuffer(memoryPool, 4)
	if !assert.NoError(t, bufferErr) {
		return
	}

	releaseErr := memoryPool.release(unmanagedBuffer)
	assert.ErrorIs(t, releaseErr, ErrBufferNotManagedByPool)
}

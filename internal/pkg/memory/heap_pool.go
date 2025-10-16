package memory

import (
	"errors"
	"fmt"
	"sync"

	memorySDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/memory"
)

type HeapPool struct {
	references     map[*HeapBuffer]int
	referencesLock sync.RWMutex

	bufferCapacity int
	poolCapacity   int
}

func NewHeapPool(bufferCapacity int, poolCapacity int) (*HeapPool, error) {
	if poolCapacity <= 0 {
		return nil, errors.New("pool capacity must be greater than zero")
	}
	
	pool := &HeapPool{
		references:     make(map[*HeapBuffer]int, poolCapacity),
		referencesLock: sync.RWMutex{},

		bufferCapacity: bufferCapacity,
		poolCapacity:   poolCapacity,
	}

	for bufferIndex := 0; bufferIndex < poolCapacity; bufferIndex++ {
		buffer, err := newHeapBuffer(pool, bufferCapacity)
		if err != nil {
			return nil, fmt.Errorf("error creating buffer %d: %w", bufferIndex, err)
		}

		pool.references[buffer] = 0
	}

	return pool, nil
}

var _ memorySDK.Pool = (*HeapPool)(nil)

// Borrow returns an idle frame buffer and marks it as in use.
func (pool *HeapPool) Borrow(size int) (memorySDK.Buffer, error) {
	pool.referencesLock.Lock()
	defer pool.referencesLock.Unlock()

	if size > pool.bufferCapacity {
		return nil, memorySDK.ErrBufferSizeNotSupported
	}

	var heapBuffer *HeapBuffer

	for poolBuffer, references := range pool.references {
		if references == 0 {
			heapBuffer = poolBuffer
			break
		}
	}

	if heapBuffer == nil {
		return nil, memorySDK.ErrNoFreeBuffers
	}

	pool.references[heapBuffer] = 1

	return heapBuffer, nil
}

// Retain increments the reference count for a managed frame buffer.
func (pool *HeapPool) retain(buffer memorySDK.Buffer) error {
	pool.referencesLock.Lock()
	defer pool.referencesLock.Unlock()

	heapBuffer, isHeapBuffer := buffer.(*HeapBuffer)
	if !isHeapBuffer {
		return ErrNotHeapBuffer
	}

	references, exists := pool.references[heapBuffer]
	if !exists {
		return ErrBufferNotManagedByPool
	}

	if references == 0 {
		return memorySDK.ErrRetainAfterReleaseToPool
	}

	pool.references[heapBuffer] = references + 1

	return nil
}

// release decrements the reference count and makes the buffer idle when it reaches zero.
func (pool *HeapPool) release(buffer memorySDK.Buffer) error {
	pool.referencesLock.Lock()
	defer pool.referencesLock.Unlock()

	heapBuffer, isHeapBuffer := buffer.(*HeapBuffer)
	if !isHeapBuffer {
		return ErrNotHeapBuffer
	}

	references, exists := pool.references[heapBuffer]
	if !exists {
		return ErrBufferNotManagedByPool
	}

	if references == 0 {
		return memorySDK.ErrBufferAlreadyReleased
	}

	updatedReferences := references - 1
	pool.references[heapBuffer] = updatedReferences

	if updatedReferences == 0 {
		heapBuffer.reset()
	}

	return nil
}

var ErrNotHeapBuffer = errors.New("not heap buffer")
var ErrBufferNotManagedByPool = errors.New("buffer not managed by pool")

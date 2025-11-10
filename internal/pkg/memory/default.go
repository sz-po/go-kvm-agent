package memory

import (
	"errors"
	"sync"

	memorySDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/memory"
)

var defaultMemoryPool memorySDK.Pool
var defaultMemoryPoolLock sync.RWMutex

func DefaultMemoryPoolProvider() (memorySDK.Pool, error) {
	return GetDefaultMemoryPool()
}

func GetDefaultMemoryPool() (memorySDK.Pool, error) {
	defaultMemoryPoolLock.RLock()
	defer defaultMemoryPoolLock.RUnlock()

	if defaultMemoryPool == nil {
		return nil, ErrNoDefaultMemoryPool
	}

	return defaultMemoryPool, nil
}

func SetDefaultMemoryPool(pool memorySDK.Pool) error {
	defaultMemoryPoolLock.Lock()
	defer defaultMemoryPoolLock.Unlock()

	if pool == nil {
		return ErrInvalidDefaultMemoryPool
	}

	if defaultMemoryPool != nil {
		return ErrDefaultMemoryPoolAlreadySet
	}

	defaultMemoryPool = pool

	return nil
}

var ErrInvalidDefaultMemoryPool = errors.New("invalid default memory pool")
var ErrNoDefaultMemoryPool = errors.New("no default memory pool")
var ErrDefaultMemoryPoolAlreadySet = errors.New("default memory pool already set")

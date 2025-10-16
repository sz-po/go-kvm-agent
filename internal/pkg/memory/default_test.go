package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	memorypkg "github.com/szymonpodeszwa/go-kvm-agent/pkg/memory"
)

func resetDefaultPool(t *testing.T) {
	t.Helper()

	defaultMemoryPoolLock.Lock()
	defaultMemoryPool = nil
	defaultMemoryPoolLock.Unlock()
}

func TestGetDefaultMemoryPoolWithoutInitialization(t *testing.T) {
	resetDefaultPool(t)

	pool, err := GetDefaultMemoryPool()
	assert.Nil(t, pool)
	assert.ErrorIs(t, err, ErrNoDefaultMemoryPool)
}

func TestSetDefaultMemoryPoolRejectsNil(t *testing.T) {
	resetDefaultPool(t)

	err := SetDefaultMemoryPool(nil)
	assert.ErrorIs(t, err, ErrInvalidDefaultMemoryPool)

	pool, getErr := GetDefaultMemoryPool()
	assert.Nil(t, pool)
	assert.ErrorIs(t, getErr, ErrNoDefaultMemoryPool)
}

func TestSetAndGetDefaultMemoryPool(t *testing.T) {
	resetDefaultPool(t)

	poolMock := memorypkg.NewPoolMock(t)

	setErr := SetDefaultMemoryPool(poolMock)
	if !assert.NoError(t, setErr) {
		return
	}

	retrievedPool, getErr := GetDefaultMemoryPool()
	if !assert.NoError(t, getErr) {
		return
	}

	assert.Equal(t, poolMock, retrievedPool)
}

func TestSetDefaultMemoryPoolOnlyOnce(t *testing.T) {
	resetDefaultPool(t)

	firstPool := memorypkg.NewPoolMock(t)
	secondPool := memorypkg.NewPoolMock(t)

	firstSetErr := SetDefaultMemoryPool(firstPool)
	if !assert.NoError(t, firstSetErr) {
		return
	}

	secondSetErr := SetDefaultMemoryPool(secondPool)
	assert.ErrorIs(t, secondSetErr, ErrDefaultMemoryPoolAlreadySet)

	retrievedPool, getErr := GetDefaultMemoryPool()
	if !assert.NoError(t, getErr) {
		return
	}

	assert.Equal(t, firstPool, retrievedPool)
}

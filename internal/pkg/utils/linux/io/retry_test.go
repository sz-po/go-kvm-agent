//go:build linux

package io

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	errRetryable    = errors.New("retryable error")
	errNonRetryable = errors.New("non-retryable error")
)

func TestRetryOnErrorWithValue_Success(t *testing.T) {
	ctx := context.Background()
	expectedValue := 42

	fn := func() (int, error) {
		return expectedValue, nil
	}

	value, err := RetryOnErrorWithValue(ctx, fn, errRetryable)

	assert.NoError(t, err)
	assert.Equal(t, expectedValue, value)
}

func TestRetryOnErrorWithValue_RetryableErrorThenSuccess(t *testing.T) {
	ctx := context.Background()
	expectedValue := "success"
	attempts := 0

	fn := func() (string, error) {
		attempts++
		if attempts < 3 {
			return "", errRetryable
		}
		return expectedValue, nil
	}

	value, err := RetryOnErrorWithValue(ctx, fn, errRetryable)

	assert.NoError(t, err)
	assert.Equal(t, expectedValue, value)
	assert.Equal(t, 3, attempts)
}

func TestRetryOnErrorWithValue_NonRetryableError(t *testing.T) {
	ctx := context.Background()

	fn := func() (int, error) {
		return 0, errNonRetryable
	}

	value, err := RetryOnErrorWithValue(ctx, fn, errRetryable)

	assert.Error(t, err)
	assert.ErrorIs(t, err, errNonRetryable)
	assert.Contains(t, err.Error(), "non retryable")
	assert.Equal(t, 0, value)
}

func TestRetryOnErrorWithValue_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fn := func() (string, error) {
		return "never reached", errRetryable
	}

	value, err := RetryOnErrorWithValue(ctx, fn, errRetryable)

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, "", value)
}

func TestRetryOnErrorWithValue_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	fn := func() (int, error) {
		return 0, errRetryable
	}

	value, err := RetryOnErrorWithValue(ctx, fn, errRetryable)

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Equal(t, 0, value)
}

func TestRetryOnErrorWithValue_MultipleRetryableErrors(t *testing.T) {
	ctx := context.Background()
	errRetryable1 := errors.New("retryable error 1")
	errRetryable2 := errors.New("retryable error 2")
	expectedValue := "success"
	attempts := 0

	fn := func() (string, error) {
		attempts++
		switch attempts {
		case 1:
			return "", errRetryable1
		case 2:
			return "", errRetryable2
		default:
			return expectedValue, nil
		}
	}

	value, err := RetryOnErrorWithValue(ctx, fn, errRetryable1, errRetryable2)

	assert.NoError(t, err)
	assert.Equal(t, expectedValue, value)
	assert.Equal(t, 3, attempts)
}

func TestRetryOnErrorWithValue_ErrorNotInRetryableList(t *testing.T) {
	ctx := context.Background()
	errOther := errors.New("other error")

	fn := func() (int, error) {
		return 0, errOther
	}

	value, err := RetryOnErrorWithValue(ctx, fn, errRetryable)

	assert.Error(t, err)
	assert.ErrorIs(t, err, errOther)
	assert.Contains(t, err.Error(), "non retryable")
	assert.Equal(t, 0, value)
}

func TestRetryOnErrorWithValue_WrappedRetryableError(t *testing.T) {
	ctx := context.Background()
	wrappedErr := errors.Join(errRetryable, errors.New("additional context"))
	expectedValue := "success"
	attempts := 0

	fn := func() (string, error) {
		attempts++
		if attempts < 2 {
			return "", wrappedErr
		}
		return expectedValue, nil
	}

	value, err := RetryOnErrorWithValue(ctx, fn, errRetryable)

	assert.NoError(t, err)
	assert.Equal(t, expectedValue, value)
	assert.Equal(t, 2, attempts)
}

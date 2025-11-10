//go:build linux

package io

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func RetryOnErrorWithValue[T any](ctx context.Context, fn func() (T, error), retryableErrors ...error) (T, error) {
	var latestErr error
	retries := 0

	for {
		if err := ctx.Err(); err != nil {
			var zero T
			if latestErr != nil {
				err = fmt.Errorf("%w: %w", latestErr, err)
			}
			return zero, fmt.Errorf("retried %d times: %w", retries, err)
		}

		value, fnErr := fn()
		if fnErr == nil {
			return value, nil
		}

		latestErr = fnErr

		isRetryable := false
		for _, retryableErr := range retryableErrors {
			if errors.Is(fnErr, retryableErr) {
				isRetryable = true
				break
			}
		}

		if isRetryable {
			delay := min(1<<retries, 500)
			time.Sleep(time.Millisecond * time.Duration(delay))
			retries++
			continue
		}

		var err error
		if latestErr != nil {
			err = fmt.Errorf("%w: %w", latestErr, fnErr)
		} else {
			err = fnErr
		}

		return value, fmt.Errorf("retries: %d: non retryable: %w", retries, err)
	}
}

func RetryOnError(ctx context.Context, fn func() error, retryableErrors ...error) error {
	_, err := RetryOnErrorWithValue(ctx, func() (struct{}, error) {
		return struct{}{}, fn()
	}, retryableErrors...)
	return err
}

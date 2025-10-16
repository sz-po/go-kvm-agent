package utils

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsConnectionClosedError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error returns false",
			err:      nil,
			expected: false,
		},
		{
			name:     "io.EOF returns true",
			err:      io.EOF,
			expected: true,
		},
		{
			name:     "wrapped io.EOF returns true",
			err:      fmt.Errorf("connection error: %w", io.EOF),
			expected: true,
		},
		{
			name:     "net.ErrClosed returns true",
			err:      net.ErrClosed,
			expected: true,
		},
		{
			name:     "wrapped net.ErrClosed returns true",
			err:      fmt.Errorf("write failed: %w", net.ErrClosed),
			expected: true,
		},
		{
			name:     "syscall EPIPE returns true",
			err:      &os.SyscallError{Err: syscall.EPIPE},
			expected: true,
		},
		{
			name:     "syscall ECONNRESET returns true",
			err:      &os.SyscallError{Err: syscall.ECONNRESET},
			expected: true,
		},
		{
			name:     "wrapped syscall EPIPE returns true",
			err:      fmt.Errorf("write error: %w", &os.SyscallError{Err: syscall.EPIPE}),
			expected: true,
		},
		{
			name:     "wrapped syscall ECONNRESET returns true",
			err:      fmt.Errorf("connection error: %w", &os.SyscallError{Err: syscall.ECONNRESET}),
			expected: true,
		},
		{
			name:     "other syscall error returns false",
			err:      &os.SyscallError{Err: syscall.ECONNREFUSED},
			expected: false,
		},
		{
			name:     "generic error returns false",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "context canceled returns false",
			err:      fmt.Errorf("context canceled"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsConnectionClosedError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

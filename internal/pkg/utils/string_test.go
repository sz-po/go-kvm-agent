package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsKebabCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid single word",
			input:    "hello",
			expected: true,
		},
		{
			name:     "valid kebab case",
			input:    "hello-world",
			expected: true,
		},
		{
			name:     "valid with numbers",
			input:    "hello-world-123",
			expected: true,
		},
		{
			name:     "valid starting with number",
			input:    "123-hello",
			expected: true,
		},
		{
			name:     "valid multiple dashes",
			input:    "this-is-a-long-kebab-case-string",
			expected: true,
		},
		{
			name:     "invalid uppercase",
			input:    "Hello-World",
			expected: false,
		},
		{
			name:     "invalid camelCase",
			input:    "helloWorld",
			expected: false,
		},
		{
			name:     "invalid snake_case",
			input:    "hello_world",
			expected: false,
		},
		{
			name:     "invalid space",
			input:    "hello world",
			expected: false,
		},
		{
			name:     "invalid starting with dash",
			input:    "-hello",
			expected: false,
		},
		{
			name:     "invalid ending with dash",
			input:    "hello-",
			expected: false,
		},
		{
			name:     "invalid double dash",
			input:    "hello--world",
			expected: false,
		},
		{
			name:     "invalid empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "invalid special characters",
			input:    "hello@world",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsKebabCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

package edid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeManufacturer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		code     string
		expected [2]byte
		wantErr  bool
	}{
		{
			name:     "ValidUppercase",
			code:     "ABC",
			expected: [2]byte{0x04, 0x43},
		},
		{
			name:     "ValidLowercaseWithWhitespace",
			code:     " hpq ",
			expected: [2]byte{0x22, 0x11},
		},
		{
			name:    "InvalidLength",
			code:    "AB",
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			encoded, err := EncodeManufacturer(testCase.code)
			if testCase.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, encoded)
		})
	}
}

func TestDecodeManufacturer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		bytes    [2]byte
		expected string
		wantErr  bool
	}{
		{
			name:     "ValidABC",
			bytes:    [2]byte{0x04, 0x43},
			expected: "ABC",
		},
		{
			name:     "ValidHPQ",
			bytes:    [2]byte{0x22, 0x11},
			expected: "HPQ",
		},
		{
			name:    "InvalidZeroValue",
			bytes:   [2]byte{0x00, 0x00},
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			decoded, err := DecodeManufacturer(testCase.bytes)
			if testCase.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, decoded)
		})
	}
}

func TestEncodeDecodeManufacturerRoundTrip(t *testing.T) {
	t.Parallel()

	codes := []string{"ABC", "HPQ", "LGD"}

	for _, code := range codes {
		code := code
		t.Run(code, func(t *testing.T) {
			t.Parallel()

			encoded, err := EncodeManufacturer(code)
			assert.NoError(t, err)

			decoded, err := DecodeManufacturer(encoded)
			assert.NoError(t, err)
			assert.Equal(t, code, decoded)
		})
	}
}

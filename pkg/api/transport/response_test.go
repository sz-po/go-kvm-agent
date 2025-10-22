package transport

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockReadCloser struct {
	io.Reader
	closeCalled bool
	closeError  error
}

func (mock *mockReadCloser) Close() error {
	mock.closeCalled = true
	return mock.closeError
}

func TestUnmarshalResponseBodyWithReadCloserSuccess(t *testing.T) {
	t.Parallel()

	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	jsonData := `{"name":"test","value":42}`
	mock := &mockReadCloser{
		Reader: strings.NewReader(jsonData),
	}

	result, err := UnmarshalResponseBody[TestStruct](mock)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test", result.Name)
	assert.Equal(t, 42, result.Value)
	assert.True(t, mock.closeCalled, "Close() should be called on io.ReadCloser")
}

func TestUnmarshalResponseBodyWithReadCloserDecodeError(t *testing.T) {
	t.Parallel()

	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	invalidJSON := `{"name":"test","value":"not-a-number"}`
	mock := &mockReadCloser{
		Reader: strings.NewReader(invalidJSON),
	}

	result, err := UnmarshalResponseBody[TestStruct](mock)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, mock.closeCalled, "Close() should be called even when decode fails")
}

func TestUnmarshalResponseBodyWithReader(t *testing.T) {
	t.Parallel()

	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	jsonData := `{"name":"reader-test","value":123}`
	reader := strings.NewReader(jsonData)

	result, err := UnmarshalResponseBody[TestStruct](reader)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "reader-test", result.Name)
	assert.Equal(t, 123, result.Value)
}

func TestUnmarshalResponseBodyWithByteSlice(t *testing.T) {
	t.Parallel()

	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	jsonData := []byte(`{"name":"bytes-test","value":999}`)

	result, err := UnmarshalResponseBody[TestStruct](jsonData)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "bytes-test", result.Name)
	assert.Equal(t, 999, result.Value)
}

func TestUnmarshalResponseBodyUnsupportedType(t *testing.T) {
	t.Parallel()

	type TestStruct struct {
		Name string `json:"name"`
	}

	unsupportedBody := "plain string is not supported"

	result, err := UnmarshalResponseBody[TestStruct](unsupportedBody)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported type")
}

package transport

import (
	"errors"
	"fmt"
	"strings"
)

const (
	HeaderAccept      = "accept"
	HeaderContentType = "content-type"
)

type Header map[string]string

func (header Header) Get(key string) string {
	normalizedKey := strings.ToLower(key)
	return header[normalizedKey]
}

func (header Header) Has(key string) bool {
	normalizedKey := strings.ToLower(key)

	_, exists := header[normalizedKey]

	return exists
}

func (header Header) Require(keys ...string) error {
	for _, key := range keys {
		normalizedKey := strings.ToLower(key)
		if _, ok := header[normalizedKey]; !ok {
			return fmt.Errorf("%w: %s", ErrMissingHeaderKey, key)
		}
	}

	return nil
}

func (header Header) Clone() Header {
	if header == nil {
		return nil
	}

	clone := make(Header, len(header))
	for key, value := range header {
		clone[key] = value
	}

	return clone
}

var ErrMissingHeaderKey = errors.New("missing header key")

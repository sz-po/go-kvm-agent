package transport

import (
	"errors"
	"fmt"
)

type Header map[string]string

func (header Header) Require(keys ...string) error {
	for _, key := range keys {
		if _, ok := header[key]; !ok {
			return fmt.Errorf("%w: %s", ErrMissingHeaderKey, key)
		}
	}

	return nil
}

var ErrMissingHeaderKey = errors.New("missing header key")

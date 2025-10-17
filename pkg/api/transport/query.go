package transport

import (
	"errors"
	"fmt"
)

type Query map[string]string

func (query Query) Require(keys ...string) error {
	for _, key := range keys {
		if _, ok := query[key]; !ok {
			return fmt.Errorf("%w: %s", ErrMissingQueryKey, key)
		}
	}

	return nil
}

var ErrMissingQueryKey = errors.New("missing query key")

package transport

import (
	"errors"
	"fmt"
)

type Path map[string]string

func (path Path) Require(keys ...string) error {
	for _, key := range keys {
		if _, ok := path[key]; !ok {
			return fmt.Errorf("%w: %s", ErrMissingPathKey, key)
		}
	}

	return nil
}

var ErrMissingPathKey = errors.New("missing path key")

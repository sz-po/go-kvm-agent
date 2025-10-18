package transport

import (
	"errors"
	"fmt"
)

type PathParams map[string]string

func (path PathParams) Require(keys ...string) error {
	for _, key := range keys {
		if _, ok := path[key]; !ok {
			return fmt.Errorf("%w: %s", ErrMissingPathParamKey, key)
		}
	}

	return nil
}

func (path PathParams) Clone() PathParams {
	if path == nil {
		return nil
	}

	clone := make(PathParams, len(path))
	for key, value := range path {
		clone[key] = value
	}

	return clone
}

var ErrMissingPathParamKey = errors.New("missing path param key")

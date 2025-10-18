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

func (query Query) Clone() Query {
	if query == nil {
		return nil
	}

	clone := make(Query, len(query))
	for key, value := range query {
		clone[key] = value
	}

	return clone
}

var ErrMissingQueryKey = errors.New("missing query key")

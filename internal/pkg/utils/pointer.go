package utils

func DefaultNil[T any](pointer *T, val T) T {
	if pointer == nil {
		return val
	} else {
		return *pointer
	}
}

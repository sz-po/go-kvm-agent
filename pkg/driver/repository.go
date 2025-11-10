package driver

import (
	"context"
	"errors"
)

type DriverRepository interface {
	GetByKind(ctx context.Context, kind Kind) (Driver, error)
	GetAll(ctx context.Context) ([]Driver, error)
}

var ErrDriverNotFound = errors.New("driver not found")

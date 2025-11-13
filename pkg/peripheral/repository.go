package peripheral

import (
	"context"
	"errors"
)

type Repository interface {
	GetPeripheralById(ctx context.Context, id Id) (Peripheral, error)
	GetPeripheralByName(ctx context.Context, name Name) (Peripheral, error)

	GetAllPeripherals(ctx context.Context) ([]Peripheral, error)
}

var ErrPeripheralNotFound = errors.New("peripheral not found")

package peripheral

import (
	"context"
	"errors"
)

type Repository interface {
	GetPeripheralById(ctx context.Context, id PeripheralId) (Peripheral, error)
	GetPeripheralByName(ctx context.Context, name PeripheralName) (Peripheral, error)

	GetAllPeripherals(ctx context.Context) ([]Peripheral, error)
	GetAllDisplaySources(ctx context.Context) ([]DisplaySource, error)
	GetAllDisplaySinks(ctx context.Context) ([]DisplaySink, error)
}

var ErrPeripheralNotFound = errors.New("peripheral not found")

package machine

import (
	"context"
	"errors"
)

type Repository interface {
	GetMachineByName(ctx context.Context, name MachineName) (Machine, error)
	GetMachineById(ctx context.Context, id MachineId) (Machine, error)

	GetAllMachines(ctx context.Context) ([]Machine, error)
}

var ErrMachineNotFound = errors.New("machine not found")

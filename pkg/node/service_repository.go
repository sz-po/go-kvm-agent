package node

import (
	"context"
	"errors"
)

type ServiceRepository interface {
	GetServiceById(ctx context.Context, serviceId ServiceId) (Service, error)
	GetAllServiceIds(ctx context.Context) ([]ServiceId, error)
}

var (
	ErrServiceNotFound = errors.New("service not found")
)

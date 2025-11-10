package node

import (
	"context"
	"sync"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
)

type ServiceRepositoryOpt func(*ServiceRepository) error

type ServiceRepository struct {
	services     map[node.ServiceId]node.Service
	servicesLock *sync.RWMutex
}

var _ node.ServiceRepository = (*ServiceRepository)(nil)

func NewServiceRepository(opts ...ServiceRepositoryOpt) (*ServiceRepository, error) {
	repository := &ServiceRepository{
		services:     make(map[node.ServiceId]node.Service),
		servicesLock: &sync.RWMutex{},
	}

	for _, opt := range opts {
		err := opt(repository)
		if err != nil {
			return repository, err
		}
	}

	return repository, nil
}

func (repository *ServiceRepository) GetServiceById(ctx context.Context, serviceId node.ServiceId) (node.Service, error) {
	repository.servicesLock.RLock()
	defer repository.servicesLock.RUnlock()

	service, found := repository.services[serviceId]
	if !found {
		return nil, node.ErrServiceNotFound
	}

	return service, nil
}

func (repository *ServiceRepository) GetAllServiceIds(ctx context.Context) ([]node.ServiceId, error) {
	repository.servicesLock.RLock()
	defer repository.servicesLock.RUnlock()

	var serviceIds []node.ServiceId
	for serviceId := range repository.services {
		serviceIds = append(serviceIds, serviceId)
	}

	return serviceIds, nil
}

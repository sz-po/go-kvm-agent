package driver

import (
	"context"
	"errors"

	driverSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/driver"
)

type LocalRepository struct {
	drivers map[driverSDK.Kind]driverSDK.Driver
}

type LocalRepositoryOpt func(*LocalRepository) error

func WithDriver(driver driverSDK.Driver) LocalRepositoryOpt {
	return func(repository *LocalRepository) error {
		kind := driver.GetKind()

		if _, ok := repository.drivers[kind]; ok {
			return ErrDriverKindAlreadyRegistered
		}

		repository.drivers[kind] = driver
		return nil
	}
}

func NewLocalRepository(opts ...LocalRepositoryOpt) (*LocalRepository, error) {
	repository := &LocalRepository{
		drivers: make(map[driverSDK.Kind]driverSDK.Driver),
	}

	for _, opt := range opts {
		err := opt(repository)
		if err != nil {
			return nil, err
		}
	}

	return repository, nil
}

func (repository *LocalRepository) GetByKind(ctx context.Context, kind driverSDK.Kind) (driverSDK.Driver, error) {
	driver, found := repository.drivers[kind]
	if !found {
		return nil, driverSDK.ErrDriverNotFound
	}

	return driver, nil
}

func (repository *LocalRepository) GetAll(ctx context.Context) ([]driverSDK.Driver, error) {
	var result []driverSDK.Driver

	for _, driver := range repository.drivers {
		result = append(result, driver)
	}

	return result, nil
}

var ErrDriverKindAlreadyRegistered = errors.New("driver kind already registered")

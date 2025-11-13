package peripheral

import (
	"context"
	"errors"

	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type RepositoryOpt func(repository *Repository) error

type Repository struct {
	peripheralIdIndex   map[peripheralSDK.Id]peripheralSDK.Peripheral
	peripheralNameIndex map[peripheralSDK.Name]peripheralSDK.Peripheral
}

func WithPeripheral(peripheral peripheralSDK.Peripheral) RepositoryOpt {
	return func(repository *Repository) error {
		peripheralId := peripheral.GetId()
		peripheralName := peripheral.GetName()

		if _, found := repository.peripheralIdIndex[peripheralId]; found {
			return ErrPeripheralIdAlreadyTaken
		}

		if _, found := repository.peripheralNameIndex[peripheralName]; found {
			return ErrPeripheralNameAlreadyTaken
		}

		repository.peripheralIdIndex[peripheralId] = peripheral
		repository.peripheralNameIndex[peripheralName] = peripheral

		return nil
	}
}

func NewRepository(opts ...RepositoryOpt) (*Repository, error) {
	repository := &Repository{
		peripheralIdIndex:   make(map[peripheralSDK.Id]peripheralSDK.Peripheral),
		peripheralNameIndex: make(map[peripheralSDK.Name]peripheralSDK.Peripheral),
	}

	for _, opt := range opts {
		err := opt(repository)
		if err != nil {
			return nil, err
		}
	}

	return repository, nil
}

func (repository *Repository) GetPeripheralById(ctx context.Context, id peripheralSDK.Id) (peripheralSDK.Peripheral, error) {
	peripheral, found := repository.peripheralIdIndex[id]
	if !found {
		return nil, peripheralSDK.ErrPeripheralNotFound
	}

	return peripheral, nil
}

func (repository *Repository) GetPeripheralByName(ctx context.Context, name peripheralSDK.Name) (peripheralSDK.Peripheral, error) {
	peripheral, found := repository.peripheralNameIndex[name]
	if !found {
		return nil, peripheralSDK.ErrPeripheralNotFound
	}

	return peripheral, nil
}

func (repository *Repository) GetAllPeripherals(ctx context.Context) ([]peripheralSDK.Peripheral, error) {
	var peripherals []peripheralSDK.Peripheral

	for _, peripheral := range repository.peripheralIdIndex {
		peripherals = append(peripherals, peripheral)
	}

	return peripherals, nil
}

var ErrPeripheralIdAlreadyTaken = errors.New("peripheral id already taken")
var ErrPeripheralNameAlreadyTaken = errors.New("peripheral name already taken")

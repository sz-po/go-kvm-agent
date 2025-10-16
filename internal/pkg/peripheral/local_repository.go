package peripheral

import (
	"context"
	"fmt"
	"sync"

	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

// LocalRepositoryOpt is a functional option for configuring a LocalPeripheralRepository during creation.
type LocalRepositoryOpt func(*LocalPeripheralRepository) error

// LocalPeripheralRepository stores and provides access to peripheral devices.
type LocalPeripheralRepository struct {
	peripheralIdIndex   map[peripheralSDK.PeripheralId]peripheralSDK.Peripheral
	peripheralNameIndex map[peripheralSDK.PeripheralName]peripheralSDK.Peripheral
	peripheralLock      *sync.RWMutex
}

var _ peripheralSDK.Repository = (*LocalPeripheralRepository)(nil)

// WithPeripherals returns a LocalRepositoryOpt that adds a single peripheral to the repository.
func WithPeripherals(peripherals ...peripheralSDK.Peripheral) LocalRepositoryOpt {
	return func(repository *LocalPeripheralRepository) error {
		for _, peripheral := range peripherals {
			peripheralId := peripheral.GetId()
			peripheralName := peripheral.GetName()

			repository.peripheralIdIndex[peripheralId] = peripheral
			repository.peripheralNameIndex[peripheralName] = peripheral
		}

		return nil
	}
}

// WithPeripheralsFromProvider returns a LocalRepositoryOpt that adds all peripherals from a source to the repository.
func WithPeripheralsFromProvider(ctx context.Context, provider peripheralSDK.PeripheralProvider) LocalRepositoryOpt {
	return func(repository *LocalPeripheralRepository) error {
		peripherals, err := provider.GetAllPeripherals(ctx)
		if err != nil {
			return fmt.Errorf("get all peripherals: %w", err)
		}

		for _, peripheral := range peripherals {
			peripheralId := peripheral.GetId()
			peripheralName := peripheral.GetName()

			repository.peripheralIdIndex[peripheralId] = peripheral
			repository.peripheralNameIndex[peripheralName] = peripheral
		}

		return nil
	}
}

// NewLocalPeripheralRepository creates a new LocalPeripheralRepository instance with the given options.
func NewLocalPeripheralRepository(opts ...LocalRepositoryOpt) (*LocalPeripheralRepository, error) {
	repository := &LocalPeripheralRepository{
		peripheralIdIndex:   make(map[peripheralSDK.PeripheralId]peripheralSDK.Peripheral),
		peripheralNameIndex: make(map[peripheralSDK.PeripheralName]peripheralSDK.Peripheral),
		peripheralLock:      &sync.RWMutex{},
	}

	for _, opt := range opts {
		err := opt(repository)
		if err != nil {
			return nil, err
		}
	}

	return repository, nil
}

func (repository *LocalPeripheralRepository) GetPeripheralById(ctx context.Context, id peripheralSDK.PeripheralId) (peripheralSDK.Peripheral, error) {
	repository.peripheralLock.RLock()
	defer repository.peripheralLock.RUnlock()

	peripheral, exists := repository.peripheralIdIndex[id]
	if !exists {
		return nil, peripheralSDK.ErrPeripheralNotFound
	}

	return peripheral, nil
}

func (repository *LocalPeripheralRepository) GetPeripheralByName(ctx context.Context, name peripheralSDK.PeripheralName) (peripheralSDK.Peripheral, error) {
	repository.peripheralLock.RLock()
	defer repository.peripheralLock.RUnlock()

	peripheral, exists := repository.peripheralNameIndex[name]
	if !exists {
		return nil, peripheralSDK.ErrPeripheralNotFound
	}

	return peripheral, nil
}

func (repository *LocalPeripheralRepository) GetAllPeripherals(ctx context.Context) ([]peripheralSDK.Peripheral, error) {
	repository.peripheralLock.RLock()
	defer repository.peripheralLock.RUnlock()

	var peripherals []peripheralSDK.Peripheral

	for _, peripheral := range repository.peripheralIdIndex {
		peripherals = append(peripherals, peripheral)
	}

	return peripherals, nil
}

func (repository *LocalPeripheralRepository) GetAllDisplaySources(ctx context.Context) ([]peripheralSDK.DisplaySource, error) {
	repository.peripheralLock.RLock()
	defer repository.peripheralLock.RUnlock()

	var displaySources []peripheralSDK.DisplaySource

	for _, peripheral := range repository.peripheralIdIndex {
		displaySource, isDisplaySource := peripheral.(peripheralSDK.DisplaySource)
		if !isDisplaySource {
			continue
		}

		displaySources = append(displaySources, displaySource)
	}

	return displaySources, nil
}

func (repository *LocalPeripheralRepository) GetAllDisplaySinks(ctx context.Context) ([]peripheralSDK.DisplaySink, error) {
	repository.peripheralLock.RLock()
	defer repository.peripheralLock.RUnlock()

	var displaySinks []peripheralSDK.DisplaySink

	for _, peripheral := range repository.peripheralIdIndex {
		displaySink, isDisplaySink := peripheral.(peripheralSDK.DisplaySink)
		if !isDisplaySink {
			continue
		}

		displaySinks = append(displaySinks, displaySink)
	}

	return displaySinks, nil
}

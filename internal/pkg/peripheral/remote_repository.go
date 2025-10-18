package peripheral

import (
	"context"
	"fmt"
	"sync"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/peripheral/remote"
	machineClient "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/client/service/machine"
	peripheralAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type RemoteRepository struct {
	machineService *machineClient.Service

	peripheralIdIndex   map[peripheralSDK.PeripheralId]peripheralSDK.Peripheral
	peripheralNameIndex map[peripheralSDK.PeripheralName]peripheralSDK.Peripheral
	peripheralLock      *sync.RWMutex

	displaySources map[peripheralSDK.PeripheralId]peripheralSDK.DisplaySource
}

var _ peripheralSDK.Repository = (*RemoteRepository)(nil)

func NewRemoteRepository(ctx context.Context, machineService *machineClient.Service) (*RemoteRepository, error) {
	peripheralsIterator, err := machineService.Peripherals()
	if err != nil {
		return nil, fmt.Errorf("peripherals iterator: %w", err)
	}

	remotePeripherals, err := peripheralsIterator.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("peripherals list: %w", err)
	}

	repository := &RemoteRepository{
		machineService:      machineService,
		peripheralIdIndex:   make(map[peripheralSDK.PeripheralId]peripheralSDK.Peripheral),
		peripheralNameIndex: make(map[peripheralSDK.PeripheralName]peripheralSDK.Peripheral),
		displaySources:      make(map[peripheralSDK.PeripheralId]peripheralSDK.DisplaySource),
		peripheralLock:      &sync.RWMutex{},
	}

	for _, remotePeripheral := range remotePeripherals {
		remotePeripheralIdentifier := peripheralAPI.PeripheralIdentifier{
			Id: &remotePeripheral.Id,
		}

		peripheralService, err := machineService.Peripheral(remotePeripheralIdentifier)
		if err != nil {
			return nil, fmt.Errorf("peripheral service: %w", err)
		}

		peripheral, err := remote.NewPeripheral(ctx, *peripheralService)
		if err != nil {
			return nil, fmt.Errorf("new remote peripheral: %w", err)
		}

		repository.peripheralIdIndex[peripheral.GetId()] = peripheral
		repository.peripheralNameIndex[peripheral.GetName()] = peripheral

		if peripheral.IsDisplaySource() {
			displaySource, err := peripheral.AsDisplaySource()
			if err != nil {
				return nil, fmt.Errorf("as display source: %w", err)
			}

			repository.displaySources[peripheral.GetId()] = displaySource
		}
	}

	return repository, nil
}

func (repository *RemoteRepository) GetPeripheralById(ctx context.Context, id peripheralSDK.PeripheralId) (peripheralSDK.Peripheral, error) {
	repository.peripheralLock.RLock()
	defer repository.peripheralLock.RUnlock()

	peripheral, exists := repository.peripheralIdIndex[id]
	if !exists {
		return nil, peripheralSDK.ErrPeripheralNotFound
	}

	return peripheral, nil
}

func (repository *RemoteRepository) GetPeripheralByName(ctx context.Context, name peripheralSDK.PeripheralName) (peripheralSDK.Peripheral, error) {
	repository.peripheralLock.RLock()
	defer repository.peripheralLock.RUnlock()

	peripheral, exists := repository.peripheralNameIndex[name]
	if !exists {
		return nil, peripheralSDK.ErrPeripheralNotFound
	}

	return peripheral, nil
}

func (repository *RemoteRepository) GetAllPeripherals(ctx context.Context) ([]peripheralSDK.Peripheral, error) {
	repository.peripheralLock.RLock()
	defer repository.peripheralLock.RUnlock()

	var peripherals []peripheralSDK.Peripheral

	for _, peripheral := range repository.peripheralIdIndex {
		peripherals = append(peripherals, peripheral)
	}

	return peripherals, nil
}

func (repository *RemoteRepository) GetAllDisplaySources(ctx context.Context) ([]peripheralSDK.DisplaySource, error) {
	repository.peripheralLock.RLock()
	defer repository.peripheralLock.RUnlock()

	var displaySources []peripheralSDK.DisplaySource

	for _, displaySource := range repository.displaySources {
		displaySources = append(displaySources, displaySource)
	}

	return displaySources, nil
}

func (repository *RemoteRepository) GetAllDisplaySinks(ctx context.Context) ([]peripheralSDK.DisplaySink, error) {
	return []peripheralSDK.DisplaySink{}, nil
}

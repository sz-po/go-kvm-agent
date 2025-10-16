package machine

import (
	"context"
	"sync"

	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
)

type LocalRepositoryOpt func(repository *LocalRepository)
type LocalRepository struct {
	machineIdIndex   map[machineSDK.MachineId]machineSDK.Machine
	machineNameIndex map[machineSDK.MachineName]machineSDK.Machine
	machineLock      *sync.RWMutex
}

func WithLocalRepositoryMachines(machines ...machineSDK.Machine) LocalRepositoryOpt {
	return func(repository *LocalRepository) {
		for _, machine := range machines {
			machineId := machine.GetId()
			machineName := machine.GetName()

			repository.machineIdIndex[machineId] = machine
			repository.machineNameIndex[machineName] = machine
		}
	}
}

func NewLocalRepository(opts ...LocalRepositoryOpt) *LocalRepository {
	repository := &LocalRepository{
		machineIdIndex:   make(map[machineSDK.MachineId]machineSDK.Machine),
		machineNameIndex: make(map[machineSDK.MachineName]machineSDK.Machine),
		machineLock:      &sync.RWMutex{},
	}

	for _, opt := range opts {
		opt(repository)
	}

	return repository
}

func (repository *LocalRepository) GetMachineByName(ctx context.Context, name machineSDK.MachineName) (machineSDK.Machine, error) {
	repository.machineLock.RLock()
	defer repository.machineLock.RUnlock()

	machine, exists := repository.machineNameIndex[name]
	if !exists {
		return nil, machineSDK.ErrMachineNotFound
	}

	return machine, nil
}

func (repository *LocalRepository) GetMachineById(ctx context.Context, id machineSDK.MachineId) (machineSDK.Machine, error) {
	repository.machineLock.RLock()
	defer repository.machineLock.RUnlock()

	machine, exists := repository.machineIdIndex[id]
	if !exists {
		return nil, machineSDK.ErrMachineNotFound
	}

	return machine, nil
}

func (repository *LocalRepository) GetAllMachines(ctx context.Context) ([]machineSDK.Machine, error) {
	repository.machineLock.RLock()
	defer repository.machineLock.RUnlock()

	var machines []machineSDK.Machine

	for _, machine := range repository.machineIdIndex {
		machines = append(machines, machine)
	}

	return machines, nil
}

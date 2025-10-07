package peripheral

import (
	"errors"

	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

var (
	// ErrPeripheralNotFound indicates that the requested peripheral does not exist in the repository.
	ErrPeripheralNotFound = errors.New("peripheral not found")
)

// PeripheralSource is an interface for any component that can provide peripherals.
type PeripheralSource interface {
	GetPeripherals() []peripheralSDK.Peripheral
}

// RepositoryFilter is a predicate function for filtering peripherals.
type RepositoryFilter func(peripheralSDK.Peripheral) bool

// RepositoryOpt is a functional option for configuring a PeripheralRepository during creation.
type RepositoryOpt func(*PeripheralRepository)

// PeripheralRepository stores and provides access to peripheral devices.
type PeripheralRepository struct {
	peripherals map[peripheralSDK.PeripheralID]peripheralSDK.Peripheral
}

// WithPeripheral returns a RepositoryOpt that adds a single peripheral to the repository.
func WithPeripheral(peripheral peripheralSDK.Peripheral) RepositoryOpt {
	return func(repository *PeripheralRepository) {
		repository.peripherals[peripheral.ID()] = peripheral
	}
}

// WithPeripheralsFromSource returns a RepositoryOpt that adds all peripherals from a source to the repository.
func WithPeripheralsFromSource(source PeripheralSource) RepositoryOpt {
	return func(repository *PeripheralRepository) {
		for _, peripheral := range source.GetPeripherals() {
			repository.peripherals[peripheral.ID()] = peripheral
		}
	}
}

// NewPeripheralRepository creates a new PeripheralRepository instance with the given options.
func NewPeripheralRepository(opts ...RepositoryOpt) *PeripheralRepository {
	repository := &PeripheralRepository{
		peripherals: make(map[peripheralSDK.PeripheralID]peripheralSDK.Peripheral),
	}

	for _, opt := range opts {
		opt(repository)
	}

	return repository
}

// GetAll returns all peripherals that match the given filters.
// If no filters are provided, returns all peripherals.
// Multiple filters are combined with AND logic (peripheral must match all filters).
// The returned slice is a copy and modifications to it will not affect the repository's state.
func (repository *PeripheralRepository) GetAll(filters ...RepositoryFilter) []peripheralSDK.Peripheral {
	peripherals := make([]peripheralSDK.Peripheral, 0)
	for _, peripheral := range repository.peripherals {
		matches := true
		for _, filter := range filters {
			if !filter(peripheral) {
				matches = false
				break
			}
		}
		if matches {
			peripherals = append(peripherals, peripheral)
		}
	}
	return peripherals
}

// GetByID returns a specific peripheral by its ID.
// Returns ErrPeripheralNotFound if the peripheral does not exist.
func (repository *PeripheralRepository) GetByID(id peripheralSDK.PeripheralID) (peripheralSDK.Peripheral, error) {
	peripheral, ok := repository.peripherals[id]
	if !ok {
		return nil, ErrPeripheralNotFound
	}
	return peripheral, nil
}

// FilterType returns a RepositoryFilter that matches peripherals of the specified type.
func FilterType(pType peripheralSDK.PeripheralType) RepositoryFilter {
	return func(peripheral peripheralSDK.Peripheral) bool {
		return peripheral.ID().Type() == pType
	}
}

// FilterRole returns a RepositoryFilter that matches peripherals with the specified role.
func FilterRole(role peripheralSDK.PeripheralRole) RepositoryFilter {
	return func(peripheral peripheralSDK.Peripheral) bool {
		return peripheral.ID().Role() == role
	}
}

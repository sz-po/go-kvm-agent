package peripheral

import (
	"errors"

	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

var (
	// ErrPeripheralNotFound indicates that the requested peripheral does not exist in the repository.
	ErrPeripheralNotFound = errors.New("peripheral not found")
)

// PeripheralProvider is an interface for any component that can provide peripherals.
type PeripheralProvider interface {
	GetPeripherals() []peripheralSDK.Peripheral
}

// RepositoryFilter is a predicate function for filtering peripherals.
type RepositoryFilter func(peripheralSDK.Peripheral) bool

// RepositoryOpt is a functional option for configuring a PeripheralRepository during creation.
type RepositoryOpt func(*PeripheralRepository) error

// PeripheralRepository stores and provides access to peripheral devices.
type PeripheralRepository struct {
	peripherals map[peripheralSDK.PeripheralId]peripheralSDK.Peripheral
}

// WithPeripheral returns a RepositoryOpt that adds a single peripheral to the repository.
func WithPeripheral(peripheral peripheralSDK.Peripheral) RepositoryOpt {
	return func(repository *PeripheralRepository) error {
		if err := peripheralSDK.ValidatePeripheralCapabilities(peripheral); err != nil {
			return err
		}

		repository.peripherals[peripheral.Id()] = peripheral

		return nil
	}
}

// WithPeripheralsFromProvider returns a RepositoryOpt that adds all peripherals from a source to the repository.
func WithPeripheralsFromProvider(source PeripheralProvider) RepositoryOpt {
	return func(repository *PeripheralRepository) error {
		for _, peripheral := range source.GetPeripherals() {
			if err := peripheralSDK.ValidatePeripheralCapabilities(peripheral); err != nil {
				return err
			}

			repository.peripherals[peripheral.Id()] = peripheral
		}
		return nil
	}
}

// NewPeripheralRepository creates a new PeripheralRepository instance with the given options.
func NewPeripheralRepository(opts ...RepositoryOpt) (*PeripheralRepository, error) {
	repository := &PeripheralRepository{
		peripherals: make(map[peripheralSDK.PeripheralId]peripheralSDK.Peripheral),
	}

	for _, opt := range opts {
		err := opt(repository)
		if err != nil {
			return nil, err
		}
	}

	return repository, nil
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

// GetByID returns a specific peripheral by its Name.
// Returns ErrPeripheralNotFound if the peripheral does not exist.
func (repository *PeripheralRepository) GetByID(id peripheralSDK.PeripheralId) (peripheralSDK.Peripheral, error) {
	peripheral, ok := repository.peripherals[id]
	if !ok {
		return nil, ErrPeripheralNotFound
	}
	return peripheral, nil
}

// FilterCapability returns a RepositoryFilter that matches peripherals with the specified capability.
func FilterCapability(capability peripheralSDK.PeripheralCapability) RepositoryFilter {
	return func(peripheral peripheralSDK.Peripheral) bool {
		for _, cap := range peripheral.Capabilities() {
			if cap.Kind == capability.Kind && cap.Role == capability.Role {
				return true
			}
		}
		return false
	}
}

// GetAllDisplaySources returns all peripherals that implement DisplaySource.
// The returned slice is type-safe and ready for iteration.
func (repository *PeripheralRepository) GetAllDisplaySources() []peripheralSDK.DisplaySource {
	peripherals := repository.GetAll(FilterCapability(peripheralSDK.DisplaySourceCapability))
	result := make([]peripheralSDK.DisplaySource, 0, len(peripherals))
	for _, peripheral := range peripherals {
		if source, ok := peripheral.(peripheralSDK.DisplaySource); ok {
			result = append(result, source)
		}
	}
	return result
}

// GetAllDisplaySinks returns all peripherals that implement DisplaySink.
// The returned slice is type-safe and ready for iteration.
func (repository *PeripheralRepository) GetAllDisplaySinks() []peripheralSDK.DisplaySink {
	peripherals := repository.GetAll(FilterCapability(peripheralSDK.DisplaySinkCapability))
	result := make([]peripheralSDK.DisplaySink, 0, len(peripherals))
	for _, peripheral := range peripherals {
		if sink, ok := peripheral.(peripheralSDK.DisplaySink); ok {
			result = append(result, sink)
		}
	}
	return result
}

// GetAllKeyboardSources returns all peripherals that implement KeyboardSource.
// The returned slice is type-safe and ready for iteration.
func (repository *PeripheralRepository) GetAllKeyboardSources() []peripheralSDK.KeyboardSource {
	peripherals := repository.GetAll(FilterCapability(peripheralSDK.KeyboardSourceCapability))
	result := make([]peripheralSDK.KeyboardSource, 0, len(peripherals))
	for _, peripheral := range peripherals {
		if source, ok := peripheral.(peripheralSDK.KeyboardSource); ok {
			result = append(result, source)
		}
	}
	return result
}

// GetAllKeyboardSinks returns all peripherals that implement KeyboardSink.
// The returned slice is type-safe and ready for iteration.
func (repository *PeripheralRepository) GetAllKeyboardSinks() []peripheralSDK.KeyboardSink {
	peripherals := repository.GetAll(FilterCapability(peripheralSDK.KeyboardSinkCapability))
	result := make([]peripheralSDK.KeyboardSink, 0, len(peripherals))
	for _, peripheral := range peripherals {
		if sink, ok := peripheral.(peripheralSDK.KeyboardSink); ok {
			result = append(result, sink)
		}
	}
	return result
}

// GetAllMouseSources returns all peripherals that implement MouseSource.
// The returned slice is type-safe and ready for iteration.
func (repository *PeripheralRepository) GetAllMouseSources() []peripheralSDK.MouseSource {
	peripherals := repository.GetAll(FilterCapability(peripheralSDK.MouseSourceCapability))
	result := make([]peripheralSDK.MouseSource, 0, len(peripherals))
	for _, peripheral := range peripherals {
		if source, ok := peripheral.(peripheralSDK.MouseSource); ok {
			result = append(result, source)
		}
	}
	return result
}

// GetAllMouseSinks returns all peripherals that implement MouseSink.
// The returned slice is type-safe and ready for iteration.
func (repository *PeripheralRepository) GetAllMouseSinks() []peripheralSDK.MouseSink {
	peripherals := repository.GetAll(FilterCapability(peripheralSDK.MouseSinkCapability))
	result := make([]peripheralSDK.MouseSink, 0, len(peripherals))
	for _, peripheral := range peripherals {
		if sink, ok := peripheral.(peripheralSDK.MouseSink); ok {
			result = append(result, sink)
		}
	}
	return result
}

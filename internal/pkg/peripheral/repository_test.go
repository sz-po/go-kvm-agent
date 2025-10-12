package peripheral

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type mockPeripheralProvider struct {
	peripherals []peripheralSDK.Peripheral
}

func (provider *mockPeripheralProvider) GetPeripherals() []peripheralSDK.Peripheral {
	return provider.peripherals
}

func TestNewPeripheralRepository(t *testing.T) {
	t.Run("creates empty repository", func(t *testing.T) {
		repository, err := NewPeripheralRepository()
		assert.NoError(t, err)
		assert.NotNil(t, repository)
		assert.Empty(t, repository.GetAll())
	})

	t.Run("creates with single peripheral", func(t *testing.T) {
		id := peripheralSDK.PeripheralId("mpv-source-1")
		source := peripheralSDK.NewDisplaySourceMock(t)
		source.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySourceCapability})
		source.EXPECT().Id().Return(id)

		repository, err := NewPeripheralRepository(WithPeripheral(source))
		assert.NoError(t, err)
		assert.Len(t, repository.GetAll(), 1)

		found, err := repository.GetByID(id)
		assert.NoError(t, err)
		assert.Equal(t, source, found)
	})

	t.Run("creates with provider", func(t *testing.T) {
		id1 := peripheralSDK.PeripheralId("mpv-source-1")
		source := peripheralSDK.NewDisplaySourceMock(t)
		source.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySourceCapability})
		source.EXPECT().Id().Return(id1)

		id2 := peripheralSDK.PeripheralId("mpv-sink-1")
		sink := peripheralSDK.NewDisplaySinkMock(t)
		sink.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySinkCapability})
		sink.EXPECT().Id().Return(id2)

		provider := &mockPeripheralProvider{
			peripherals: []peripheralSDK.Peripheral{source, sink},
		}

		repository, err := NewPeripheralRepository(WithPeripheralsFromProvider(provider))
		assert.NoError(t, err)
		assert.Len(t, repository.GetAll(), 2)
	})
}

func TestPeripheralRepository_GetAll(t *testing.T) {
	t.Run("returns all peripherals without filters", func(t *testing.T) {
		id1 := peripheralSDK.PeripheralId("mpv-source-1")
		source := peripheralSDK.NewDisplaySourceMock(t)
		source.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySourceCapability})
		source.EXPECT().Id().Return(id1)

		id2 := peripheralSDK.PeripheralId("mpv-sink-1")
		sink := peripheralSDK.NewDisplaySinkMock(t)
		sink.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySinkCapability})
		sink.EXPECT().Id().Return(id2)

		repository, err := NewPeripheralRepository(
			WithPeripheral(source),
			WithPeripheral(sink),
		)
		assert.NoError(t, err)

		all := repository.GetAll()
		assert.Len(t, all, 2)
	})

	t.Run("filters by capability", func(t *testing.T) {
		id1 := peripheralSDK.PeripheralId("mpv-source-1")
		source := peripheralSDK.NewDisplaySourceMock(t)
		source.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySourceCapability}).Times(2)
		source.EXPECT().Id().Return(id1)

		id2 := peripheralSDK.PeripheralId("mpv-sink-1")
		sink := peripheralSDK.NewDisplaySinkMock(t)
		sink.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySinkCapability}).Times(2)
		sink.EXPECT().Id().Return(id2)

		repository, err := NewPeripheralRepository(
			WithPeripheral(source),
			WithPeripheral(sink),
		)
		assert.NoError(t, err)

		sources := repository.GetAll(FilterCapability(peripheralSDK.DisplaySourceCapability))
		assert.Len(t, sources, 1)
		assert.Equal(t, source, sources[0])
	})

	t.Run("filters with multiple filters", func(t *testing.T) {
		id1 := peripheralSDK.PeripheralId("mpv-source-1")
		source1 := peripheralSDK.NewDisplaySourceMock(t)
		source1.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySourceCapability}).Times(2)
		source1.EXPECT().Id().Return(id1).Times(2) // Called twice: WithPeripheral + filterByIdPrefix

		id2 := peripheralSDK.PeripheralId("mpv-source-2")
		source2 := peripheralSDK.NewDisplaySourceMock(t)
		source2.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySourceCapability}).Times(2)
		source2.EXPECT().Id().Return(id2).Times(2) // Called twice: WithPeripheral + filterByIdPrefix

		id3 := peripheralSDK.PeripheralId("mpv-sink-1")
		sink := peripheralSDK.NewDisplaySinkMock(t)
		sink.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySinkCapability}).Times(2)
		sink.EXPECT().Id().Return(id3)

		repository, err := NewPeripheralRepository(
			WithPeripheral(source1),
			WithPeripheral(source2),
			WithPeripheral(sink),
		)
		assert.NoError(t, err)

		// Filter by capability AND ID prefix
		filterByIdPrefix := func(peripheral peripheralSDK.Peripheral) bool {
			peripheralId := string(peripheral.Id())
			return strings.HasPrefix(peripheralId, "mpv-source-")
		}

		result := repository.GetAll(
			FilterCapability(peripheralSDK.DisplaySourceCapability),
			filterByIdPrefix,
		)
		assert.Len(t, result, 2)
	})

	t.Run("returns empty slice when no matches", func(t *testing.T) {
		id := peripheralSDK.PeripheralId("mpv-source-1")
		source := peripheralSDK.NewDisplaySourceMock(t)
		source.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySourceCapability}).Times(2)
		source.EXPECT().Id().Return(id)

		repository, err := NewPeripheralRepository(WithPeripheral(source))
		assert.NoError(t, err)

		result := repository.GetAll(FilterCapability(peripheralSDK.KeyboardSourceCapability))
		assert.Empty(t, result)
	})
}

func TestPeripheralRepository_GetByID(t *testing.T) {
	t.Run("returns peripheral when exists", func(t *testing.T) {
		id := peripheralSDK.PeripheralId("mpv-source-1")
		source := peripheralSDK.NewDisplaySourceMock(t)
		source.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySourceCapability})
		source.EXPECT().Id().Return(id)

		repository, err := NewPeripheralRepository(WithPeripheral(source))
		assert.NoError(t, err)

		found, err := repository.GetByID(id)
		assert.NoError(t, err)
		assert.Equal(t, source, found)
	})

	t.Run("returns error when not found", func(t *testing.T) {
		repository, err := NewPeripheralRepository()
		assert.NoError(t, err)

		found, err := repository.GetByID(peripheralSDK.PeripheralId("nonexistent"))
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrPeripheralNotFound))
		assert.Nil(t, found)
	})
}

func TestPeripheralRepository_GetAllDisplaySources(t *testing.T) {
	t.Run("returns all display sources", func(t *testing.T) {
		id1 := peripheralSDK.PeripheralId("mpv-source-1")
		source1 := peripheralSDK.NewDisplaySourceMock(t)
		source1.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySourceCapability}).Times(2)
		source1.EXPECT().Id().Return(id1)

		id2 := peripheralSDK.PeripheralId("mpv-source-2")
		source2 := peripheralSDK.NewDisplaySourceMock(t)
		source2.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySourceCapability}).Times(2)
		source2.EXPECT().Id().Return(id2)

		id3 := peripheralSDK.PeripheralId("mpv-sink-1")
		sink := peripheralSDK.NewDisplaySinkMock(t)
		sink.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySinkCapability}).Times(2)
		sink.EXPECT().Id().Return(id3)

		repository, err := NewPeripheralRepository(
			WithPeripheral(source1),
			WithPeripheral(source2),
			WithPeripheral(sink),
		)
		assert.NoError(t, err)

		sources := repository.GetAllDisplaySources()
		assert.Len(t, sources, 2)
	})

	t.Run("returns empty slice when no display sources", func(t *testing.T) {
		id := peripheralSDK.PeripheralId("mpv-sink-1")
		sink := peripheralSDK.NewDisplaySinkMock(t)
		sink.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySinkCapability}).Times(2)
		sink.EXPECT().Id().Return(id)

		repository, err := NewPeripheralRepository(WithPeripheral(sink))
		assert.NoError(t, err)

		sources := repository.GetAllDisplaySources()
		assert.Empty(t, sources)
	})
}

func TestPeripheralRepository_GetAllDisplaySinks(t *testing.T) {
	t.Run("returns all display sinks", func(t *testing.T) {
		id1 := peripheralSDK.PeripheralId("mpv-sink-1")
		sink1 := peripheralSDK.NewDisplaySinkMock(t)
		sink1.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySinkCapability}).Times(2)
		sink1.EXPECT().Id().Return(id1)

		id2 := peripheralSDK.PeripheralId("mpv-sink-2")
		sink2 := peripheralSDK.NewDisplaySinkMock(t)
		sink2.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySinkCapability}).Times(2)
		sink2.EXPECT().Id().Return(id2)

		id3 := peripheralSDK.PeripheralId("mpv-source-1")
		source := peripheralSDK.NewDisplaySourceMock(t)
		source.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySourceCapability}).Times(2)
		source.EXPECT().Id().Return(id3)

		repository, err := NewPeripheralRepository(
			WithPeripheral(sink1),
			WithPeripheral(sink2),
			WithPeripheral(source),
		)
		assert.NoError(t, err)

		sinks := repository.GetAllDisplaySinks()
		assert.Len(t, sinks, 2)
	})

	t.Run("returns empty slice when no display sinks", func(t *testing.T) {
		id := peripheralSDK.PeripheralId("mpv-source-1")
		source := peripheralSDK.NewDisplaySourceMock(t)
		source.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySourceCapability}).Times(2)
		source.EXPECT().Id().Return(id)

		repository, err := NewPeripheralRepository(WithPeripheral(source))
		assert.NoError(t, err)

		sinks := repository.GetAllDisplaySinks()
		assert.Empty(t, sinks)
	})
}

func TestFilterCapability(t *testing.T) {
	t.Run("matches peripheral with specified capability", func(t *testing.T) {
		source := peripheralSDK.NewDisplaySourceMock(t)
		source.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySourceCapability})

		filter := FilterCapability(peripheralSDK.DisplaySourceCapability)
		assert.True(t, filter(source))
	})

	t.Run("does not match peripheral without specified capability", func(t *testing.T) {
		sink := peripheralSDK.NewDisplaySinkMock(t)
		sink.EXPECT().Capabilities().Return([]peripheralSDK.PeripheralCapability{peripheralSDK.DisplaySinkCapability})

		filter := FilterCapability(peripheralSDK.DisplaySourceCapability)
		assert.False(t, filter(sink))
	})
}

// TODO: Add tests for GetAllKeyboardSources when MockedKeyboardSource is available
// TODO: Add tests for GetAllKeyboardSinks when MockedKeyboardSink is available
// TODO: Add tests for GetAllMouseSources when MockedMouseSource is available
// TODO: Add tests for GetAllMouseSinks when MockedMouseSink is available

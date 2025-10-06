package test

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

// MockedDisplaySource is a testify-based mock implementing peripheral.DisplaySource.
type MockedDisplaySource struct {
	mock.Mock

	id peripheral.PeripheralID
}

// NewMockedDisplaySource returns a mock display source seeded with the provided ID.
func NewMockedDisplaySource(id peripheral.PeripheralID) *MockedDisplaySource {
	return &MockedDisplaySource{id: id}
}

// SetID overrides the peripheral identifier returned by ID().
func (m *MockedDisplaySource) SetID(id peripheral.PeripheralID) {
	m.id = id
}

// ID implements peripheral.Peripheral.
func (m *MockedDisplaySource) ID() peripheral.PeripheralID {
	return m.id
}

// DataChannel mocks the channel used for display events.
func (m *MockedDisplaySource) DataChannel(ctx context.Context) <-chan peripheral.DisplayEvent {
	args := m.Called(ctx)
	if ch, ok := args.Get(0).(<-chan peripheral.DisplayEvent); ok {
		return ch
	}
	return nil
}

// ControlChannel mocks the channel used for control events.
func (m *MockedDisplaySource) ControlChannel(ctx context.Context) <-chan peripheral.DisplayControlEvent {
	args := m.Called(ctx)
	if ch, ok := args.Get(0).(<-chan peripheral.DisplayControlEvent); ok {
		return ch
	}
	return nil
}

// GetCurrentDisplayMode mocks the mode query.
func (m *MockedDisplaySource) GetCurrentDisplayMode() (*peripheral.DisplayMode, error) {
	args := m.Called()
	var mode *peripheral.DisplayMode
	if val := args.Get(0); val != nil {
		if cast, ok := val.(*peripheral.DisplayMode); ok {
			mode = cast
		}
	}
	return mode, args.Error(1)
}

// Start mocks source startup.
func (m *MockedDisplaySource) Start(ctx context.Context, info peripheral.DisplayInfo) error {
	args := m.Called(ctx, info)
	return args.Error(0)
}

// Stop mocks source shutdown.
func (m *MockedDisplaySource) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

var _ peripheral.DisplaySource = (*MockedDisplaySource)(nil)

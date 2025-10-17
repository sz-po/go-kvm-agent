package routing

import (
	"context"
	"errors"

	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type DisplayRouter interface {
	// Connect source to sink.
	Connect(ctx context.Context, sourceId peripheralSDK.PeripheralId, sinkId peripheralSDK.PeripheralId) error

	// DisconnectSink disconnects sinkId from any source.
	DisconnectSink(ctx context.Context, sinkId peripheralSDK.PeripheralId) error

	// DisconnectSource disconnects all sink connected to source identified by sourceId.
	DisconnectSource(ctx context.Context, sourceId peripheralSDK.PeripheralId) error
}

var ErrDisplaySourceNotRegistered = errors.New("display source not registered")
var ErrDisplaySourceAlreadyRegistered = errors.New("display source already registered")
var ErrDisplaySinkNotRegistered = errors.New("display sink not registered")
var ErrDisplaySinkAlreadyRegistered = errors.New("display sink already registered")
var ErrDisplaySinkNotConnected = errors.New("display sink not connected")
var ErrDisplaySourceNotConnected = errors.New("display source not connected")
var ErrDisplaySinkAlreadyConnectedToSource = errors.New("display sink already connected to source")

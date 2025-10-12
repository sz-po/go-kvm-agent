package routing

import (
	"errors"

	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type DisplayRouter interface {
	// Connect source to sink.
	Connect(sourceId peripheralSDK.PeripheralId, sinkId peripheralSDK.PeripheralId) error

	// DisconnectSink disconnects sinkId from any source.
	DisconnectSink(sinkId peripheralSDK.PeripheralId) error

	// DisconnectSource disconnects all sink connected to source identified by sourceId.
	DisconnectSource(sourceId peripheralSDK.PeripheralId) error
}

var ErrDisplaySourceNotRegistered = errors.New("display source not registered")
var ErrDisplaySourceAlreadyRegistered = errors.New("display source already registered")
var ErrDisplaySinkNotRegistered = errors.New("display sink not registered")
var ErrDisplaySinkAlreadyRegistered = errors.New("display sink already registered")
var ErrDisplaySinkNotConnected = errors.New("display sink not linked")
var ErrDisplaySourceNotConnected = errors.New("display source not linked")
var ErrDisplaySourceAlreadyConnected = errors.New("display source already connected")

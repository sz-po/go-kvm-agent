package routing

import (
	"context"
	"fmt"
	"sync"

	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
	routingSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/routing"
)

type LocalDisplayRouterOpt func(*LocalDisplayRouter) error

type LocalDisplayRouter struct {
	sinkIdIndex map[peripheralSDK.PeripheralId]peripheralSDK.DisplaySink
	sinkLock    *sync.RWMutex

	sourceIdIndex map[peripheralSDK.PeripheralId]peripheralSDK.DisplaySource
	sourceLock    *sync.RWMutex

	sinkConnection       map[peripheralSDK.DisplaySink]peripheralSDK.DisplaySource
	sinkConnectionCancel map[peripheralSDK.DisplaySink]context.CancelCauseFunc
	sinkConnectionLock   *sync.RWMutex
}

func WithDisplaySink(displaySink peripheralSDK.DisplaySink) LocalDisplayRouterOpt {
	return func(router *LocalDisplayRouter) error {
		displaySinkId := displaySink.GetId()

		if _, exists := router.sinkIdIndex[displaySinkId]; exists {
			return fmt.Errorf("%w: duplicate sink id: %s", routingSDK.ErrDisplaySinkAlreadyRegistered, displaySinkId)
		}

		router.sinkIdIndex[displaySinkId] = displaySink

		return nil
	}
}

func WithDisplaySource(displaySource peripheralSDK.DisplaySource) LocalDisplayRouterOpt {
	return func(router *LocalDisplayRouter) error {
		displaySourceId := displaySource.GetId()

		if _, exists := router.sourceIdIndex[displaySourceId]; exists {
			return fmt.Errorf("%w: duplicate source id: %s", routingSDK.ErrDisplaySourceAlreadyRegistered, displaySourceId)
		}

		router.sourceIdIndex[displaySourceId] = displaySource

		return nil
	}
}

func NewLocalDisplayRouter(opts ...LocalDisplayRouterOpt) (*LocalDisplayRouter, error) {
	router := &LocalDisplayRouter{
		sinkIdIndex: make(map[peripheralSDK.PeripheralId]peripheralSDK.DisplaySink),
		sinkLock:    &sync.RWMutex{},

		sourceIdIndex: make(map[peripheralSDK.PeripheralId]peripheralSDK.DisplaySource),
		sourceLock:    &sync.RWMutex{},

		sinkConnection:     make(map[peripheralSDK.DisplaySink]peripheralSDK.DisplaySource),
		sinkConnectionLock: &sync.RWMutex{},
	}

	for _, opt := range opts {
		if err := opt(router); err != nil {
			return nil, err
		}
	}

	return router, nil
}

func (router *LocalDisplayRouter) Connect(ctx context.Context, sourceId peripheralSDK.PeripheralId, sinkId peripheralSDK.PeripheralId) error {
	displaySink, err := router.getDisplaySinkById(sinkId)
	if err != nil {
		return fmt.Errorf("get display sinkIdIndex: %w", err)
	}

	displaySource, err := router.getDisplaySourceById(sourceId)
	if err != nil {
		return fmt.Errorf("get display sourceIdIndex: %w", err)
	}

	err = router.connectDisplaySource(displaySink, displaySource)
	if err != nil {
		return fmt.Errorf("connect display sourceIdIndex: %w", err)
	}

	return nil
}

func (router *LocalDisplayRouter) DisconnectSink(ctx context.Context, sinkId peripheralSDK.PeripheralId) error {
	panic("implement me")
}

func (router *LocalDisplayRouter) DisconnectSource(ctx context.Context, sourceId peripheralSDK.PeripheralId) error {
	//TODO implement me
	panic("implement me")
}

func (router *LocalDisplayRouter) getDisplaySinkById(sinkId peripheralSDK.PeripheralId) (peripheralSDK.DisplaySink, error) {
	router.sinkLock.RLock()
	defer router.sinkLock.RUnlock()

	displaySink, exists := router.sinkIdIndex[sinkId]
	if !exists {
		return nil, routingSDK.ErrDisplaySinkNotRegistered
	}

	return displaySink, nil
}

func (router *LocalDisplayRouter) getDisplaySourceById(sourceId peripheralSDK.PeripheralId) (peripheralSDK.DisplaySource, error) {
	router.sourceLock.RLock()
	defer router.sourceLock.RUnlock()

	displaySource, exists := router.sourceIdIndex[sourceId]
	if !exists {
		return nil, routingSDK.ErrDisplaySourceNotRegistered
	}

	return displaySource, nil
}

func (router *LocalDisplayRouter) getConnectedDisplaySource(displaySink peripheralSDK.DisplaySink) (peripheralSDK.DisplaySource, error) {
	router.sinkConnectionLock.RLock()
	defer router.sinkConnectionLock.RUnlock()

	displaySource, connected := router.sinkConnection[displaySink]
	if !connected {
		return nil, routingSDK.ErrDisplaySourceNotConnected
	}

	return displaySource, nil
}

func (router *LocalDisplayRouter) connectDisplaySource(displaySink peripheralSDK.DisplaySink, displaySource peripheralSDK.DisplaySource) error {
	router.sinkConnectionLock.Lock()
	defer router.sinkConnectionLock.Unlock()

	if _, isConnected := router.sinkConnection[displaySink]; isConnected {
		return routingSDK.ErrDisplaySinkAlreadyConnectedToSource
	}

	err := displaySink.SetDisplayFrameBufferProvider(displaySource)
	if err != nil {
		return fmt.Errorf("set sink display frame buffer provider: %w", err)
	}

	router.sinkConnection[displaySink] = displaySource

	return nil
}

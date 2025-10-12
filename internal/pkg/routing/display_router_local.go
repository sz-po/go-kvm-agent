package routing

import (
	"context"
	"fmt"
	"sync"

	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
	routing2 "github.com/szymonpodeszwa/go-kvm-agent/pkg/routing"
)

type LocalDisplayRouterOpt func(*LocalDisplayRouter) error

type LocalDisplayRouter struct {
	sink     map[peripheralSDK.PeripheralId]peripheralSDK.DisplaySink
	sinkLock sync.RWMutex

	source     map[peripheralSDK.PeripheralId]peripheralSDK.DisplaySource
	sourceLock sync.RWMutex

	connectionSource map[peripheralSDK.DisplaySink]peripheralSDK.DisplaySource
	connectionCancel map[peripheralSDK.DisplaySink]context.CancelCauseFunc
	connectionLock   sync.RWMutex
}

func WithDisplaySink(displaySink peripheralSDK.DisplaySink) LocalDisplayRouterOpt {
	return func(router *LocalDisplayRouter) error {
		if _, exists := router.sink[displaySink.Id()]; exists {
			return routing2.ErrDisplaySinkAlreadyRegistered
		}

		router.sink[displaySink.Id()] = displaySink
		return nil
	}
}

func WithDisplaySource(displaySource peripheralSDK.DisplaySource) LocalDisplayRouterOpt {
	return func(router *LocalDisplayRouter) error {
		if _, exists := router.source[displaySource.Id()]; exists {
			return routing2.ErrDisplaySourceAlreadyRegistered
		}

		router.source[displaySource.Id()] = displaySource
		return nil
	}
}

func NewLocalDisplayRouter(opts ...LocalDisplayRouterOpt) (*LocalDisplayRouter, error) {
	router := &LocalDisplayRouter{
		sink:     make(map[peripheralSDK.PeripheralId]peripheralSDK.DisplaySink),
		sinkLock: sync.RWMutex{},

		source:     make(map[peripheralSDK.PeripheralId]peripheralSDK.DisplaySource),
		sourceLock: sync.RWMutex{},

		connectionSource: make(map[peripheralSDK.DisplaySink]peripheralSDK.DisplaySource),
		connectionCancel: make(map[peripheralSDK.DisplaySink]context.CancelCauseFunc),
		connectionLock:   sync.RWMutex{},
	}

	for _, opt := range opts {
		if err := opt(router); err != nil {
			return nil, err
		}
	}

	return router, nil
}

func (router *LocalDisplayRouter) Connect(sourceId peripheralSDK.PeripheralId, sinkId peripheralSDK.PeripheralId) error {
	displaySink, err := router.getDisplaySinkById(sinkId)
	if err != nil {
		return fmt.Errorf("get display sink: %w", err)
	}

	displaySource, err := router.getDisplaySourceById(sourceId)
	if err != nil {
		return fmt.Errorf("get display source: %w", err)
	}

	if router.hasConnectedDisplaySourceSource(displaySink) {
		return routing2.ErrDisplaySourceAlreadyConnected
	}

	err = router.connectDisplaySource(displaySink, displaySource)
	if err != nil {
		return fmt.Errorf("connect display source: %w", err)
	}

	return nil
}

func (router *LocalDisplayRouter) DisconnectSink(sinkId peripheralSDK.PeripheralId) error {
	//TODO implement me
	panic("implement me")
}

func (router *LocalDisplayRouter) DisconnectSource(sourceId peripheralSDK.PeripheralId) error {
	//TODO implement me
	panic("implement me")
}

func (router *LocalDisplayRouter) getDisplaySinkById(sinkId peripheralSDK.PeripheralId) (peripheralSDK.DisplaySink, error) {
	router.sinkLock.RLock()
	defer router.sinkLock.RUnlock()

	displaySink, exists := router.sink[sinkId]
	if !exists {
		return nil, routing2.ErrDisplaySinkNotRegistered
	}

	return displaySink, nil
}

func (router *LocalDisplayRouter) getDisplaySourceById(sourceId peripheralSDK.PeripheralId) (peripheralSDK.DisplaySource, error) {
	router.sourceLock.RLock()
	defer router.sourceLock.RUnlock()

	displaySource, exists := router.source[sourceId]
	if !exists {
		return nil, routing2.ErrDisplaySourceNotRegistered
	}

	return displaySource, nil
}

func (router *LocalDisplayRouter) hasConnectedDisplaySourceSource(displaySink peripheralSDK.DisplaySink) bool {
	router.connectionLock.RLock()
	defer router.connectionLock.RUnlock()

	_, connected := router.source[displaySink.Id()]
	return connected
}

func (router *LocalDisplayRouter) getConnectedDisplaySource(displaySink peripheralSDK.DisplaySink) (peripheralSDK.DisplaySource, error) {
	router.connectionLock.RLock()
	defer router.connectionLock.RUnlock()

	displaySource, connected := router.connectionSource[displaySink]
	if !connected {
		return nil, routing2.ErrDisplaySourceNotConnected
	}

	return displaySource, nil
}

func (router *LocalDisplayRouter) connectDisplaySource(displaySink peripheralSDK.DisplaySink, displaySource peripheralSDK.DisplaySource) error {
	initialDisplayMode, err := displaySource.GetCurrentDisplayMode()
	if err != nil {
		return fmt.Errorf("display source: get display mode: %w", err)
	}

	err = displaySink.SetDisplayMode(*initialDisplayMode)
	if err != nil {
		return fmt.Errorf("display sink: set display mode: %w", err)
	}

	router.connectionLock.Lock()
	defer router.connectionLock.Unlock()

	ctx, cancel := context.WithCancelCause(context.Background())
	router.connectionCancel[displaySink] = cancel

	displayDataChannel := displaySource.DisplayDataChannel(ctx)
	displayControlChannel := displaySource.DisplayControlChannel(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				router.connectionLock.Lock()
				delete(router.connectionCancel, displaySink)
				router.connectionLock.Unlock()
				return
			case dataEvent := <-displayDataChannel:
				err = displaySink.HandleDisplayDataEvent(dataEvent)
				if err != nil {
					fmt.Println("error handling display data event: ", err)
				}
			case controlEvent := <-displayControlChannel:
				err = displaySink.HandleDisplayControlEvent(controlEvent)
				if err != nil {
					fmt.Println("error handling display control event: ", err)
				}
			}
		}
	}()

	return nil
}

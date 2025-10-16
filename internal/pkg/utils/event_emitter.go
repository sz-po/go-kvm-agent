package utils

import (
	"context"
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

type EventEmitterOpt[T any] func(eventEmitter *EventEmitter[T])

type EventEmitterListenerId string

func createRandomEventEmitterListenerId() EventEmitterListenerId {
	return EventEmitterListenerId(uuid.NewString())
}

type EventEmitter[T any] struct {
	listeners     map[EventEmitterListenerId]chan T
	listenersLock *sync.RWMutex

	logger    *slog.Logger
	queueSize int
}

func WithEventEmitterLogger[T any](logger *slog.Logger) EventEmitterOpt[T] {
	return func(eventEmitter *EventEmitter[T]) {
		eventEmitter.logger = logger
	}
}

func WithEventEmitterQueueSize[T any](queueSize int) EventEmitterOpt[T] {
	return func(eventEmitter *EventEmitter[T]) {
		eventEmitter.queueSize = queueSize
	}
}

func NewEventEmitter[T any](opts ...EventEmitterOpt[T]) *EventEmitter[T] {
	eventEmitter := &EventEmitter[T]{
		listeners:     make(map[EventEmitterListenerId]chan T),
		listenersLock: &sync.RWMutex{},
		queueSize:     0,
		logger:        slog.New(slog.DiscardHandler),
	}

	for _, opt := range opts {
		opt(eventEmitter)
	}

	return eventEmitter
}

func (emitter *EventEmitter[T]) Listen(ctx context.Context) <-chan T {
	emitter.listenersLock.Lock()

	listenerId := createRandomEventEmitterListenerId()
	listener := make(chan T, emitter.queueSize)
	emitter.listeners[listenerId] = listener

	emitter.listenersLock.Unlock()

	emitter.logger.Debug("Attached event listener.", slog.String("listenerId", string(listenerId)))

	go func() {
		<-ctx.Done()

		emitter.listenersLock.Lock()
		delete(emitter.listeners, listenerId)
		emitter.listenersLock.Unlock()

		emitter.logger.Debug("Detached event listener.", slog.String("listenerId", string(listenerId)))

		close(listener)

		emitter.logger.Debug("Closed event listener.", slog.String("listenerId", string(listenerId)))
	}()

	return listener
}

func (emitter *EventEmitter[T]) Emit(event T) {
	emitter.listenersLock.RLock()
	defer emitter.listenersLock.RUnlock()

	for _, listener := range emitter.listeners {
		select {
		case listener <- event:
		default:
			//emitter.logger.Warn("Dropping event due to queue full.", slog.String("listenerId", string(listenerId)))
		}
	}
}

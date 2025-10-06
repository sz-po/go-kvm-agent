package routing

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

var (
	// ErrDisplayRouterAlreadyStarted indicates that Start was called more than once without an intervening Stop.
	ErrDisplayRouterAlreadyStarted = errors.New("display router already started")
	// ErrDisplayRouterNotStarted indicates that Stop was called before the router was started.
	ErrDisplayRouterNotStarted = errors.New("display router not started")
	// ErrUnknownSource indicates that the provided source is not registered in the router.
	ErrUnknownSource = errors.New("display source is not registered")
	// ErrUnknownSink indicates that the provided sink is not registered in the router.
	ErrUnknownSink = errors.New("display sink is not registered")
	// ErrSinkAlreadyConnected indicates that the sink is already connected to a source.
	ErrSinkAlreadyConnected = errors.New("display sink already connected to a source")
	// ErrSinkNotConnected indicates that the sink is not connected to any source.
	ErrSinkNotConnected = errors.New("display sink is not connected")
	// ErrDisplaySourceAlreadyRegistered indicates that the same source was registered more than once.
	ErrDisplaySourceAlreadyRegistered = errors.New("display source already registered")
	// ErrDisplaySinkAlreadyRegistered indicates that the same sink was registered more than once.
	ErrDisplaySinkAlreadyRegistered = errors.New("display sink already registered")
	// ErrDisplayOptionNil indicates that a nil peripheral was provided via option configuration.
	ErrDisplayOptionNil = errors.New("display peripheral option cannot be nil")
)

// DisplayRouter routes display data from multiple sources to multiple sinks.
// It handles frame distribution and reacts to display mode changes.
type DisplayRouter struct {
	logger *slog.Logger

	routes     map[peripheral.PeripheralID][]peripheral.DisplaySink
	sinkOwners map[peripheral.PeripheralID]peripheral.PeripheralID
	sourceMap  map[peripheral.PeripheralID]peripheral.DisplaySource
	sinkMap    map[peripheral.PeripheralID]peripheral.DisplaySink

	mu sync.RWMutex

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// DisplayRouterOption configures a DisplayRouter instance during construction.
type DisplayRouterOption func(*DisplayRouter) error

// WithDisplaySource sets the source that provides display events.
func WithDisplaySource(source peripheral.DisplaySource) DisplayRouterOption {
	return func(r *DisplayRouter) error {
		if source == nil {
			return ErrDisplayOptionNil
		}
		sourceID := source.ID()
		if r.sourceMap == nil {
			r.sourceMap = make(map[peripheral.PeripheralID]peripheral.DisplaySource)
		}
		if _, exists := r.sourceMap[sourceID]; exists {
			return ErrDisplaySourceAlreadyRegistered
		}
		r.sourceMap[sourceID] = source
		if r.routes == nil {
			r.routes = make(map[peripheral.PeripheralID][]peripheral.DisplaySink)
		}
		if _, ok := r.routes[sourceID]; !ok {
			r.routes[sourceID] = nil
		}
		return nil
	}
}

// WithDisplaySink appends a sink that will receive routed display events.
func WithDisplaySink(sink peripheral.DisplaySink) DisplayRouterOption {
	return func(r *DisplayRouter) error {
		if sink == nil {
			return ErrDisplayOptionNil
		}
		sinkID := sink.ID()
		if r.sinkMap == nil {
			r.sinkMap = make(map[peripheral.PeripheralID]peripheral.DisplaySink)
		}
		if _, exists := r.sinkMap[sinkID]; exists {
			return ErrDisplaySinkAlreadyRegistered
		}
		r.sinkMap[sinkID] = sink
		return nil
	}
}

// WithLogger injects a custom logger for the router.
func WithLogger(logger *slog.Logger) DisplayRouterOption {
	return func(r *DisplayRouter) error {
		if logger != nil {
			r.logger = logger
		}
		return nil
	}
}

// NewDisplayRouter creates a new DisplayRouter with the given options.
func NewDisplayRouter(opts ...DisplayRouterOption) (*DisplayRouter, error) {
	router := &DisplayRouter{
		logger:     slog.Default(),
		routes:     make(map[peripheral.PeripheralID][]peripheral.DisplaySink),
		sinkOwners: make(map[peripheral.PeripheralID]peripheral.PeripheralID),
		sourceMap:  make(map[peripheral.PeripheralID]peripheral.DisplaySource),
		sinkMap:    make(map[peripheral.PeripheralID]peripheral.DisplaySink),
	}

	for _, opt := range opts {
		if err := opt(router); err != nil {
			return nil, err
		}
	}

	return router, nil
}

// ConnectSink associates the provided sink with the specified source.
// It ensures that a sink can only be connected to one source at a time.
func (r *DisplayRouter) ConnectSink(source peripheral.DisplaySource, sink peripheral.DisplaySink) error {
	if source == nil {
		return ErrUnknownSource
	}
	if sink == nil {
		return ErrUnknownSink
	}

	sourceID := source.ID()
	sinkID := sink.ID()

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, sourceOk := r.sourceMap[sourceID]; !sourceOk {
		return ErrUnknownSource
	}
	registeredSink, sinkOk := r.sinkMap[sinkID]
	if !sinkOk {
		return ErrUnknownSink
	}

	if owner, exists := r.sinkOwners[sinkID]; exists {
		if owner == sourceID {
			return nil
		}
		return ErrSinkAlreadyConnected
	}

	sinks := r.routes[sourceID]
	for _, existing := range sinks {
		if existing.ID() == sinkID {
			r.sinkOwners[sinkID] = sourceID
			return nil
		}
	}

	r.routes[sourceID] = append(sinks, registeredSink)
	r.sinkOwners[sinkID] = sourceID

	r.logger.Info("display sink connected",
		slog.String("sourceId", sourceID.String()),
		slog.String("sinkId", sinkID.String()))

	return nil
}

// DisconnectSink detaches the provided sink from its current source.
func (r *DisplayRouter) DisconnectSink(sink peripheral.DisplaySink) error {
	if sink == nil {
		return ErrUnknownSink
	}

	sinkID := sink.ID()

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, sinkOk := r.sinkMap[sinkID]; !sinkOk {
		return ErrUnknownSink
	}

	ownerID, connected := r.sinkOwners[sinkID]
	if !connected {
		return ErrSinkNotConnected
	}

	sinks := r.routes[ownerID]
	updated := make([]peripheral.DisplaySink, 0, len(sinks))
	removed := false
	for _, existing := range sinks {
		if existing.ID() == sinkID {
			removed = true
			continue
		}
		updated = append(updated, existing)
	}
	if !removed {
		return ErrSinkNotConnected
	}
	r.routes[ownerID] = updated
	delete(r.sinkOwners, sinkID)

	r.logger.Info("display sink disconnected",
		slog.String("sourceId", ownerID.String()),
		slog.String("sinkId", sinkID.String()))

	return nil
}

func (r *DisplayRouter) sinksForSource(sourceID peripheral.PeripheralID) []peripheral.DisplaySink {
	r.mu.RLock()
	sinks := append([]peripheral.DisplaySink(nil), r.routes[sourceID]...)
	r.mu.RUnlock()
	return sinks
}

// Start begins routing display events from source to sinks.
// It subscribes to source channels and distributes events to all sinks.
func (r *DisplayRouter) Start(ctx context.Context) error {
	if r.cancel != nil {
		return ErrDisplayRouterAlreadyStarted
	}

	r.mu.RLock()
	sources := make([]peripheral.DisplaySource, 0, len(r.sourceMap))
	for _, src := range r.sourceMap {
		sources = append(sources, src)
	}
	sourceCount := len(sources)
	sinkCount := len(r.sinkMap)
	r.mu.RUnlock()

	if sourceCount == 0 {
		r.logger.Warn("display router started without sources")
	}

	routerCtx, cancel := context.WithCancel(ctx)
	r.cancel = cancel

	for _, source := range sources {
		dataChannel := source.DataChannel(routerCtx)
		controlChannel := source.ControlChannel(routerCtx)

		// Route data events for the source
		r.wg.Add(1)
		go func(src peripheral.DisplaySource, ch <-chan peripheral.DisplayEvent) {
			defer r.wg.Done()
			r.routeDataEvents(routerCtx, src, ch)
		}(source, dataChannel)

		// Handle control events for the source
		r.wg.Add(1)
		go func(src peripheral.DisplaySource, ch <-chan peripheral.DisplayControlEvent) {
			defer r.wg.Done()
			r.handleControlEvents(routerCtx, src, ch)
		}(source, controlChannel)
	}

	r.logger.Info("display router started",
		slog.Int("sourceCount", sourceCount),
		slog.Int("sinkCount", sinkCount))

	return nil
}

// Stop stops the router and waits for all routing goroutines to finish.
func (r *DisplayRouter) Stop(ctx context.Context) error {
	if r.cancel == nil {
		return ErrDisplayRouterNotStarted
	}

	r.cancel()
	r.cancel = nil

	// Wait for goroutines to finish
	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		r.logger.Info("display router stopped")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// routeDataEvents distributes data events from source to all sinks.
func (r *DisplayRouter) routeDataEvents(ctx context.Context, source peripheral.DisplaySource, dataChannel <-chan peripheral.DisplayEvent) {
	sourceID := source.ID()
	sourceKey := sourceID.String()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-dataChannel:
			if !ok {
				r.logger.Warn("data channel closed",
					slog.String("sourceId", sourceKey))
				return
			}

			sinks := r.sinksForSource(sourceID)
			if len(sinks) == 0 {
				continue
			}

			// Distribute to all sinks
			for _, sink := range sinks {
				sinkID := sink.ID()
				sinkKey := sinkID.String()
				if err := sink.HandleDataEvent(event); err != nil {
					r.logger.Error("failed to handle data event",
						slog.String("error", err.Error()),
						slog.String("sourceId", sourceKey),
						slog.String("sinkId", sinkKey),
						slog.Int("eventType", int(event.Type())))
				}
			}
		}
	}
}

// handleControlEvents processes control events from source.
// It reacts to DisplayModeChangedEvent by updating all sinks.
func (r *DisplayRouter) handleControlEvents(ctx context.Context, source peripheral.DisplaySource, controlChannel <-chan peripheral.DisplayControlEvent) {
	sourceID := source.ID()
	sourceKey := sourceID.String()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-controlChannel:
			if !ok {
				r.logger.Warn("control channel closed",
					slog.String("sourceId", sourceKey))
				return
			}

			switch e := event.(type) {
			case peripheral.DisplayModeChangedEvent:
				r.logger.Info("display mode changed",
					slog.String("sourceId", sourceKey),
					slog.Int("oldWidth", int(e.OldMode.Width)),
					slog.Int("oldHeight", int(e.OldMode.Height)),
					slog.Int("newWidth", int(e.NewMode.Width)),
					slog.Int("newHeight", int(e.NewMode.Height)))

				// Update all sinks with new mode
				sinks := r.sinksForSource(sourceID)
				for _, sink := range sinks {
					sinkID := sink.ID()
					sinkKey := sinkID.String()
					if err := sink.SetDisplayMode(e.NewMode); err != nil {
						r.logger.Error("failed to set display mode",
							slog.String("error", err.Error()),
							slog.String("sourceId", sourceKey),
							slog.String("sinkId", sinkKey))
					}
				}
			}
			// Other control events are ignored - not router's responsibility
		}
	}
}

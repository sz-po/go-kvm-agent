package peripheral

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/codec"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/utils"
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type DisplaySinkAdapterOpt func(*DisplaySinkAdapter)

type DisplaySinkAdapter struct {
	displaySink peripheralSDK.DisplaySink
	serviceId   nodeSDK.ServiceId
	logger      *slog.Logger
}

func WithDisplaySinkAdapterLogger(logger *slog.Logger) DisplaySinkAdapterOpt {
	return func(adapter *DisplaySinkAdapter) {
		adapter.logger = logger
	}
}

func NewDisplaySinkAdapter(displaySink peripheralSDK.DisplaySink, opts ...DisplaySinkAdapterOpt) *DisplaySinkAdapter {
	adapter := &DisplaySinkAdapter{
		displaySink: displaySink,
		serviceId:   DisplaySinkServiceId.WithArgument(string(displaySink.GetId())),
		logger:      slog.New(slog.DiscardHandler),
	}

	for _, opt := range opts {
		opt(adapter)
	}

	adapter.logger = adapter.logger.With(
		slog.String("serviceId", string(adapter.serviceId)),
		slog.String("peripheralId", displaySink.GetId().String()),
	)

	return adapter
}

func (adapter *DisplaySinkAdapter) GetServiceId() nodeSDK.ServiceId {
	return adapter.serviceId
}

func (adapter *DisplaySinkAdapter) Handle(ctx context.Context, stream io.ReadWriteCloser) {
	defer func() {
		_ = stream.Close()
	}()

	jsonCodec := codec.NewJsonCodec(stream)

	var requestHeader api.RequestHeader
	if err := jsonCodec.Decode(&requestHeader); err != nil {
		adapter.logger.Warn("Failed to decode request header.", slog.String("error", err.Error()))
		return
	}

	logger := adapter.logger.With(slog.String("serviceMethodName", string(requestHeader.MethodName)))

	var handleErr error

	switch requestHeader.MethodName {
	case DisplaySinkSetFrameBufferProviderMethod:
		handleErr = utils.HandleServiceRequest(ctx, jsonCodec, adapter.handleSetDisplayFrameBufferProvider)
	case DisplaySinkClearFrameBufferProviderMethod:
		handleErr = utils.HandleServiceRequest(ctx, jsonCodec, adapter.handleClearDisplayFrameBufferProvider)
	default:
		_ = jsonCodec.Encode(&api.ResponseHeader{Error: api.ErrUnsupportedMethod.Error()})
		logger.Warn("Unsupported request method.")
		return
	}

	if handleErr != nil {
		logger.Error("Failed to handle request.", slog.String("error", handleErr.Error()))
		return
	}

	logger.Debug("Request handled successfully.")
}

func (adapter *DisplaySinkAdapter) handleSetDisplayFrameBufferProvider(ctx context.Context, request DisplaySinkSetFrameBufferProviderRequest) (*DisplaySinkSetFrameBufferProviderResponse, error) {
	transport, hasTransport := ctx.Value("transport").(api.Transport)
	if !hasTransport {
		return nil, fmt.Errorf("transport not found in context")
	}

	displaySource := newDisplaySourceClient(transport, request.NodeId, request.Peripheral)
	if err := adapter.displaySink.SetDisplayFrameBufferProvider(displaySource); err != nil {
		return nil, err
	}

	return &DisplaySinkSetFrameBufferProviderResponse{}, nil
}

func (adapter *DisplaySinkAdapter) handleClearDisplayFrameBufferProvider(ctx context.Context, request DisplaySinkClearFrameBufferProviderRequest) (*DisplaySinkClearFrameBufferProviderResponse, error) {
	if err := adapter.displaySink.ClearDisplayFrameBufferProvider(); err != nil {
		return nil, err
	}

	return &DisplaySinkClearFrameBufferProviderResponse{}, nil
}

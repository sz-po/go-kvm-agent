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

type DisplaySourceAdapterOpt func(*DisplaySourceAdapter)

type DisplaySourceAdapter struct {
	displaySource peripheralSDK.DisplaySource
	serviceId     nodeSDK.ServiceId
	logger        *slog.Logger
}

func WithDisplaySourceAdapterLogger(logger *slog.Logger) DisplaySourceAdapterOpt {
	return func(adapter *DisplaySourceAdapter) {
		adapter.logger = logger
	}
}

func NewDisplaySourceAdapter(displaySource peripheralSDK.DisplaySource, opts ...DisplaySourceAdapterOpt) *DisplaySourceAdapter {
	adapter := &DisplaySourceAdapter{
		displaySource: displaySource,
		serviceId:     DisplaySourceServiceId.WithArgument(string(displaySource.GetId())),
		logger:        slog.New(slog.DiscardHandler),
	}

	for _, opt := range opts {
		opt(adapter)
	}

	adapter.logger = adapter.logger.With(
		slog.String("serviceId", string(adapter.serviceId)),
		slog.String("peripheralId", displaySource.GetId().String()),
	)

	return adapter
}

func (adapter *DisplaySourceAdapter) GetServiceId() nodeSDK.ServiceId {
	return adapter.serviceId
}

func (adapter *DisplaySourceAdapter) Handle(ctx context.Context, stream io.ReadWriteCloser) {
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
	case DisplaySourceGetFrameBufferMethod:
		handleErr = adapter.handleGetFrameBuffer(ctx, jsonCodec, stream)
	case DisplaySourceGetDisplayModeMethod:
		handleErr = utils.HandleServiceRequest(ctx, jsonCodec, adapter.handleGetDisplayMode)
	case DisplaySourceGetPixelFormatMethod:
		handleErr = utils.HandleServiceRequest(ctx, jsonCodec, adapter.handleGetPixelFormat)
	case DisplaySourceGetMetricsMethod:
		handleErr = utils.HandleServiceRequest(ctx, jsonCodec, adapter.handleGetMetrics)
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

func (adapter *DisplaySourceAdapter) handleGetFrameBuffer(ctx context.Context, jsonCodec api.Codec, writer io.Writer) error {
	var request DisplaySourceGetFrameBufferRequest
	if err := jsonCodec.Decode(&request); err != nil {
		_ = jsonCodec.Encode(&api.ResponseHeader{Error: api.ErrMalformedRequest.Error()})
		return fmt.Errorf("decode request: %w", err)
	}

	frameBuffer, err := adapter.displaySource.GetDisplayFrameBuffer(ctx)
	if err != nil {
		_ = jsonCodec.Encode(&api.ResponseHeader{Error: err.Error()})
		return fmt.Errorf("get display frame buffer: %w", err)
	}

	defer func() {
		if releaseErr := frameBuffer.Release(); releaseErr != nil {
			adapter.logger.Warn("Failed to release frame buffer.", slog.String("error", releaseErr.Error()))
		}
	}()

	response := &DisplaySourceGetFrameBufferResponse{
		Size: frameBuffer.GetSize(),
	}

	if err := jsonCodec.Encode(&api.ResponseHeader{}); err != nil {
		return fmt.Errorf("encode response header: %w", err)
	}

	if err := jsonCodec.Encode(response); err != nil {
		return fmt.Errorf("encode response: %w", err)
	}

	if _, err := frameBuffer.WriteTo(writer); err != nil {
		return fmt.Errorf("write frame buffer payload: %w", err)
	}

	return nil
}

func (adapter *DisplaySourceAdapter) handleGetDisplayMode(ctx context.Context, request DisplaySourceGetDisplayModeRequest) (*DisplaySourceGetDisplayModeResponse, error) {
	displayMode, err := adapter.displaySource.GetDisplayMode(ctx)
	if err != nil {
		return nil, err
	}

	return &DisplaySourceGetDisplayModeResponse{
		DisplayMode: displayMode,
	}, nil
}

func (adapter *DisplaySourceAdapter) handleGetPixelFormat(ctx context.Context, request DisplaySourceGetPixelFormatRequest) (*DisplaySourceGetPixelFormatResponse, error) {
	pixelFormat, err := adapter.displaySource.GetDisplayPixelFormat(ctx)
	if err != nil {
		return nil, err
	}

	return &DisplaySourceGetPixelFormatResponse{
		PixelFormat: pixelFormat,
	}, nil
}

func (adapter *DisplaySourceAdapter) handleGetMetrics(ctx context.Context, request DisplaySourceGetMetricsRequest) (*DisplaySourceGetMetricsResponse, error) {
	metrics := adapter.displaySource.GetDisplaySourceMetrics()

	return &DisplaySourceGetMetricsResponse{
		Metrics: metrics,
	}, nil
}

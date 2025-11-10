package peripheral

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/codec"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/utils"
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type PeripheralAdapterOpt func(adapter *PeripheralAdapter)

type PeripheralAdapter struct {
	implementation peripheral.Peripheral
	serviceId      nodeSDK.ServiceId
	logger         *slog.Logger
}

func NewPeripheralAdapter(implementation peripheral.Peripheral, opts ...PeripheralAdapterOpt) *PeripheralAdapter {
	service := &PeripheralAdapter{
		implementation: implementation,
		serviceId:      PeripheralServiceId.WithArgument(string(implementation.GetId())),
		logger:         slog.New(slog.DiscardHandler),
	}

	for _, opt := range opts {
		opt(service)
	}

	service.logger = service.logger.With(slog.String("serviceId", string(service.serviceId)))

	return service
}

func (adapter *PeripheralAdapter) GetServiceId() nodeSDK.ServiceId {
	return adapter.serviceId
}

func (adapter *PeripheralAdapter) Handle(stream io.ReadWriteCloser) {
	defer func() {
		_ = stream.Close()
	}()

	requestCtx, requestCtxCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer requestCtxCancel()

	jsonCodec := codec.NewJsonCodec(stream)

	var requestHeader api.RequestHeader
	err := jsonCodec.Decode(&requestHeader)
	if err != nil {
		adapter.logger.Warn("Failed to decode request header.",
			slog.String("error", err.Error()),
		)
		return
	}

	logger := adapter.logger.With(slog.String("serviceMethodName", string(requestHeader.MethodName)))

	switch requestHeader.MethodName {
	case PeripheralTerminateMethod:
		err := utils.HandleServiceRequest(requestCtx, jsonCodec, adapter.handleTerminate)
		if err != nil {
			logger.Error("Failed to handle request.", slog.String("error", err.Error()))
			return
		}
	default:
		jsonCodec.Encode(&api.ResponseHeader{
			Error: api.ErrUnsupportedMethod,
		})
		logger.Warn("Unsupported request method.")
		return
	}

	logger.Debug("Request handled successfully.")
}

func (adapter *PeripheralAdapter) handleTerminate(ctx context.Context, request PeripheralTerminateRequest) (*PeripheralTerminateResponse, error) {
	err := adapter.implementation.Terminate(ctx)
	if err != nil {
		return nil, err
	}

	return &PeripheralTerminateResponse{}, nil
}

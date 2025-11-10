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
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type RepositoryAdapterOpt func(*RepositoryAdapter)

type RepositoryAdapter struct {
	repository peripheralSDK.Repository
	serviceId  nodeSDK.ServiceId
	logger     *slog.Logger
}

func NewRepositoryAdapter(repository peripheralSDK.Repository, opts ...RepositoryAdapterOpt) *RepositoryAdapter {
	adapter := &RepositoryAdapter{
		repository: repository,
		serviceId:  PeripheralRepositoryServiceId,
		logger:     slog.New(slog.DiscardHandler),
	}

	for _, opt := range opts {
		opt(adapter)
	}

	adapter.logger = adapter.logger.With(slog.String("serviceId", string(adapter.serviceId)))

	return adapter
}

func (adapter *RepositoryAdapter) GetServiceId() nodeSDK.ServiceId {
	return adapter.serviceId
}

func (adapter *RepositoryAdapter) Handle(stream io.ReadWriteCloser) {
	defer func() {
		_ = stream.Close()
	}()

	requestCtx, requestCtxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer requestCtxCancel()

	jsonCodec := codec.NewJsonCodec(stream)

	var requestHeader api.RequestHeader
	if err := jsonCodec.Decode(&requestHeader); err != nil {
		adapter.logger.Warn("Failed to decode request header.",
			slog.String("error", err.Error()),
		)
		return
	}

	logger := adapter.logger.With(slog.String("serviceMethodName", string(requestHeader.MethodName)))

	var handleErr error

	switch requestHeader.MethodName {
	case RepositoryGetPeripheralByIdMethod:
		handleErr = utils.HandleServiceRequest(requestCtx, jsonCodec, adapter.handleGetPeripheralById)
	case RepositoryGetPeripheralByNameMethod:
		handleErr = utils.HandleServiceRequest(requestCtx, jsonCodec, adapter.handleGetPeripheralByName)
	case RepositoryGetAllPeripheralsMethod:
		handleErr = utils.HandleServiceRequest(requestCtx, jsonCodec, adapter.handleGetAllPeripherals)
	case RepositoryGetAllDisplaySourcesMethod:
		handleErr = utils.HandleServiceRequest(requestCtx, jsonCodec, adapter.handleGetAllDisplaySources)
	case RepositoryGetAllDisplaySinksMethod:
		handleErr = utils.HandleServiceRequest(requestCtx, jsonCodec, adapter.handleGetAllDisplaySinks)
	default:
		_ = jsonCodec.Encode(&api.ResponseHeader{
			Error: api.ErrUnsupportedMethod,
		})
		logger.Warn("Unsupported request method.")
		return
	}

	if handleErr != nil {
		logger.Error("Failed to handle request.", slog.String("error", handleErr.Error()))
		return
	}

	logger.Debug("Request handled successfully.")
}

func (adapter *RepositoryAdapter) handleGetPeripheralById(ctx context.Context, request RepositoryGetPeripheralByIdRequest) (*RepositoryGetPeripheralByIdResponse, error) {
	peripheral, err := adapter.repository.GetPeripheralById(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	return &RepositoryGetPeripheralByIdResponse{
		Peripheral: createPeripheralDescriptor(peripheral),
	}, nil
}

func (adapter *RepositoryAdapter) handleGetPeripheralByName(ctx context.Context, request RepositoryGetPeripheralByNameRequest) (*RepositoryGetPeripheralByNameResponse, error) {
	peripheral, err := adapter.repository.GetPeripheralByName(ctx, request.Name)
	if err != nil {
		return nil, err
	}

	return &RepositoryGetPeripheralByNameResponse{
		Peripheral: createPeripheralDescriptor(peripheral),
	}, nil
}

func (adapter *RepositoryAdapter) handleGetAllPeripherals(ctx context.Context, request RepositoryGetAllPeripheralsRequest) (*RepositoryGetAllPeripheralsResponse, error) {
	peripherals, err := adapter.repository.GetAllPeripherals(ctx)
	if err != nil {
		return nil, err
	}

	repositoryPeripherals := make([]peripheralDescriptor, 0, len(peripherals))
	for _, peripheral := range peripherals {
		repositoryPeripherals = append(repositoryPeripherals, createPeripheralDescriptor(peripheral))
	}

	return &RepositoryGetAllPeripheralsResponse{
		Peripherals: repositoryPeripherals,
	}, nil
}

func (adapter *RepositoryAdapter) handleGetAllDisplaySources(ctx context.Context, request RepositoryGetAllDisplaySourcesRequest) (*RepositoryGetAllDisplaySourcesResponse, error) {
	displaySources, err := adapter.repository.GetAllDisplaySources(ctx)
	if err != nil {
		return nil, err
	}

	repositoryPeripherals := make([]peripheralDescriptor, 0, len(displaySources))
	for _, displaySource := range displaySources {
		repositoryPeripherals = append(repositoryPeripherals, createPeripheralDescriptor(displaySource))
	}

	return &RepositoryGetAllDisplaySourcesResponse{
		DisplaySources: repositoryPeripherals,
	}, nil
}

func (adapter *RepositoryAdapter) handleGetAllDisplaySinks(ctx context.Context, request RepositoryGetAllDisplaySinksRequest) (*RepositoryGetAllDisplaySinksResponse, error) {
	displaySinks, err := adapter.repository.GetAllDisplaySinks(ctx)
	if err != nil {
		return nil, err
	}

	repositoryPeripherals := make([]peripheralDescriptor, 0, len(displaySinks))
	for _, displaySink := range displaySinks {
		repositoryPeripherals = append(repositoryPeripherals, createPeripheralDescriptor(displaySink))
	}

	return &RepositoryGetAllDisplaySinksResponse{
		DisplaySinks: repositoryPeripherals,
	}, nil
}

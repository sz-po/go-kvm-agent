package node

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/codec"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/utils"
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
)

type NodeAdapterOpt func(*NodeAdapter)

type NodeAdapter struct {
	implementation nodeSDK.Node
	serviceId      nodeSDK.ServiceId
	logger         *slog.Logger
}

func NewNodeAdapter(implementation nodeSDK.Node, opts ...NodeAdapterOpt) *NodeAdapter {
	service := &NodeAdapter{
		implementation: implementation,
		serviceId:      NodeServiceId,
		logger:         slog.New(slog.DiscardHandler),
	}

	for _, opt := range opts {
		opt(service)
	}

	service.logger = service.logger.With(slog.String("serviceId", string(service.serviceId)))

	return service
}

func (adapter *NodeAdapter) GetServiceId() nodeSDK.ServiceId {
	return adapter.serviceId
}

func (adapter *NodeAdapter) Handle(stream io.ReadWriteCloser) {
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
	case NodeGetIdMethod:
		err := utils.HandleServiceRequest(requestCtx, jsonCodec, adapter.handleGetId)
		if err != nil {
			logger.Error("Failed to handle request.", slog.String("error", err.Error()))
			return
		}
	case NodeGetHostNameMethod:
		err := utils.HandleServiceRequest(requestCtx, jsonCodec, adapter.handleGetHostName)
		if err != nil {
			logger.Error("Failed to handle request.", slog.String("error", err.Error()))
			return
		}
	case NodeGetUptimeMethod:
		err := utils.HandleServiceRequest(requestCtx, jsonCodec, adapter.handleGetUptime)
		if err != nil {
			logger.Error("Failed to handle request.", slog.String("error", err.Error()))
			return
		}
	case NodeGetPlatformMethod:
		err := utils.HandleServiceRequest(requestCtx, jsonCodec, adapter.handleGetPlatform)
		if err != nil {
			logger.Error("Failed to handle request.", slog.String("error", err.Error()))
			return
		}
	case NodeGetRoleMethod:
		err := utils.HandleServiceRequest(requestCtx, jsonCodec, adapter.handleGetRole)
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

func (adapter *NodeAdapter) handleGetId(ctx context.Context, request NodeGetIdRequest) (*NodeGetIdResponse, error) {
	id, err := adapter.implementation.GetId(ctx)
	if err != nil {
		return nil, err
	}

	return &NodeGetIdResponse{
		Id: *id,
	}, nil
}

func (adapter *NodeAdapter) handleGetHostName(ctx context.Context, request NodeGetHostNameRequest) (*NodeGetHostNameResponse, error) {
	hostName, err := adapter.implementation.GetHostName(ctx)
	if err != nil {
		return nil, err
	}

	return &NodeGetHostNameResponse{
		HostName: *hostName,
	}, nil
}

func (adapter *NodeAdapter) handleGetUptime(ctx context.Context, request NodeGetUptimeRequest) (*NodeGetUptimeResponse, error) {
	uptime, err := adapter.implementation.GetUptime(ctx)
	if err != nil {
		return nil, err
	}

	return &NodeGetUptimeResponse{
		Uptime: api.Duration(*uptime),
	}, nil
}

func (adapter *NodeAdapter) handleGetPlatform(ctx context.Context, request NodeGetPlatformRequest) (*NodeGetPlatformResponse, error) {
	platform, err := adapter.implementation.GetPlatform(ctx)
	if err != nil {
		return nil, err
	}

	return &NodeGetPlatformResponse{
		Platform: *platform,
	}, nil
}

func (adapter *NodeAdapter) handleGetRole(ctx context.Context, request NodeGetRoleRequest) (*NodeGetRoleResponse, error) {
	roles, err := adapter.implementation.GetRoles(ctx)
	if err != nil {
		return nil, err
	}

	return &NodeGetRoleResponse{
		Roles: roles,
	}, nil
}

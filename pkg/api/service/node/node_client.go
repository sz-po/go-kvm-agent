package node

import (
	"context"
	"fmt"
	"time"

	apiSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/api"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/codec"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/service/node/peripheral"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/utils"
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type NodeClientOpt func(*NodeClient)

type NodeClient struct {
	nodeId    nodeSDK.NodeId
	transport apiSDK.Transport
}

var _ nodeSDK.Node = (*NodeClient)(nil)

func NewNodeClient(nodeId nodeSDK.NodeId, transport apiSDK.Transport, opts ...NodeClientOpt) *NodeClient {
	client := &NodeClient{
		nodeId:    nodeId,
		transport: transport,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (client *NodeClient) GetId(ctx context.Context) (*nodeSDK.NodeId, error) {
	stream, err := client.transport.OpenServiceStream(ctx, NodeServiceId, client.nodeId)
	if err != nil {
		return nil, fmt.Errorf("open stream: %w", err)
	}

	defer func() {
		_ = stream.Close()
	}()

	jsonCodec := codec.NewJsonCodec(stream)

	response, err := utils.HandleClientRequest[NodeGetIdRequest, NodeGetIdResponse](ctx, jsonCodec, NodeGetIdMethod, NodeGetIdRequest{})
	if err != nil {
		return nil, fmt.Errorf("call: %w", err)
	}

	return &response.Id, nil
}

func (client *NodeClient) GetHostName(ctx context.Context) (hostName *string, err error) {
	stream, err := client.transport.OpenServiceStream(ctx, NodeServiceId, client.nodeId)
	if err != nil {
		return nil, fmt.Errorf("open stream: %w", err)
	}

	defer func() {
		_ = stream.Close()
	}()

	jsonCodec := codec.NewJsonCodec(stream)

	response, err := utils.HandleClientRequest[NodeGetHostNameRequest, NodeGetHostNameResponse](ctx, jsonCodec, NodeGetHostNameMethod, NodeGetHostNameRequest{})
	if err != nil {
		return nil, fmt.Errorf("call: %w", err)
	}

	return &response.HostName, nil
}

func (client *NodeClient) GetUptime(ctx context.Context) (*time.Duration, error) {
	stream, err := client.transport.OpenServiceStream(ctx, NodeServiceId, client.nodeId)
	if err != nil {
		return nil, fmt.Errorf("open stream: %w", err)
	}

	defer func() {
		_ = stream.Close()
	}()

	jsonCodec := codec.NewJsonCodec(stream)

	response, err := utils.HandleClientRequest[NodeGetUptimeRequest, NodeGetUptimeResponse](ctx, jsonCodec, NodeGetUptimeMethod, NodeGetUptimeRequest{})
	if err != nil {
		return nil, fmt.Errorf("call: %w", err)
	}

	uptime := time.Duration(response.Uptime)
	return &uptime, nil
}

func (client *NodeClient) GetPlatform(ctx context.Context) (*nodeSDK.NodePlatform, error) {
	stream, err := client.transport.OpenServiceStream(ctx, NodeServiceId, client.nodeId)
	if err != nil {
		return nil, fmt.Errorf("open stream: %w", err)
	}

	defer func() {
		_ = stream.Close()
	}()

	jsonCodec := codec.NewJsonCodec(stream)

	response, err := utils.HandleClientRequest[NodeGetPlatformRequest, NodeGetPlatformResponse](ctx, jsonCodec, NodeGetPlatformMethod, NodeGetPlatformRequest{})
	if err != nil {
		return nil, fmt.Errorf("call: %w", err)
	}

	return &response.Platform, nil
}

func (client *NodeClient) GetRoles(ctx context.Context) ([]nodeSDK.NodeRole, error) {
	stream, err := client.transport.OpenServiceStream(ctx, NodeServiceId, client.nodeId)
	if err != nil {
		return nil, fmt.Errorf("open stream: %w", err)
	}

	defer func() {
		_ = stream.Close()
	}()

	jsonCodec := codec.NewJsonCodec(stream)

	response, err := utils.HandleClientRequest[NodeGetRoleRequest, NodeGetRoleResponse](ctx, jsonCodec, NodeGetRoleMethod, NodeGetRoleRequest{})
	if err != nil {
		return nil, fmt.Errorf("call: %w", err)
	}

	return response.Roles, nil
}

func (client *NodeClient) Peripheral() (*peripheral.PeripheralClient, error) {
	panic("not implemented")
}

func (client *NodeClient) Peripherals() peripheralSDK.Repository {
	return peripheral.NewRepositoryClient(client.nodeId, client.transport)
}

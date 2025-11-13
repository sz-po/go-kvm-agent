package peripheral

import (
	"context"
	"fmt"

	apiSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/api"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/codec"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/utils"
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type RepositoryClientOpt func(*RepositoryClient)

type RepositoryClient struct {
	nodeId    nodeSDK.NodeId
	transport apiSDK.Transport
}

var _ peripheralSDK.Repository = (*RepositoryClient)(nil)

func NewRepositoryClient(nodeId nodeSDK.NodeId, transport apiSDK.Transport, opts ...RepositoryClientOpt) *RepositoryClient {
	client := &RepositoryClient{
		nodeId:    nodeId,
		transport: transport,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (client *RepositoryClient) GetPeripheralById(ctx context.Context, id peripheralSDK.Id) (peripheralSDK.Peripheral, error) {
	stream, err := client.transport.OpenServiceStream(ctx, PeripheralRepositoryServiceId, client.nodeId)
	if err != nil {
		return nil, fmt.Errorf("open stream: %w", err)
	}
	defer func() {
		_ = stream.Close()
	}()

	jsonCodec := codec.NewJsonCodec(stream)

	response, err := utils.HandleClientRequest[RepositoryGetPeripheralByIdRequest, RepositoryGetPeripheralByIdResponse](ctx, jsonCodec, RepositoryGetPeripheralByIdMethod, RepositoryGetPeripheralByIdRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("call %s: %w", RepositoryGetPeripheralByIdMethod, err)
	}

	return client.wrapPeripheral(response.Peripheral), nil
}

func (client *RepositoryClient) GetPeripheralByName(ctx context.Context, name peripheralSDK.Name) (peripheralSDK.Peripheral, error) {
	stream, err := client.transport.OpenServiceStream(ctx, PeripheralRepositoryServiceId, client.nodeId)
	if err != nil {
		return nil, fmt.Errorf("open stream: %w", err)
	}
	defer func() {
		_ = stream.Close()
	}()

	jsonCodec := codec.NewJsonCodec(stream)

	response, err := utils.HandleClientRequest[RepositoryGetPeripheralByNameRequest, RepositoryGetPeripheralByNameResponse](ctx, jsonCodec, RepositoryGetPeripheralByNameMethod, RepositoryGetPeripheralByNameRequest{Name: name})
	if err != nil {
		return nil, fmt.Errorf("call %s: %w", RepositoryGetPeripheralByNameMethod, err)
	}

	return client.wrapPeripheral(response.Peripheral), nil
}

func (client *RepositoryClient) GetAllPeripherals(ctx context.Context) ([]peripheralSDK.Peripheral, error) {
	stream, err := client.transport.OpenServiceStream(ctx, PeripheralRepositoryServiceId, client.nodeId)
	if err != nil {
		return nil, fmt.Errorf("open stream: %w", err)
	}
	defer func() {
		_ = stream.Close()
	}()

	jsonCodec := codec.NewJsonCodec(stream)

	response, err := utils.HandleClientRequest[RepositoryGetAllPeripheralsRequest, RepositoryGetAllPeripheralsResponse](ctx, jsonCodec, RepositoryGetAllPeripheralsMethod, RepositoryGetAllPeripheralsRequest{})
	if err != nil {
		return nil, fmt.Errorf("call %s: %w", RepositoryGetAllPeripheralsMethod, err)
	}

	peripherals := make([]peripheralSDK.Peripheral, 0, len(response.Peripherals))
	for _, descriptor := range response.Peripherals {
		peripherals = append(peripherals, client.wrapPeripheral(descriptor))
	}

	return peripherals, nil
}

func (client *RepositoryClient) wrapPeripheral(descriptor peripheralDescriptor) peripheralSDK.Peripheral {
	return newPeripheralClient(client.transport, client.nodeId, descriptor)
}

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

type PeripheralClientOpt func(*PeripheralClient)

type PeripheralClient struct {
	nodeId               nodeSDK.NodeId
	serviceId            nodeSDK.ServiceId
	peripheralDescriptor peripheralDescriptor

	transport apiSDK.Transport
}

var _ peripheralSDK.Peripheral = (*PeripheralClient)(nil)

func newPeripheralClient(transport apiSDK.Transport, nodeId nodeSDK.NodeId, peripheralDescriptor peripheralDescriptor, opts ...PeripheralClientOpt) *PeripheralClient {
	client := &PeripheralClient{
		nodeId:               nodeId,
		serviceId:            PeripheralServiceId.WithArgument(string(peripheralDescriptor.Id)),
		peripheralDescriptor: peripheralDescriptor,

		transport: transport,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (client *PeripheralClient) GetId() peripheralSDK.Id {
	return client.peripheralDescriptor.Id
}

func (client *PeripheralClient) GetName() peripheralSDK.Name {
	return client.peripheralDescriptor.Name
}

func (client *PeripheralClient) GetCapabilities() []peripheralSDK.PeripheralCapability {
	return client.peripheralDescriptor.Capabilities
}

func (client *PeripheralClient) Terminate(ctx context.Context) error {
	stream, err := client.transport.OpenServiceStream(ctx, client.serviceId, client.nodeId)
	if err != nil {
		return fmt.Errorf("open stream: %w", err)
	}

	defer func() {
		_ = stream.Close()
	}()

	jsonCodec := codec.NewJsonCodec(stream)

	_, err = utils.HandleClientRequest[PeripheralTerminateRequest, PeripheralTerminateResponse](ctx, jsonCodec, PeripheralTerminateMethod, PeripheralTerminateRequest{})
	if err != nil {
		return fmt.Errorf("call: %w", err)
	}

	return nil
}

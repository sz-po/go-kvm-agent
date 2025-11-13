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

type DisplaySinkClient struct {
	nodeId           nodeSDK.NodeId
	serviceId        nodeSDK.ServiceId
	transport        apiSDK.Transport
	peripheralClient *PeripheralClient
}

var _ peripheralSDK.DisplaySink = (*DisplaySinkClient)(nil)

func newDisplaySinkClient(transport apiSDK.Transport, nodeId nodeSDK.NodeId, descriptor peripheralDescriptor) *DisplaySinkClient {
	return &DisplaySinkClient{
		nodeId:           nodeId,
		serviceId:        DisplaySinkServiceId.WithArgument(string(descriptor.Id)),
		transport:        transport,
		peripheralClient: newPeripheralClient(transport, nodeId, descriptor),
	}
}

func AsDisplaySink(peripheralClient *PeripheralClient) *DisplaySinkClient {
	return &DisplaySinkClient{
		nodeId:           peripheralClient.nodeId,
		serviceId:        DisplaySinkServiceId.WithArgument(string(peripheralClient.peripheralDescriptor.Id)),
		transport:        peripheralClient.transport,
		peripheralClient: peripheralClient,
	}
}

func (client *DisplaySinkClient) GetId() peripheralSDK.Id {
	return client.peripheralClient.GetId()
}

func (client *DisplaySinkClient) GetName() peripheralSDK.Name {
	return client.peripheralClient.GetName()
}

func (client *DisplaySinkClient) GetCapabilities() []peripheralSDK.PeripheralCapability {
	return client.peripheralClient.GetCapabilities()
}

func (client *DisplaySinkClient) Terminate(ctx context.Context) error {
	return client.peripheralClient.Terminate(ctx)
}

func (client *DisplaySinkClient) SetDisplayFrameBufferProvider(provider peripheralSDK.DisplayFrameBufferProvider) error {
	displaySourceClient, isDisplaySourceClient := provider.(*DisplaySourceClient)
	if !isDisplaySourceClient {
		return fmt.Errorf("unsupported display frame buffer provider type %T", provider)
	}

	ctx := context.Background()

	stream, err := client.transport.OpenServiceStream(ctx, client.serviceId, client.nodeId)
	if err != nil {
		return fmt.Errorf("open stream: %w", err)
	}
	defer func() {
		_ = stream.Close()
	}()

	jsonCodec := codec.NewJsonCodec(stream)

	_, err = utils.HandleClientRequest[DisplaySinkSetFrameBufferProviderRequest, DisplaySinkSetFrameBufferProviderResponse](
		ctx,
		jsonCodec,
		DisplaySinkSetFrameBufferProviderMethod,
		DisplaySinkSetFrameBufferProviderRequest{
			NodeId:     displaySourceClient.nodeId,
			Peripheral: displaySourceClient.peripheralClient.peripheralDescriptor,
		},
	)
	if err != nil {
		return fmt.Errorf("call %s: %w", DisplaySinkSetFrameBufferProviderMethod, err)
	}

	return nil
}

func (client *DisplaySinkClient) ClearDisplayFrameBufferProvider() error {
	ctx := context.Background()

	stream, err := client.transport.OpenServiceStream(ctx, client.serviceId, client.nodeId)
	if err != nil {
		return fmt.Errorf("open stream: %w", err)
	}
	defer func() {
		_ = stream.Close()
	}()

	jsonCodec := codec.NewJsonCodec(stream)

	if _, err := utils.HandleClientRequest[DisplaySinkClearFrameBufferProviderRequest, DisplaySinkClearFrameBufferProviderResponse](
		ctx,
		jsonCodec,
		DisplaySinkClearFrameBufferProviderMethod,
		DisplaySinkClearFrameBufferProviderRequest{},
	); err != nil {
		return fmt.Errorf("call %s: %w", DisplaySinkClearFrameBufferProviderMethod, err)
	}

	return nil
}

package peripheral

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/memory"
	apiSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/api"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/codec"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/utils"
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type DisplaySourceClient struct {
	nodeId           nodeSDK.NodeId
	serviceId        nodeSDK.ServiceId
	transport        apiSDK.Transport
	peripheralClient *PeripheralClient
}

var _ peripheralSDK.DisplaySource = (*DisplaySourceClient)(nil)

func newDisplaySourceClient(transport apiSDK.Transport, nodeId nodeSDK.NodeId, descriptor peripheralDescriptor) *DisplaySourceClient {
	return &DisplaySourceClient{
		nodeId:           nodeId,
		serviceId:        DisplaySourceServiceId.WithArgument(string(descriptor.Id)),
		transport:        transport,
		peripheralClient: newPeripheralClient(transport, nodeId, descriptor),
	}
}

func AsDisplaySource(peripheralClient *PeripheralClient) *DisplaySourceClient {
	return &DisplaySourceClient{
		nodeId:           peripheralClient.nodeId,
		serviceId:        DisplaySourceServiceId.WithArgument(string(peripheralClient.peripheralDescriptor.Id)),
		transport:        peripheralClient.transport,
		peripheralClient: peripheralClient,
	}
}

func (client *DisplaySourceClient) GetId() peripheralSDK.Id {
	return client.peripheralClient.GetId()
}

func (client *DisplaySourceClient) GetName() peripheralSDK.Name {
	return client.peripheralClient.GetName()
}

func (client *DisplaySourceClient) GetCapabilities() []peripheralSDK.PeripheralCapability {
	return client.peripheralClient.GetCapabilities()
}

func (client *DisplaySourceClient) Terminate(ctx context.Context) error {
	return client.peripheralClient.Terminate(ctx)
}

func (client *DisplaySourceClient) GetDisplayFrameBuffer(ctx context.Context) (*peripheralSDK.DisplayFrameBuffer, error) {
	memoryPool, err := memory.DefaultMemoryPoolProvider()
	if err != nil {
		return nil, fmt.Errorf("get memory pool provider: %w", err)
	}

	stream, err := client.transport.OpenServiceStream(ctx, client.serviceId, client.nodeId)
	if err != nil {
		return nil, fmt.Errorf("open stream: %w", err)
	}
	defer func() {
		_ = stream.Close()
	}()

	jsonCodec := codec.NewJsonCodec(stream)

	requestHeader := &apiSDK.RequestHeader{MethodName: DisplaySourceGetFrameBufferMethod}
	if err := jsonCodec.Encode(requestHeader); err != nil {
		return nil, fmt.Errorf("encode request header: %w", err)
	}

	if err := jsonCodec.Encode(DisplaySourceGetFrameBufferRequest{}); err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	bufferedReader := bufio.NewReader(stream)

	responseHeaderLine, err := bufferedReader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("read response header: %w", err)
	}

	responseHeaderLine = bytes.TrimSpace(responseHeaderLine)
	if len(responseHeaderLine) == 0 {
		return nil, fmt.Errorf("read response header: empty payload")
	}

	var responseHeader apiSDK.ResponseHeader
	if err := json.Unmarshal(responseHeaderLine, &responseHeader); err != nil {
		return nil, fmt.Errorf("decode response header: %w", err)
	}
	if len(responseHeader.Error) > 0 {
		return nil, fmt.Errorf("remote error: %s", responseHeader.Error)
	}

	responseBodyLine, err := bufferedReader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	responseBodyLine = bytes.TrimSpace(responseBodyLine)
	if len(responseBodyLine) == 0 {
		return nil, fmt.Errorf("read response body: empty payload")
	}

	var response DisplaySourceGetFrameBufferResponse
	if err := json.Unmarshal(responseBodyLine, &response); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}
	if response.Size <= 0 {
		return nil, fmt.Errorf("invalid frame buffer size %d", response.Size)
	}

	memoryBuffer, err := memoryPool.Borrow(response.Size)
	if err != nil {
		return nil, fmt.Errorf("borrow memory buffer: %w", err)
	}

	if _, err := io.CopyN(memoryBuffer, bufferedReader, int64(response.Size)); err != nil {
		_ = memoryBuffer.Release()
		return nil, fmt.Errorf("read frame buffer payload: %w", err)
	}

	frameBuffer := peripheralSDK.NewDisplayFrameBuffer(memoryBuffer)

	return frameBuffer, nil
}

func (client *DisplaySourceClient) GetDisplayMode(ctx context.Context) (*peripheralSDK.DisplayMode, error) {
	stream, err := client.transport.OpenServiceStream(ctx, client.serviceId, client.nodeId)
	if err != nil {
		return nil, fmt.Errorf("open stream: %w", err)
	}
	defer func() {
		_ = stream.Close()
	}()

	jsonCodec := codec.NewJsonCodec(stream)

	response, err := utils.HandleClientRequest[DisplaySourceGetDisplayModeRequest, DisplaySourceGetDisplayModeResponse](
		ctx,
		jsonCodec,
		DisplaySourceGetDisplayModeMethod,
		DisplaySourceGetDisplayModeRequest{},
	)
	if err != nil {
		return nil, fmt.Errorf("call %s: %w", DisplaySourceGetDisplayModeMethod, err)
	}

	return response.DisplayMode, nil
}

func (client *DisplaySourceClient) GetDisplayPixelFormat(ctx context.Context) (*peripheralSDK.DisplayPixelFormat, error) {
	stream, err := client.transport.OpenServiceStream(ctx, client.serviceId, client.nodeId)
	if err != nil {
		return nil, fmt.Errorf("open stream: %w", err)
	}
	defer func() {
		_ = stream.Close()
	}()

	jsonCodec := codec.NewJsonCodec(stream)

	response, err := utils.HandleClientRequest[DisplaySourceGetPixelFormatRequest, DisplaySourceGetPixelFormatResponse](
		ctx,
		jsonCodec,
		DisplaySourceGetPixelFormatMethod,
		DisplaySourceGetPixelFormatRequest{},
	)
	if err != nil {
		return nil, fmt.Errorf("call %s: %w", DisplaySourceGetPixelFormatMethod, err)
	}

	return response.PixelFormat, nil
}

func (client *DisplaySourceClient) GetDisplaySourceMetrics() peripheralSDK.DisplaySourceMetrics {
	ctx := context.Background()

	stream, err := client.transport.OpenServiceStream(ctx, client.serviceId, client.nodeId)
	if err != nil {
		return peripheralSDK.DisplaySourceMetrics{}
	}
	defer func() {
		_ = stream.Close()
	}()

	jsonCodec := codec.NewJsonCodec(stream)

	response, err := utils.HandleClientRequest[DisplaySourceGetMetricsRequest, DisplaySourceGetMetricsResponse](
		ctx,
		jsonCodec,
		DisplaySourceGetMetricsMethod,
		DisplaySourceGetMetricsRequest{},
	)
	if err != nil {
		return peripheralSDK.DisplaySourceMetrics{}
	}

	return response.Metrics
}

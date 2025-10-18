package display_source

import (
	"fmt"
	"net/http"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

const (
	MetricsEndpointName = "metrics"
)

type GetMetricsRequestPath struct {
	MachineIdentifier    machine.MachineIdentifier
	PeripheralIdentifier peripheral.PeripheralIdentifier
}

func (path GetMetricsRequestPath) String() (*string, error) {
	machineIdentifier, err := path.MachineIdentifier.String()
	if err != nil {
		return nil, err
	}

	peripheralIdentifier, err := path.PeripheralIdentifier.String()
	if err != nil {
		return nil, err
	}

	result := fmt.Sprintf("/%s/%s/%s/%s/%s/%s", machine.EndpointName, *machineIdentifier, peripheral.EndpointName, *peripheralIdentifier, EndpointName, MetricsEndpointName)

	return &result, nil
}

func (path GetMetricsRequestPath) Params() (*transport.PathParams, error) {
	machineIdentifier, err := path.MachineIdentifier.String()
	if err != nil {
		return nil, err
	}

	peripheralIdentifier, err := path.PeripheralIdentifier.String()
	if err != nil {
		return nil, err
	}

	return &transport.PathParams{
		machine.MachineIdentifierPathFieldName:       *machineIdentifier,
		peripheral.PeripheralIdentifierPathFieldName: *peripheralIdentifier,
	}, nil
}

type GetMetricsRequest struct {
	Path GetMetricsRequestPath
}

func (request *GetMetricsRequest) Request() (*transport.Request, error) {
	path, err := request.Path.String()
	if err != nil {
		return nil, fmt.Errorf("path: %w", err)
	}

	pathParams, err := request.Path.Params()
	if err != nil {
		return nil, fmt.Errorf("path params: %w", err)
	}

	return &transport.Request{
		Method:    http.MethodGet,
		Path:      *path,
		PathParam: *pathParams,
	}, nil
}

type GetMetricsResponseBody struct {
	Metrics peripheralSDK.DisplaySourceMetrics `json:"metrics"`
}

func ParseGetMetricsResponseBody(input any) (*GetMetricsResponseBody, error) {
	body, err := transport.UnmarshalResponseBody[GetMetricsResponseBody](input)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return body, nil
}

type GetMetricsResponse struct {
	Body GetMetricsResponseBody
}

func ParseGetMetricsResponse(response transport.Response) (*GetMetricsResponse, error) {
	body, err := ParseGetMetricsResponseBody(response.Body)
	if err != nil {
		return nil, fmt.Errorf("parse body: %w", err)
	}

	return &GetMetricsResponse{
		Body: *body,
	}, nil
}

func (response *GetMetricsResponse) Response() transport.Response {
	return transport.Response{
		StatusCode: http.StatusOK,
		Header: transport.Header{
			transport.HeaderContentType: "application/json",
		},
		Body: response.Body,
	}
}

func ParseGetMetricsRequest(request transport.Request) (*GetMetricsRequest, error) {
	path, err := ParseGetMetricsRequestPath(request.PathParam)
	if err != nil {
		return nil, fmt.Errorf("parse path: %w", err)
	}

	return &GetMetricsRequest{
		Path: *path,
	}, nil
}

func ParseGetMetricsRequestPath(path transport.PathParams) (*GetMetricsRequestPath, error) {
	err := path.Require(machine.MachineIdentifierPathFieldName, peripheral.PeripheralIdentifierPathFieldName)
	if err != nil {
		return nil, err
	}

	machineIdentifier, err := machine.ParseMachineIdentifier(path[machine.MachineIdentifierPathFieldName])
	if err != nil {
		return nil, fmt.Errorf("parse machine identifier: %w", err)
	}

	peripheralIdentifier, err := peripheral.ParsePeripheralIdentifier(path[peripheral.PeripheralIdentifierPathFieldName])
	if err != nil {
		return nil, fmt.Errorf("parse peripheral identifier: %w", err)
	}

	return &GetMetricsRequestPath{
		MachineIdentifier:    *machineIdentifier,
		PeripheralIdentifier: *peripheralIdentifier,
	}, nil
}

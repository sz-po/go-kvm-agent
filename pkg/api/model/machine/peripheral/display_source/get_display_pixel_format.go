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
	PixelFormatEndpointName = "pixel-format"
)

type GetDisplayPixelFormatRequestPath struct {
	MachineIdentifier    machine.MachineIdentifier
	PeripheralIdentifier peripheral.PeripheralIdentifier
}

func (path GetDisplayPixelFormatRequestPath) String() (*string, error) {
	machineIdentifier, err := path.MachineIdentifier.String()
	if err != nil {
		return nil, err
	}

	peripheralIdentifier, err := path.PeripheralIdentifier.String()
	if err != nil {
		return nil, err
	}

	result := fmt.Sprintf("/%s/%s/%s/%s/%s/%s", machine.EndpointName, *machineIdentifier, peripheral.EndpointName, *peripheralIdentifier, EndpointName, PixelFormatEndpointName)

	return &result, nil
}

func (path GetDisplayPixelFormatRequestPath) Params() (*transport.PathParams, error) {
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

type GetDisplayPixelFormatRequest struct {
	Path GetDisplayPixelFormatRequestPath
}

func (request *GetDisplayPixelFormatRequest) Request() (*transport.Request, error) {
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

func ParseGetDisplayPixelFormatRequest(request transport.Request) (*GetDisplayPixelFormatRequest, error) {
	path, err := ParseGetDisplayPixelFormatRequestPath(request.PathParam)
	if err != nil {
		return nil, fmt.Errorf("parse path: %w", err)
	}

	return &GetDisplayPixelFormatRequest{
		Path: *path,
	}, nil
}

func ParseGetDisplayPixelFormatRequestPath(path transport.PathParams) (*GetDisplayPixelFormatRequestPath, error) {
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

	return &GetDisplayPixelFormatRequestPath{
		MachineIdentifier:    *machineIdentifier,
		PeripheralIdentifier: *peripheralIdentifier,
	}, nil
}

type GetDisplayPixelFormatResponseBody struct {
	PixelFormat peripheralSDK.DisplayPixelFormat `json:"pixelFormat"`
}

func ParseGetDisplayPixelFormatResponseBody(input any) (*GetDisplayPixelFormatResponseBody, error) {
	body, err := transport.UnmarshalResponseBody[GetDisplayPixelFormatResponseBody](input)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return body, nil
}

type GetDisplayPixelFormatResponse struct {
	Body GetDisplayPixelFormatResponseBody
}

func ParseGetDisplayPixelFormatResponse(response transport.Response) (*GetDisplayPixelFormatResponse, error) {
	body, err := ParseGetDisplayPixelFormatResponseBody(response.Body)
	if err != nil {
		return nil, fmt.Errorf("parse body: %w", err)
	}

	return &GetDisplayPixelFormatResponse{
		Body: *body,
	}, nil
}

func (response *GetDisplayPixelFormatResponse) Response() transport.Response {
	return transport.Response{
		StatusCode: http.StatusOK,
		Header: transport.Header{
			transport.HeaderContentType: "application/json",
		},
		Body: response.Body,
	}
}

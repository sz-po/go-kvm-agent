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
	DisplayModeEndpointName = "display-mode"
)

type GetDisplayModeRequestPath struct {
	MachineIdentifier    machine.MachineIdentifier
	PeripheralIdentifier peripheral.PeripheralIdentifier
}

func (path GetDisplayModeRequestPath) String() (*string, error) {
	machineIdentifier, err := path.MachineIdentifier.String()
	if err != nil {
		return nil, err
	}

	peripheralIdentifier, err := path.PeripheralIdentifier.String()
	if err != nil {
		return nil, err
	}

	result := fmt.Sprintf("/%s/%s/%s/%s/%s/%s", machine.EndpointName, *machineIdentifier, peripheral.EndpointName, *peripheralIdentifier, EndpointName, DisplayModeEndpointName)

	return &result, nil
}

func (path GetDisplayModeRequestPath) Params() (*transport.PathParams, error) {
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

type GetDisplayModeRequest struct {
	Path GetDisplayModeRequestPath
}

func (request *GetDisplayModeRequest) Request() (*transport.Request, error) {
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

func ParseGetDisplayModeRequest(request transport.Request) (*GetDisplayModeRequest, error) {
	path, err := ParseGetDisplayModeRequestPath(request.PathParam)
	if err != nil {
		return nil, fmt.Errorf("parse path: %w", err)
	}

	return &GetDisplayModeRequest{
		Path: *path,
	}, nil
}

func ParseGetDisplayModeRequestPath(path transport.PathParams) (*GetDisplayModeRequestPath, error) {
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

	return &GetDisplayModeRequestPath{
		MachineIdentifier:    *machineIdentifier,
		PeripheralIdentifier: *peripheralIdentifier,
	}, nil
}

type GetDisplayModeResponseBody struct {
	DisplayMode peripheralSDK.DisplayMode `json:"displayMode"`
}

func ParseGetDisplayModeResponseBody(input any) (*GetDisplayModeResponseBody, error) {
	body, err := transport.UnmarshalResponseBody[GetDisplayModeResponseBody](input)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return body, nil
}

type GetDisplayModeResponse struct {
	Body GetDisplayModeResponseBody
}

func ParseGetDisplayModeResponse(response transport.Response) (*GetDisplayModeResponse, error) {
	body, err := ParseGetDisplayModeResponseBody(response.Body)
	if err != nil {
		return nil, fmt.Errorf("parse body: %w", err)
	}

	return &GetDisplayModeResponse{
		Body: *body,
	}, nil
}

func (response *GetDisplayModeResponse) Response() transport.Response {
	return transport.Response{
		StatusCode: http.StatusOK,
		Header: transport.Header{
			transport.HeaderContentType: "application/json",
		},
		Body: response.Body,
	}
}

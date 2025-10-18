package peripheral

import (
	"fmt"
	"net/http"

	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
)

type GetRequestPath struct {
	MachineIdentifier    machineAPI.MachineIdentifier
	PeripheralIdentifier PeripheralIdentifier
}

func (path GetRequestPath) String() (*string, error) {
	machineIdentifier, err := path.MachineIdentifier.String()
	if err != nil {
		return nil, err
	}

	peripheralIdentifier, err := path.PeripheralIdentifier.String()
	if err != nil {
		return nil, err
	}

	result := fmt.Sprintf("/%s/%s/%s/%s", machineAPI.EndpointName, *machineIdentifier, EndpointName, *peripheralIdentifier)

	return &result, nil
}

func (path GetRequestPath) Params() (*transport.PathParams, error) {
	machineIdentifier, err := path.MachineIdentifier.String()
	if err != nil {
		return nil, err
	}

	peripheralIdentifier, err := path.PeripheralIdentifier.String()
	if err != nil {
		return nil, err
	}

	return &transport.PathParams{
		machineAPI.MachineIdentifierPathFieldName: *machineIdentifier,
		PeripheralIdentifierPathFieldName:         *peripheralIdentifier,
	}, nil
}

type GetRequest struct {
	Path GetRequestPath
}

func (request *GetRequest) Request() (*transport.Request, error) {
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

func ParseGetRequestPath(path transport.PathParams) (*GetRequestPath, error) {
	err := path.Require(machineAPI.MachineIdentifierPathFieldName, PeripheralIdentifierPathFieldName)
	if err != nil {
		return nil, err
	}

	machineIdentifier, err := machineAPI.ParseMachineIdentifier(path[machineAPI.MachineIdentifierPathFieldName])
	if err != nil {
		return nil, fmt.Errorf("parse machine identifier: %w", err)
	}

	peripheralIdentifier, err := ParsePeripheralIdentifier(path[PeripheralIdentifierPathFieldName])
	if err != nil {
		return nil, fmt.Errorf("parse peripheral identifier: %w", err)
	}

	return &GetRequestPath{
		MachineIdentifier:    *machineIdentifier,
		PeripheralIdentifier: *peripheralIdentifier,
	}, nil
}

func ParseGetRequest(request transport.Request) (*GetRequest, error) {
	path, err := ParseGetRequestPath(request.PathParam)
	if err != nil {
		return nil, fmt.Errorf("parse path: %w", err)
	}

	return &GetRequest{
		Path: *path,
	}, nil
}

type GetResponseBody struct {
	Peripheral Peripheral `json:"peripheral"`
}

func ParseGetResponseBody(input any) (*GetResponseBody, error) {
	body, err := transport.UnmarshalResponseBody[GetResponseBody](input)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return body, nil
}

type GetResponse struct {
	Body GetResponseBody
}

func ParseGetResponse(response transport.Response) (*GetResponse, error) {
	body, err := ParseGetResponseBody(response.Body)
	if err != nil {
		return nil, fmt.Errorf("parse body: %w", err)
	}

	return &GetResponse{
		Body: *body,
	}, nil
}

func (response *GetResponse) Response() transport.Response {
	return transport.Response{
		StatusCode: http.StatusOK,
		Header: transport.Header{
			transport.HeaderContentType: "application/json",
		},
		Body: response.Body,
	}
}

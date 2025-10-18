package peripheral

import (
	"fmt"
	"net/http"

	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
)

type ListRequestPath struct {
	MachineIdentifier machineAPI.MachineIdentifier
}

func (path ListRequestPath) String() (*string, error) {
	machineIdentifier, err := path.MachineIdentifier.String()
	if err != nil {
		return nil, err
	}

	result := fmt.Sprintf("/%s/%s/%s", machineAPI.EndpointName, *machineIdentifier, EndpointName)

	return &result, nil
}

func (path ListRequestPath) Params() (*transport.PathParams, error) {
	machineIdentifier, err := path.MachineIdentifier.String()
	if err != nil {
		return nil, err
	}

	return &transport.PathParams{
		machineAPI.MachineIdentifierPathFieldName: *machineIdentifier,
	}, nil
}

type ListRequest struct {
	Path ListRequestPath
}

func (request *ListRequest) Request() (*transport.Request, error) {
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

type ListResponseBody struct {
	Peripherals []Peripheral `json:"peripherals"`
	TotalCount  int          `json:"totalCount"`
}

type ListResponse struct {
	Body ListResponseBody
}

func ParseListResponseBody(input any) (*ListResponseBody, error) {
	body, err := transport.UnmarshalResponseBody[ListResponseBody](input)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return body, nil
}

func ParseListResponse(response transport.Response) (*ListResponse, error) {
	body, err := ParseListResponseBody(response.Body)
	if err != nil {
		return nil, fmt.Errorf("parse body: %w", err)
	}

	return &ListResponse{
		Body: *body,
	}, nil
}

func ParseListRequestPath(path transport.PathParams) (*ListRequestPath, error) {
	err := path.Require(machineAPI.MachineIdentifierPathFieldName)
	if err != nil {
		return nil, err
	}

	machineIdentifier, err := machineAPI.ParseMachineIdentifier(path[machineAPI.MachineIdentifierPathFieldName])
	if err != nil {
		return nil, fmt.Errorf("parse machine identifier: %w", err)
	}

	return &ListRequestPath{
		MachineIdentifier: *machineIdentifier,
	}, nil
}

func ParseListRequest(request transport.Request) (*ListRequest, error) {
	path, err := ParseListRequestPath(request.PathParam)
	if err != nil {
		return nil, fmt.Errorf("parse path: %w", err)
	}

	return &ListRequest{
		Path: *path,
	}, nil
}

func (response ListResponse) Response() transport.Response {
	return transport.Response{
		StatusCode: http.StatusOK,
		Header: transport.Header{
			transport.HeaderContentType: "application/json",
		},
		Body: response.Body,
	}
}

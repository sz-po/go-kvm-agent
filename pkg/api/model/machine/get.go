package machine

import (
	"fmt"
	"net/http"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
)

type GetRequestPath struct {
	MachineIdentifier MachineIdentifier
}

func (path GetRequestPath) String() (*string, error) {
	machineIdentifier, err := path.MachineIdentifier.String()
	if err != nil {
		return nil, err
	}

	result := fmt.Sprintf("/%s/%s", EndpointName, *machineIdentifier)

	return &result, nil
}

func (path GetRequestPath) Params() (*transport.PathParams, error) {
	machineIdentifier, err := path.MachineIdentifier.String()
	if err != nil {
		return nil, err
	}

	return &transport.PathParams{
		MachineIdentifierPathFieldName: *machineIdentifier,
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

type GetResponseBody struct {
	Machine Machine `json:"machine"`
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

func ParseGetRequestPath(path transport.PathParams) (*GetRequestPath, error) {
	err := path.Require(MachineIdentifierPathFieldName)
	if err != nil {
		return nil, err
	}

	machineIdentifier, err := ParseMachineIdentifier(path[MachineIdentifierPathFieldName])
	if err != nil {
		return nil, fmt.Errorf("parse machine identifier: %w", err)
	}

	return &GetRequestPath{
		MachineIdentifier: *machineIdentifier,
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

func (response *GetResponse) Response() transport.Response {
	return transport.Response{
		StatusCode: http.StatusOK,
		Header: transport.Header{
			transport.HeaderContentType: "application/json",
		},
		Body: response.Body,
	}
}

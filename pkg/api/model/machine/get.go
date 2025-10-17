package machine

import (
	"fmt"
	"net/http"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
)

type GetRequestPath struct {
	MachineIdentifier MachineIdentifier
}

type GetRequest struct {
	Path GetRequestPath
}

type GetResponseBody struct {
	Machine Machine `json:"machine"`
}

type GetResponse struct {
	Body GetResponseBody
}

func ParseGetRequestPath(path transport.Path) (*GetRequestPath, error) {
	err := path.Require("machineIdentifier")
	if err != nil {
		return nil, err
	}

	machineIdentifier, err := ParseMachineIdentifier(path["machineIdentifier"])
	if err != nil {
		return nil, fmt.Errorf("parse machine identifier: %w", err)
	}

	return &GetRequestPath{
		MachineIdentifier: *machineIdentifier,
	}, nil
}

func ParseGetRequest(request transport.Request) (*GetRequest, error) {
	path, err := ParseGetRequestPath(request.Path)
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
			"Content-Type": "application/json",
		},
		Body: response.Body,
	}
}

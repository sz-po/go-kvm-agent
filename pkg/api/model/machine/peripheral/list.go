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

type ListRequest struct {
	Path ListRequestPath
}

type ListResponseBody struct {
	Result     []Peripheral `json:"result"`
	TotalCount int          `json:"totalCount"`
}

type ListResponse struct {
	Body ListResponseBody
}

func ParseListRequestPath(path transport.Path) (*ListRequestPath, error) {
	err := path.Require("machineIdentifier")
	if err != nil {
		return nil, err
	}

	machineIdentifier, err := machineAPI.ParseMachineIdentifier(path["machineIdentifier"])
	if err != nil {
		return nil, fmt.Errorf("parse machine identifier: %w", err)
	}

	return &ListRequestPath{
		MachineIdentifier: *machineIdentifier,
	}, nil
}

func ParseListRequest(request transport.Request) (*ListRequest, error) {
	path, err := ParseListRequestPath(request.Path)
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
			"Content-Type": "application/json",
		},
		Body: response.Body,
	}
}

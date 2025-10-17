package peripheral

import (
	"fmt"
	"net/http"

	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type GetRequestPath struct {
	MachineIdentifier    machineAPI.MachineIdentifier
	PeripheralIdentifier PeripheralIdentifier
}

type GetRequest struct {
	Path GetRequestPath
}

type GetResponseBody struct {
	Id         peripheralSDK.PeripheralId           `json:"id"`
	Name       peripheralSDK.PeripheralName         `json:"name"`
	Capability []peripheralSDK.PeripheralCapability `json:"capability"`
}

type GetResponse struct {
	Body GetResponseBody
}

func ParseGetRequestPath(path transport.Path) (*GetRequestPath, error) {
	err := path.Require("machineIdentifier", "peripheralIdentifier")
	if err != nil {
		return nil, err
	}

	machineIdentifier, err := machineAPI.ParseMachineIdentifier(path["machineIdentifier"])
	if err != nil {
		return nil, fmt.Errorf("parse machine identifier: %w", err)
	}

	peripheralIdentifier, err := ParsePeripheralIdentifier(path["peripheralIdentifier"])
	if err != nil {
		return nil, fmt.Errorf("parse peripheral identifier: %w", err)
	}

	return &GetRequestPath{
		MachineIdentifier:    *machineIdentifier,
		PeripheralIdentifier: *peripheralIdentifier,
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

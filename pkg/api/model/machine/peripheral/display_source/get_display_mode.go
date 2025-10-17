package display_source

import (
	"fmt"
	"net/http"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type GetDisplayModeRequestPath struct {
	MachineIdentifier    machine.MachineIdentifier
	PeripheralIdentifier peripheral.PeripheralIdentifier
}

type GetDisplayModeRequest struct {
	Path GetDisplayModeRequestPath
}

func ParseGetDisplayModeRequest(request transport.Request) (*GetDisplayModeRequest, error) {
	path, err := ParseGetDisplayModeRequestPath(request.Path)
	if err != nil {
		return nil, fmt.Errorf("parse path: %w", err)
	}

	return &GetDisplayModeRequest{
		Path: *path,
	}, nil
}

func ParseGetDisplayModeRequestPath(path transport.Path) (*GetDisplayModeRequestPath, error) {
	err := path.Require("machineIdentifier", "peripheralIdentifier")
	if err != nil {
		return nil, err
	}

	machineIdentifier, err := machine.ParseMachineIdentifier(path["machineIdentifier"])
	if err != nil {
		return nil, fmt.Errorf("parse machine identifier: %w", err)
	}

	peripheralIdentifier, err := peripheral.ParsePeripheralIdentifier(path["peripheralIdentifier"])
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

type GetDisplayModeResponse struct {
	Body GetDisplayModeResponseBody
}

func (response *GetDisplayModeResponse) Response() transport.Response {
	return transport.Response{
		StatusCode: http.StatusOK,
		Header: transport.Header{
			"Content-Type": "application/json",
		},
		Body: response.Body,
	}
}

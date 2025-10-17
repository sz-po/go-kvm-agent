package display_source

import (
	"fmt"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
)

type GetFramebufferRequestPath struct {
	MachineIdentifier    machine.MachineIdentifier
	PeripheralIdentifier peripheral.PeripheralIdentifier
}

type GetFramebufferRequest struct {
	Path GetFramebufferRequestPath
}

type GetFramebufferResponseHeaders struct {
	ContentType string
}

type GetFramebufferResponse struct {
}

func ParseGetFramebufferRequest(request transport.Request) (*GetFramebufferRequest, error) {
	path, err := ParseGetFramebufferRequestPath(request.Path)
	if err != nil {
		return nil, fmt.Errorf("parse path: %w", err)
	}

	return &GetFramebufferRequest{
		Path: *path,
	}, nil
}

func ParseGetFramebufferRequestPath(path transport.Path) (*GetFramebufferRequestPath, error) {
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

	return &GetFramebufferRequestPath{
		MachineIdentifier:    *machineIdentifier,
		PeripheralIdentifier: *peripheralIdentifier,
	}, nil
}

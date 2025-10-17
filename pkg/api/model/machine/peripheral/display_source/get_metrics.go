package display_source

import (
	"fmt"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type GetMetricsRequestPath struct {
	MachineIdentifier    machine.MachineIdentifier
	PeripheralIdentifier peripheral.PeripheralIdentifier
}

type GetMetricsRequest struct {
	Path GetMetricsRequestPath
}

type GetMetricsResponse struct {
	Metrics peripheralSDK.DisplaySourceMetrics `json:"metrics"`
}

func ParseGetMetricsRequest(request transport.Request) (*GetMetricsRequest, error) {
	path, err := ParseGetMetricsRequestPath(request.Path)
	if err != nil {
		return nil, fmt.Errorf("parse path: %w", err)
	}

	return &GetMetricsRequest{
		Path: *path,
	}, nil
}

func ParseGetMetricsRequestPath(path transport.Path) (*GetMetricsRequestPath, error) {
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

	return &GetMetricsRequestPath{
		MachineIdentifier:    *machineIdentifier,
		PeripheralIdentifier: *peripheralIdentifier,
	}, nil
}

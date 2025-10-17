package display

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
)

type ConnectRequestBody struct {
	DisplaySource struct {
		MachineIdentifier    machine.MachineIdentifier       `json:"machineIdentifier"`
		PeripheralIdentifier peripheral.PeripheralIdentifier `json:"peripheralIdentifier"`
	}
	DisplaySink struct {
		MachineIdentifier    machine.MachineIdentifier       `json:"machineIdentifier"`
		PeripheralIdentifier peripheral.PeripheralIdentifier `json:"peripheralIdentifier"`
	}
}

type ConnectRequest struct {
	Body ConnectRequestBody
}

func ParseConnectRequest(request transport.Request) (*ConnectRequest, error) {
	body, err := ParseConnectRequestBody(request.Body)
	if err != nil {
		return nil, fmt.Errorf("parse body: %w", err)
	}

	return &ConnectRequest{
		Body: *body,
	}, nil
}

func ParseConnectRequestBody(reader io.Reader) (*ConnectRequestBody, error) {
	var body ConnectRequestBody

	err := json.NewDecoder(reader).Decode(&body)
	if err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	err = body.DisplaySource.MachineIdentifier.Validate()
	if err != nil {
		return nil, fmt.Errorf("validate display source machine identifier: %w", err)
	}

	err = body.DisplaySource.PeripheralIdentifier.Validate()
	if err != nil {
		return nil, fmt.Errorf("validate display source peripheral identifier: %w", err)
	}

	err = body.DisplaySink.MachineIdentifier.Validate()
	if err != nil {
		return nil, fmt.Errorf("validate display sink machine identifier: %w", err)
	}

	err = body.DisplaySink.PeripheralIdentifier.Validate()
	if err != nil {
		return nil, fmt.Errorf("validate display sink peripheral identifier: %w", err)
	}

	return &body, nil
}

type ConnectResponse struct {
}

func (connectResponse *ConnectResponse) Response() transport.Response {
	return transport.Response{
		StatusCode: http.StatusAccepted,
	}
}

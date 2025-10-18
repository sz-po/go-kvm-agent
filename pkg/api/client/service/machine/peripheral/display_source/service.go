package display_source

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/elnormous/contenttype"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/client/transport"
	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	peripheralAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
	displaySourceAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral/display_source"
	memorySDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/memory"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type Service struct {
	roundTripper transport.RoundTripper

	machineIdentifier    machineAPI.MachineIdentifier
	peripheralIdentifier peripheralAPI.PeripheralIdentifier
}

func NewService(roundTripper transport.RoundTripper, machineIdentifier machineAPI.MachineIdentifier, peripheralIdentifier peripheralAPI.PeripheralIdentifier) (*Service, error) {
	return &Service{
		roundTripper: roundTripper,

		machineIdentifier:    machineIdentifier,
		peripheralIdentifier: peripheralIdentifier,
	}, nil
}

func (service *Service) GetDisplayMode(ctx context.Context) (*peripheralSDK.DisplayMode, error) {
	request := displaySourceAPI.GetDisplayModeRequest{
		Path: displaySourceAPI.GetDisplayModeRequestPath{
			MachineIdentifier:    service.machineIdentifier,
			PeripheralIdentifier: service.peripheralIdentifier,
		},
	}

	response, err := transport.CallUsingRequestProvider[displaySourceAPI.GetDisplayModeResponse](ctx, service.roundTripper, &request, displaySourceAPI.ParseGetDisplayModeResponse)
	if err != nil {
		return nil, fmt.Errorf("call: %w", err)
	}

	return &response.Body.DisplayMode, nil
}

func (service *Service) GetPixelFormat(ctx context.Context) (*peripheralSDK.DisplayPixelFormat, error) {
	request := displaySourceAPI.GetDisplayPixelFormatRequest{
		Path: displaySourceAPI.GetDisplayPixelFormatRequestPath{
			MachineIdentifier:    service.machineIdentifier,
			PeripheralIdentifier: service.peripheralIdentifier,
		},
	}

	response, err := transport.CallUsingRequestProvider[displaySourceAPI.GetDisplayPixelFormatResponse](ctx, service.roundTripper, &request, displaySourceAPI.ParseGetDisplayPixelFormatResponse)
	if err != nil {
		return nil, fmt.Errorf("call: %w", err)
	}

	return &response.Body.PixelFormat, nil
}

func (service *Service) GetFramebuffer(ctx context.Context, memoryPool memorySDK.Pool, bufferSize int, mediaType contenttype.MediaType) (*peripheralSDK.DisplayFrameBuffer, error) {
	request := displaySourceAPI.GetFramebufferRequest{
		Path: displaySourceAPI.GetFramebufferRequestPath{
			MachineIdentifier:    service.machineIdentifier,
			PeripheralIdentifier: service.peripheralIdentifier,
		},
		Headers: displaySourceAPI.GetFramebufferRequestHeaders{
			Accept: mediaType.String(),
		},
	}

	response, err := transport.CallUsingRequestProvider[displaySourceAPI.GetFramebufferResponse](ctx, service.roundTripper, &request, displaySourceAPI.ParseGetFramebufferResponse)
	if err != nil {
		return nil, fmt.Errorf("call: %w", err)
	}

	memoryBuffer, err := memoryPool.Borrow(bufferSize)
	if err != nil {
		return nil, fmt.Errorf("borrow memory buffer: %w", err)
	}

	_, err = response.Body.WriteTo(memoryBuffer)
	if err != nil {
		releaseErr := memoryBuffer.Release()
		if releaseErr != nil {
			slog.Warn("Error while releasing memory buffer.", slog.String("error", releaseErr.Error()))
		}

		return nil, fmt.Errorf("write data to memory buffer: %w", err)
	}

	frameBuffer := peripheralSDK.NewDisplayFrameBuffer(memoryBuffer)

	return frameBuffer, nil
}

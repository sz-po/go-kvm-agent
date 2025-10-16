package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	stdhttp "net/http"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/http"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
	routingSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/routing"
)

type DisplayRouterConnectRequest struct {
	DisplaySource struct {
		MachineName    machineSDK.MachineName       `json:"machineName"`
		PeripheralName peripheralSDK.PeripheralName `json:"peripheralName"`
	}
	DisplaySink struct {
		MachineName    machineSDK.MachineName       `json:"machineName"`
		PeripheralName peripheralSDK.PeripheralName `json:"peripheralName"`
	}
}

type DisplayRouterConnect struct {
	displayRouter     routingSDK.DisplayRouter
	machineRepository machineSDK.Repository
}

func NewDisplayRouterConnect(displayRouter routingSDK.DisplayRouter, machineRepository machineSDK.Repository) *DisplayRouterConnect {
	return &DisplayRouterConnect{
		displayRouter:     displayRouter,
		machineRepository: machineRepository,
	}
}

var _ http.ServerHandler = (*DisplayRouterConnect)(nil)

func (handler *DisplayRouterConnect) Register(router chi.Router) {
	router.Post("/router/display/connect", handler.ServeHTTP)
}

func (handler *DisplayRouterConnect) ServeHTTP(responseWriter stdhttp.ResponseWriter, request *stdhttp.Request) {
	ctx := request.Context()

	var requestBody DisplayRouterConnectRequest
	err := json.NewDecoder(request.Body).Decode(&requestBody)
	if err != nil {
		responseWriter.WriteHeader(stdhttp.StatusBadRequest)
		_, _ = responseWriter.Write([]byte(err.Error()))
		return
	}

	displaySourceMachine, err := handler.machineRepository.GetMachineByName(ctx, requestBody.DisplaySource.MachineName)
	if errors.Is(err, machineSDK.ErrMachineNotFound) {
		responseWriter.WriteHeader(stdhttp.StatusNotFound)
		_, _ = responseWriter.Write([]byte(err.Error()))
		return
	} else if err != nil {
		responseWriter.WriteHeader(500)
		_, _ = responseWriter.Write([]byte(err.Error()))
		return
	}

	displaySinkMachine, err := handler.machineRepository.GetMachineByName(ctx, requestBody.DisplaySink.MachineName)
	if errors.Is(err, machineSDK.ErrMachineNotFound) {
		responseWriter.WriteHeader(stdhttp.StatusNotFound)
		_, _ = responseWriter.Write([]byte(err.Error()))
		return
	} else if err != nil {
		responseWriter.WriteHeader(500)
		_, _ = responseWriter.Write([]byte(err.Error()))
		return
	}

	displaySourcePeripheral, err := displaySourceMachine.Peripherals().GetPeripheralByName(ctx, requestBody.DisplaySource.PeripheralName)
	if errors.Is(err, peripheralSDK.ErrPeripheralNotFound) {
		responseWriter.WriteHeader(stdhttp.StatusNotFound)
		_, _ = responseWriter.Write([]byte(err.Error()))
		return
	} else if err != nil {
		responseWriter.WriteHeader(500)
		_, _ = responseWriter.Write([]byte(err.Error()))
		return
	}

	displaySinkPeripheral, err := displaySinkMachine.Peripherals().GetPeripheralByName(ctx, requestBody.DisplaySink.PeripheralName)
	if errors.Is(err, peripheralSDK.ErrPeripheralNotFound) {
		responseWriter.WriteHeader(stdhttp.StatusNotFound)
		_, _ = responseWriter.Write([]byte(err.Error()))
		return
	} else if err != nil {
		responseWriter.WriteHeader(stdhttp.StatusInternalServerError)
		_, _ = responseWriter.Write([]byte(err.Error()))
		return
	}

	displaySourceId := displaySourcePeripheral.GetId()
	displaySinkId := displaySinkPeripheral.GetId()

	err = handler.handle(ctx, displaySourceId, displaySinkId)

	if err != nil {
		responseWriter.WriteHeader(stdhttp.StatusInternalServerError)
		_, _ = responseWriter.Write([]byte(err.Error()))
		return
	}

	responseWriter.WriteHeader(stdhttp.StatusAccepted)
	return
}

func (handler *DisplayRouterConnect) handle(ctx context.Context, displaySourceId peripheralSDK.PeripheralId, displaySinkId peripheralSDK.PeripheralId) error {
	return handler.displayRouter.Connect(ctx, displaySourceId, displaySinkId)
}

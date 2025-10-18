package display

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/handler/helper"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/transport"
	displayAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/router/display"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	routingSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/routing"
)

func connectHandlerProvider(displayRouter routingSDK.DisplayRouter, machineRepository machineSDK.Repository) func(router chi.Router) {
	return func(router chi.Router) {
		router.Post(fmt.Sprintf("/%s", displayAPI.ConnectEndpointName), connectHandler(displayRouter, machineRepository))
	}
}

func connectHandler(displayRouter routingSDK.DisplayRouter, machineRepository machineSDK.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		request := transport.ParseRequest(r)

		connectRequest, err := displayAPI.ParseConnectRequest(request)
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		displaySourceMachine, err := helper.GetMachineByIdentifier(ctx, machineRepository, connectRequest.Body.DisplaySource.MachineIdentifier)
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		displaySource, err := helper.GetPeripheralByIdentifier(ctx, displaySourceMachine.Peripherals(), connectRequest.Body.DisplaySource.PeripheralIdentifier)
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		displaySinkMachine, err := helper.GetMachineByIdentifier(ctx, machineRepository, connectRequest.Body.DisplaySink.MachineIdentifier)
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		displaySink, err := helper.GetPeripheralByIdentifier(ctx, displaySinkMachine.Peripherals(), connectRequest.Body.DisplaySink.PeripheralIdentifier)
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		err = displayRouter.Connect(ctx, displaySource.GetId(), displaySink.GetId())
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		response := &displayAPI.ConnectResponse{}

		transport.WriteResponse(w, r, response)
	}
}

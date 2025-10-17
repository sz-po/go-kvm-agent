package peripheral

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/handler/helper"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/transport"
	peripheralAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
)

func getHandlerProvider(machineRepository machineSDK.Repository) func(router chi.Router) {
	return func(router chi.Router) {
		router.Get("/{peripheralIdentifier}", getHandler(machineRepository))
	}
}

func getHandler(machineRepository machineSDK.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		request, err := peripheralAPI.ParseGetRequest(transport.ParseRequest(r))
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		machine, err := helper.GetMachineByIdentifier(ctx, machineRepository, request.Path.MachineIdentifier)
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		peripheral, err := helper.GetPeripheralByIdentifier(ctx, machine.Peripherals(), request.Path.PeripheralIdentifier)
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		response := &peripheralAPI.GetResponse{
			Body: peripheralAPI.GetResponseBody{
				Id:         peripheral.GetId(),
				Name:       peripheral.GetName(),
				Capability: peripheral.GetCapabilities(),
			},
		}

		transport.WriteResponse(w, r, response)
	}
}

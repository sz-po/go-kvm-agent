package display_source

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/handler/helper"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/transport"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral/display_source"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

func getDisplayModeHandlerProvider(machineRepository machineSDK.Repository) func(router chi.Router) {
	return func(router chi.Router) {
		router.Get("/display-mode", getDisplayModeHandler(machineRepository))
	}
}

func getDisplayModeHandler(machineRepository machineSDK.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		request, err := display_source.ParseGetDisplayModeRequest(transport.ParseRequest(r))
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

		displaySource, err := peripheralSDK.AsDisplaySource(peripheral)
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		displayMode, err := displaySource.GetDisplayMode(ctx)
		if err != nil {
			transport.HandleError(w, r, err)
		}

		response := &display_source.GetDisplayModeResponse{
			Body: display_source.GetDisplayModeResponseBody{
				DisplayMode: *displayMode,
			},
		}

		transport.WriteResponse(w, r, response)
	}
}

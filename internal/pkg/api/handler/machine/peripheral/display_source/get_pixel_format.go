package display_source

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/handler/helper"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/transport"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral/display_source"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
)

func getPixelFormatHandlerProvider(machineRepository machineSDK.Repository) func(router chi.Router) {
	return func(router chi.Router) {
		router.Get(fmt.Sprintf("/%s", display_source.PixelFormatEndpointName), getPixelFormatHandler(machineRepository))
	}
}

func getPixelFormatHandler(machineRepository machineSDK.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		request, err := display_source.ParseGetDisplayPixelFormatRequest(transport.ParseRequest(r))
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		displaySource, err := helper.GetMachineDisplaySourceByIdentifier(ctx, machineRepository, request.Path.MachineIdentifier, request.Path.PeripheralIdentifier)
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		pixelFormat, err := displaySource.GetDisplayPixelFormat(ctx)

		response := &display_source.GetDisplayPixelFormatResponse{
			Body: display_source.GetDisplayPixelFormatResponseBody{
				PixelFormat: *pixelFormat,
			},
		}

		transport.WriteResponse(w, r, response)
	}
}

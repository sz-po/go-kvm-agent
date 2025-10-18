package peripheral

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/handler/helper"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/transport"
	peripheralAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
)

func listHandlerProvider(machineRepository machineSDK.Repository) func(router chi.Router) {
	return func(router chi.Router) {
		router.Get("/", listHandler(machineRepository))
	}
}

func listHandler(machineRepository machineSDK.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		request, err := peripheralAPI.ParseListRequest(transport.ParseRequest(r))
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		peripheralMachine, err := helper.GetMachineByIdentifier(ctx, machineRepository, request.Path.MachineIdentifier)
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		peripherals, err := peripheralMachine.Peripherals().GetAllPeripherals(ctx)
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		var result []peripheralAPI.Peripheral

		for _, peripheral := range peripherals {
			result = append(result, peripheralAPI.Peripheral{
				Id:           peripheral.GetId(),
				Name:         peripheral.GetName(),
				Capabilities: peripheral.GetCapabilities(),
			})
		}

		response := &peripheralAPI.ListResponse{
			Body: peripheralAPI.ListResponseBody{
				Peripherals: result,
				TotalCount:  len(result),
			},
		}

		transport.WriteResponse(w, r, response)
	}
}

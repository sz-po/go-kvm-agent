package machine

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/handler/helper"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/transport"
	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
)

func getHandlerProvider(machineRepository machineSDK.Repository) func(router chi.Router) {
	return func(router chi.Router) {
		router.Get(fmt.Sprintf("/{%s}", machineAPI.MachineIdentifierPathFieldName), getHandler(machineRepository))
	}
}

func getHandler(machineRepository machineSDK.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		request, err := machineAPI.ParseGetRequest(transport.ParseRequest(r))
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		machine, err := helper.GetMachineByIdentifier(ctx, machineRepository, request.Path.MachineIdentifier)
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		response := &machineAPI.GetResponse{
			Body: machineAPI.GetResponseBody{
				Machine: machineAPI.Machine{
					Id:   machine.GetId(),
					Name: machine.GetName(),
				},
			},
		}

		transport.WriteResponse(w, r, response)
	}
}

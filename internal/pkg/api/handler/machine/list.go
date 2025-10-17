package machine

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/transport"
	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
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

		_, err := machineAPI.ParseListRequest(transport.ParseRequest(r))
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		machines, err := machineRepository.GetAllMachines(ctx)
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		var result []machineAPI.Machine

		for _, machine := range machines {
			apiMachine := machineAPI.Machine{
				Id:   machine.GetId(),
				Name: machine.GetName(),
			}

			result = append(result, apiMachine)
		}

		transport.WriteResponse(w, r, &machineAPI.ListResponse{
			Body: machineAPI.ListResponseBody{
				Result:     result,
				TotalCount: len(result),
			},
		})
	}
}

package machine

import (
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/handler/machine/peripheral"
	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
)

func HandlerProvider(machineRepository machine.Repository) func(router chi.Router) {
	return func(router chi.Router) {
		router.Route(fmt.Sprintf("/%s", machineAPI.EndpointName), func(router chi.Router) {
			router.Route(fmt.Sprintf("/{%s}", machineAPI.MachineIdentifierPathFieldName), peripheral.HandlerProvider(machineRepository))

			listHandlerProvider(machineRepository)(router)
			getHandlerProvider(machineRepository)(router)
		})

	}
}

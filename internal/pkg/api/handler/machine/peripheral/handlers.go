package peripheral

import (
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/handler/machine/peripheral/display_source"
	peripheralAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
)

func HandlerProvider(machineRepository machineSDK.Repository) func(router chi.Router) {
	return func(router chi.Router) {
		router.Route(fmt.Sprintf("/%s", peripheralAPI.EndpointName), func(router chi.Router) {
			router.Route(fmt.Sprintf("/{%s}", peripheralAPI.PeripheralIdentifierPathFieldName), display_source.HandlerProvider(machineRepository))

			listHandlerProvider(machineRepository)(router)
			getHandlerProvider(machineRepository)(router)
		})
	}
}

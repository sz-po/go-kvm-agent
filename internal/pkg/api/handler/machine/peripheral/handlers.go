package peripheral

import (
	"github.com/go-chi/chi/v5"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/handler/machine/peripheral/display_source"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
)

func HandlerProvider(machineRepository machineSDK.Repository) func(router chi.Router) {
	return func(router chi.Router) {
		router.Route("/peripheral", func(router chi.Router) {
			router.Route("/{peripheralIdentifier}", display_source.HandlerProvider(machineRepository))

			listHandlerProvider(machineRepository)(router)
			getHandlerProvider(machineRepository)(router)
		})
	}
}

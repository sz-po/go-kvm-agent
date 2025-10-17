package machine

import (
	"github.com/go-chi/chi/v5"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/handler/machine/peripheral"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
)

func HandlerProvider(machineRepository machine.Repository) func(router chi.Router) {
	return func(router chi.Router) {
		router.Route("/machine", func(router chi.Router) {
			router.Route("/{machineIdentifier}", peripheral.HandlerProvider(machineRepository))

			listHandlerProvider(machineRepository)(router)
			getHandlerProvider(machineRepository)(router)
		})

	}
}

package display_source

import (
	"github.com/go-chi/chi/v5"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
)

func HandlerProvider(machineRepository machineSDK.Repository) func(router chi.Router) {
	return func(router chi.Router) {
		router.Route("/display-source", func(router chi.Router) {
			getDisplayModeHandlerProvider(machineRepository)(router)
		})
	}
}

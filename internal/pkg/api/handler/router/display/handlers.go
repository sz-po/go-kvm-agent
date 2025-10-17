package display

import (
	"github.com/go-chi/chi/v5"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	routingSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/routing"
)

func HandlerProvider(displayRouter routingSDK.DisplayRouter, machineRepository machineSDK.Repository) func(router chi.Router) {
	return func(router chi.Router) {
		router.Route("/display", func(router chi.Router) {
			connectHandlerProvider(displayRouter, machineRepository)(router)
		})
	}
}

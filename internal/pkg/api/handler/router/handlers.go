package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/handler/router/display"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	routingSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/routing"
)

func HandlerProvider(displayRouter routingSDK.DisplayRouter, machineRepository machineSDK.Repository) func(router chi.Router) {
	return func(router chi.Router) {
		router.Route("/router", func(router chi.Router) {
			display.HandlerProvider(displayRouter, machineRepository)(router)
		})
	}
}

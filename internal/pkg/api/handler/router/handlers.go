package router

import (
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/handler/router/display"
	routerAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/router"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	routingSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/routing"
)

func HandlerProvider(displayRouter routingSDK.DisplayRouter, machineRepository machineSDK.Repository) func(router chi.Router) {
	return func(router chi.Router) {
		router.Route(fmt.Sprintf("/%s", routerAPI.EndpointName), func(router chi.Router) {
			display.HandlerProvider(displayRouter, machineRepository)(router)
		})
	}
}

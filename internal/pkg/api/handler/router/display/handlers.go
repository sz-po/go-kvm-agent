package display

import (
	"fmt"

	"github.com/go-chi/chi/v5"
	displayAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/router/display"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	routingSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/routing"
)

func HandlerProvider(displayRouter routingSDK.DisplayRouter, machineRepository machineSDK.Repository) func(router chi.Router) {
	return func(router chi.Router) {
		router.Route(fmt.Sprintf("/%s", displayAPI.EndpointName), func(router chi.Router) {
			connectHandlerProvider(displayRouter, machineRepository)(router)
		})
	}
}

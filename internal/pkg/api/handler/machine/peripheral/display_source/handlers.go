package display_source

import (
	"fmt"

	"github.com/go-chi/chi/v5"
	displaySourceAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral/display_source"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
)

func HandlerProvider(machineRepository machineSDK.Repository) func(router chi.Router) {
	return func(router chi.Router) {
		router.Route(fmt.Sprintf("/%s", displaySourceAPI.EndpointName), func(router chi.Router) {
			getDisplayModeHandlerProvider(machineRepository)(router)
			getFramebufferHandlerProvider(machineRepository)(router)
			getPixelFormatHandlerProvider(machineRepository)(router)
		})
	}
}

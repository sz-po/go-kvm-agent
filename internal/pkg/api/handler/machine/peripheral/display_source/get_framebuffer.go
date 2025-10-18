package display_source

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/elnormous/contenttype"

	"github.com/go-chi/chi/v5"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/handler/helper"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/transport"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral/display_source"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

func getFramebufferHandlerProvider(machineRepository machineSDK.Repository) func(router chi.Router) {
	return func(router chi.Router) {
		router.Get(fmt.Sprintf("/%s", display_source.FramebufferEndpointName), getFramebufferHandler(machineRepository))
	}
}

func getFramebufferHandler(machineRepository machineSDK.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		allowedMediaTypes := []contenttype.MediaType{
			display_source.FramebufferMediaTypeRGB24,
		}

		request, err := display_source.ParseGetFramebufferRequest(transport.ParseRequest(r), allowedMediaTypes)
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		displaySource, err := helper.GetMachineDisplaySourceByIdentifier(ctx, machineRepository, request.Path.MachineIdentifier, request.Path.PeripheralIdentifier)
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		frameBuffer, err := displaySource.GetDisplayFrameBuffer(ctx)
		if err != nil {
			transport.HandleError(w, r, err)
			return
		}

		defer func() {
			if releaseErr := frameBuffer.Release(); releaseErr != nil {
				slog.Warn("Error while releasing framebuffer.", slog.String("error", releaseErr.Error()))
			}
		}()

		var response display_source.GetFramebufferResponse

		switch request.MediaType.String() {
		case display_source.FramebufferMediaTypeRGB24.String():
			response = getFramebufferAsRGB24(frameBuffer)
		default:
			transport.HandleError(w, r, fmt.Errorf("unsupported media type: %s", request.MediaType.String()))
			return
		}

		transport.WriteResponse(w, r, &response)
	}
}

func getFramebufferAsRGB24(frameBuffer *peripheralSDK.DisplayFrameBuffer) display_source.GetFramebufferResponse {
	return display_source.GetFramebufferResponse{
		Headers: display_source.GetFramebufferResponseHeaders{
			ContentType: display_source.FramebufferMediaTypeRGB24,
		},
		Body: frameBuffer,
	}
}

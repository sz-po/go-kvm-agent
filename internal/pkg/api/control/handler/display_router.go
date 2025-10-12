package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	controlApi "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/control"
	routingSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/routing"
)

type DisplayRouterHandler struct {
	router        *chi.Mux
	displayRouter routingSDK.DisplayRouter
}

func NewDisplayRouterHandler(displayRouter routingSDK.DisplayRouter) *DisplayRouterHandler {
	router := chi.NewRouter()

	handler := &DisplayRouterHandler{
		router:        router,
		displayRouter: displayRouter,
	}

	router.Post("/connect", handler.handleConnect)

	return handler
}

func (handler *DisplayRouterHandler) PathPrefix() string {
	return "/router/display"
}

func (handler *DisplayRouterHandler) handleConnect(response http.ResponseWriter, request *http.Request) {
	var requestPayload controlApi.DisplayRouterConnectRequest

	err := json.NewDecoder(request.Body).Decode(&requestPayload)
	if err != nil {
		mustHandleError(response, err)
		return
	}

	err = handler.displayRouter.Connect(requestPayload.DisplaySourceId, requestPayload.DisplaySinkId)
	if err != nil {
		slog.Warn("Failed to connect to display router via API.", slog.String("error", err.Error()))
		mustHandleError(response, err)
		return
	}

	response.WriteHeader(http.StatusNoContent)
}

func (handler *DisplayRouterHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	handler.router.ServeHTTP(writer, request)
}

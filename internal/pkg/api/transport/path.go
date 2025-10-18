package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
)

func parsePathParams(request *http.Request) transport.PathParams {
	path := make(transport.PathParams)

	routeContext := chi.RouteContext(request.Context())
	if routeContext == nil {
		return path
	}

	for index, key := range routeContext.URLParams.Keys {
		if index < len(routeContext.URLParams.Values) {
			path[key] = routeContext.URLParams.Values[index]
		}
	}

	return path
}

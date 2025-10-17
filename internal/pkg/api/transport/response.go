package transport

import (
	"net/http"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
)

type ResponseProvider interface {
	Response() transport.Response
}

func WriteResponse(responseWriter http.ResponseWriter, request *http.Request, responseProvider ResponseProvider) {
	response := responseProvider.Response()

	writeHeader(responseWriter, response.Header)
	responseWriter.WriteHeader(response.StatusCode)
	writeBody(responseWriter, response.Body)
}

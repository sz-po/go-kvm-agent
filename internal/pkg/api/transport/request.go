package transport

import (
	"net/http"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
)

func ParseRequest(request *http.Request) transport.Request {
	return transport.Request{
		Method:    request.Method,
		Path:      request.URL.Path,
		PathParam: parsePathParams(request),
		Query:     parseQuery(request),
		Header:    parseHeaders(request),
		Body:      request.Body,
	}
}

package transport

import (
	"net/http"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
)

func ParseRequest(request *http.Request) transport.Request {
	return transport.Request{
		Path:   parsePath(request),
		Query:  parseQuery(request),
		Header: parseHeader(request),
		Body:   request.Body,
	}
}

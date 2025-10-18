package transport

import "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"

type RequestProvider interface {
	Request() (*transport.Request, error)
}

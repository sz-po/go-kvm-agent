package api

import (
	"errors"

	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
)

type RequestHeader struct {
	MethodName nodeSDK.MethodName `json:"methodName"`
}

type ResponseHeader struct {
	Error error `json:"error,omitempty"`
}

var (
	ErrUnsupportedMethod = errors.New("unsupported method")
	ErrMalformedRequest  = errors.New("malformed request")
)

package transport

import (
	"context"
	"fmt"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
)

type RoundTripper interface {
	Call(ctx context.Context, request transport.Request) (*transport.Response, error)
}

func CallUsingRequestProvider[T any](ctx context.Context, roundTripper RoundTripper, requestProvider RequestProvider, parseFn func(response transport.Response) (*T, error)) (*T, error) {
	request, err := requestProvider.Request()
	if err != nil {
		return nil, fmt.Errorf("request provider: %w", err)
	}

	response, err := roundTripper.Call(ctx, *request)
	if err != nil {
		return nil, err
	}

	parsedResponse, err := parseFn(*response)
	if err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return parsedResponse, nil
}

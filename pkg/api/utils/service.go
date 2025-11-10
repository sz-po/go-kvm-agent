package utils

import (
	"context"
	"fmt"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api"
)

type ServiceRequestHandler[REQ any, RES any] func(context.Context, REQ) (*RES, error)

func HandleServiceRequest[REQ any, RES any](ctx context.Context, codec api.Codec, handler ServiceRequestHandler[REQ, RES]) error {
	var request REQ
	err := codec.Decode(&request)

	if err != nil {
		_ = codec.Encode(&api.ResponseHeader{Error: api.ErrMalformedRequest})
		return api.ErrMalformedRequest
	}

	response, err := handler(ctx, request)
	if err != nil {
		_ = codec.Encode(&api.ResponseHeader{Error: err})
		return err
	}

	err = codec.Encode(&api.ResponseHeader{})
	if err != nil {
		return fmt.Errorf("encoding response heeader: %w", err)
	}

	err = codec.Encode(response)
	if err != nil {
		return fmt.Errorf("encoding response: %w", err)
	}

	return nil
}

package utils

import (
	"context"
	"errors"
	"fmt"

	apiSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/api"
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
)

func HandleClientRequest[REQ any, RES any](ctx context.Context, codec apiSDK.Codec, methodName nodeSDK.MethodName, request REQ) (*RES, error) {
	requestHeader := &apiSDK.RequestHeader{
		MethodName: methodName,
	}
	err := codec.Encode(requestHeader)
	if err != nil {
		return nil, fmt.Errorf("encode request header: %w", err)
	}

	err = codec.Encode(request)
	if err != nil {
		return nil, fmt.Errorf("decode request: %w", err)
	}

	responseHeader := &apiSDK.ResponseHeader{}
	err = codec.Decode(responseHeader)
	if err != nil {
		return nil, fmt.Errorf("decode response header: %w", err)
	}

	if len(responseHeader.Error) > 0 {
		return nil, errors.New(responseHeader.Error)
	}

	var response RES
	err = codec.Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &response, nil
}

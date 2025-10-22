package display_source

import (
	"fmt"
	"io"
	"net/http"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
)

const (
	FramebufferEndpointName = "framebuffer"
)

type GetFramebufferRequestPath struct {
	MachineIdentifier    machine.MachineIdentifier
	PeripheralIdentifier peripheral.PeripheralIdentifier
}

func (path GetFramebufferRequestPath) String() (*string, error) {
	machineIdentifier, err := path.MachineIdentifier.String()
	if err != nil {
		return nil, err
	}

	peripheralIdentifier, err := path.PeripheralIdentifier.String()
	if err != nil {
		return nil, err
	}

	result := fmt.Sprintf("/%s/%s/%s/%s/%s/%s", machine.EndpointName, *machineIdentifier, peripheral.EndpointName, *peripheralIdentifier, EndpointName, FramebufferEndpointName)

	return &result, nil
}

func (path GetFramebufferRequestPath) Params() (*transport.PathParams, error) {
	machineIdentifier, err := path.MachineIdentifier.String()
	if err != nil {
		return nil, err
	}

	peripheralIdentifier, err := path.PeripheralIdentifier.String()
	if err != nil {
		return nil, err
	}

	return &transport.PathParams{
		machine.MachineIdentifierPathFieldName:       *machineIdentifier,
		peripheral.PeripheralIdentifierPathFieldName: *peripheralIdentifier,
	}, nil
}

type GetFramebufferRequestHeaders struct {
	Accept string
}

type GetFramebufferRequest struct {
	Path      GetFramebufferRequestPath
	Headers   GetFramebufferRequestHeaders
	MediaType transport.MediaType
}

func (request *GetFramebufferRequest) Request() (*transport.Request, error) {
	path, err := request.Path.String()
	if err != nil {
		return nil, fmt.Errorf("path: %w", err)
	}

	pathParams, err := request.Path.Params()
	if err != nil {
		return nil, fmt.Errorf("path params: %w", err)
	}

	return &transport.Request{
		Method:    http.MethodGet,
		Path:      *path,
		PathParam: *pathParams,
		Header: transport.Header{
			transport.HeaderAccept: request.Headers.Accept,
		},
	}, nil
}

func ParseGetFramebufferRequest(request transport.Request, acceptableMediaTypes []transport.MediaType) (*GetFramebufferRequest, error) {
	path, err := ParseGetFramebufferRequestPath(request.PathParam)
	if err != nil {
		return nil, fmt.Errorf("parse path: %w", err)
	}

	headers, err := ParseGetFramebufferRequestHeaders(request.Header)
	if err != nil {
		return nil, fmt.Errorf("parse headers: %w", err)
	}

	acceptedMediaType, err := transport.NegotiateAcceptedMediaType(headers.Accept, acceptableMediaTypes)
	if err != nil {
		return nil, fmt.Errorf("get acceptable media type: %w", err)
	}

	return &GetFramebufferRequest{
		Path:      *path,
		Headers:   *headers,
		MediaType: acceptedMediaType,
	}, nil
}

func ParseGetFramebufferRequestHeaders(header transport.Header) (*GetFramebufferRequestHeaders, error) {
	err := header.Require(transport.HeaderAccept)
	if err != nil {
		return nil, err
	}

	return &GetFramebufferRequestHeaders{
		Accept: header.Get(transport.HeaderAccept),
	}, nil
}

func ParseGetFramebufferRequestPath(path transport.PathParams) (*GetFramebufferRequestPath, error) {
	err := path.Require(machine.MachineIdentifierPathFieldName, peripheral.PeripheralIdentifierPathFieldName)
	if err != nil {
		return nil, err
	}

	machineIdentifier, err := machine.ParseMachineIdentifier(path[machine.MachineIdentifierPathFieldName])
	if err != nil {
		return nil, fmt.Errorf("parse machine identifier: %w", err)
	}

	peripheralIdentifier, err := peripheral.ParsePeripheralIdentifier(path[peripheral.PeripheralIdentifierPathFieldName])
	if err != nil {
		return nil, fmt.Errorf("parse peripheral identifier: %w", err)
	}

	return &GetFramebufferRequestPath{
		MachineIdentifier:    *machineIdentifier,
		PeripheralIdentifier: *peripheralIdentifier,
	}, nil
}

type GetFramebufferResponseHeaders struct {
	ContentType transport.MediaType
}

func ParseGetFramebufferResponseHeaders(headers transport.Header) (*GetFramebufferResponseHeaders, error) {
	err := headers.Require(transport.HeaderContentType)
	if err != nil {
		return nil, err
	}

	contentType, err := transport.NewMediaType(headers.Get(transport.HeaderContentType))
	if err != nil {
		return nil, fmt.Errorf("parse content type: %w", err)
	}

	return &GetFramebufferResponseHeaders{
		ContentType: *contentType,
	}, nil
}

type GetFramebufferResponse struct {
	Body    io.WriterTo
	Headers GetFramebufferResponseHeaders
}

func ParseGetFramebufferResponse(response transport.Response) (*GetFramebufferResponse, error) {
	headers, err := ParseGetFramebufferResponseHeaders(response.Header)
	if err != nil {
		return nil, fmt.Errorf("parse headers: %w", err)
	}

	switch body := response.Body.(type) {
	case io.WriterTo:
		return &GetFramebufferResponse{
			Body:    body,
			Headers: *headers,
		}, nil
	case io.ReadCloser:
		return &GetFramebufferResponse{
			Body:    transport.ResponseWriterTo{ReadCloser: body},
			Headers: *headers,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported response body type: %T", response.Body)
	}

}

func (response *GetFramebufferResponse) Response() transport.Response {
	return transport.Response{
		StatusCode: http.StatusOK,
		Header: transport.Header{
			transport.HeaderContentType: response.Headers.ContentType.String(),
		},
		Body: response.Body,
	}
}

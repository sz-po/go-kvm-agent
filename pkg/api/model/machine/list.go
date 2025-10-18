package machine

import (
	"fmt"
	"net/http"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
)

type ListRequest struct {
}

func ParseListRequest(request transport.Request) (*ListRequest, error) {
	return &ListRequest{}, nil
}

func (request *ListRequest) Request() (*transport.Request, error) {
	return &transport.Request{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/%s", EndpointName),
	}, nil
}

type ListResponseBody struct {
	Machines   []Machine `json:"machines"`
	TotalCount int       `json:"totalCount"`
}

func ParseListResponseBody(input any) (*ListResponseBody, error) {
	body, err := transport.UnmarshalResponseBody[ListResponseBody](input)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return body, nil
}

type ListResponse struct {
	Body ListResponseBody
}

func ParseListResponse(response transport.Response) (*ListResponse, error) {
	body, err := ParseListResponseBody(response.Body)
	if err != nil {
		return nil, fmt.Errorf("parse body: %w", err)
	}

	return &ListResponse{
		Body: *body,
	}, nil
}

func (response ListResponse) Response() transport.Response {
	return transport.Response{
		StatusCode: http.StatusOK,
		Header: transport.Header{
			transport.HeaderContentType: "application/json",
		},
		Body: response.Body,
	}
}

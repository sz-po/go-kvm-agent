package machine

import (
	"net/http"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
)

type ListRequest struct {
}

func ParseListRequest(request transport.Request) (*ListRequest, error) {
	return &ListRequest{}, nil
}

type ListResponseBody struct {
	Result     []Machine `json:"result"`
	TotalCount int       `json:"totalCount"`
}

type ListResponse struct {
	Body ListResponseBody
}

func (response ListResponse) Response() transport.Response {
	return transport.Response{
		StatusCode: http.StatusOK,
		Header: transport.Header{
			"Content-Type": "application/json",
		},
		Body: response.Body,
	}
}

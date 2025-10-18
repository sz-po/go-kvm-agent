package transport

import (
	"fmt"
	"io"
)

type Request struct {
	Method    string
	Path      string
	PathParam PathParams
	Query     map[string]string
	Header    map[string]string
	Body      io.Reader
}

func (request Request) Request() (*Request, error) {
	return &request, nil
}

func (request Request) WithPathPrefix(prefix string) *Request {
	return &Request{
		Method:    request.Method,
		Path:      fmt.Sprintf("%s%s", prefix, request.Path),
		PathParam: request.PathParam.Clone(),
		Query:     request.Query,
		Header:    request.Header,
		Body:      request.Body,
	}
}

package transport

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
)

func TestHTTPRoundTripperCall(t *testing.T) {
	t.Parallel()

	requestPayload := "example request payload"

	httpServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		assert.Equal(t, http.MethodPost, httpRequest.Method)
		assert.Equal(t, "/example/path", httpRequest.URL.Path)
		assert.Equal(t, "value", httpRequest.URL.Query().Get("queryKey"))
		assert.Equal(t, "headerValue", httpRequest.Header.Get("Custom-Header"))

		receivedBody, err := io.ReadAll(httpRequest.Body)
		assert.NoError(t, err)
		assert.Equal(t, requestPayload, string(receivedBody))

		closeErr := httpRequest.Body.Close()
		assert.NoError(t, closeErr)

		responseWriter.Header().Set("Content-Type", "application/json")
		responseWriter.Header().Set("X-Custom", "custom-value")
		responseWriter.WriteHeader(http.StatusCreated)

		_, writeErr := responseWriter.Write([]byte("{\"status\":\"ok\"}"))
		assert.NoError(t, writeErr)
	}))
	t.Cleanup(httpServer.Close)

	httpServerURL, err := url.Parse(httpServer.URL)
	assert.NoError(t, err)

	host, portString, err := net.SplitHostPort(httpServerURL.Host)
	assert.NoError(t, err)

	port, err := strconv.Atoi(portString)
	assert.NoError(t, err)

	roundTripper, err := NewHTTPRoundTripper(httpServerURL.Scheme, host, port)
	assert.NoError(t, err)

	transportResponse, err := roundTripper.Call(
		context.Background(),
		transport.Request{
			Method: http.MethodPost,
			Path:   "/example/path",
			Query: map[string]string{
				"queryKey": "value",
			},
			Header: map[string]string{
				"Custom-Header": "headerValue",
			},
			Body: strings.NewReader(requestPayload),
		},
	)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusCreated, transportResponse.StatusCode)
	assert.Equal(t, "application/json", transportResponse.Header["content-type"])
	assert.Equal(t, "custom-value", transportResponse.Header["x-custom"])

	responseBody, ok := transportResponse.Body.([]byte)
	assert.True(t, ok)
	assert.Equal(t, "{\"status\":\"ok\"}", string(responseBody))
}

func TestNewHTTPRoundTripperUnsupportedScheme(t *testing.T) {
	t.Parallel()

	roundTripper, err := NewHTTPRoundTripper("ftp", "localhost", 8080)

	assert.Nil(t, roundTripper)
	assert.Error(t, err)
}

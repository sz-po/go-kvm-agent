package transport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
)

type HTTPRoundTripperOption func(roundTripper *HTTPRoundTripper)

type HTTPRoundTripper struct {
	httpClient *http.Client
	baseURL    *url.URL
	logger     *slog.Logger
}

func NewHTTPRoundTripper(scheme string, remoteAddress string, remotePort int, options ...HTTPRoundTripperOption) (*HTTPRoundTripper, error) {
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme: %s", scheme)
	}

	host := net.JoinHostPort(remoteAddress, strconv.Itoa(remotePort))

	baseURL := &url.URL{
		Scheme: scheme,
		Host:   host,
	}

	httpClient := &http.Client{
		Timeout: time.Second * 5,
	}

	roundTripper := &HTTPRoundTripper{
		httpClient: httpClient,
		baseURL:    baseURL,
	}

	for _, option := range options {
		option(roundTripper)
	}

	return roundTripper, nil
}

func WithLogger(logger *slog.Logger) HTTPRoundTripperOption {
	return func(roundTripper *HTTPRoundTripper) {
		roundTripper.logger = logger
	}
}

func (roundTripper *HTTPRoundTripper) Call(ctx context.Context, request transport.Request) (*transport.Response, error) {
	startTime := time.Now()

	httpRequest, err := roundTripper.buildHTTPRequest(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	httpResponse, err := roundTripper.httpClient.Do(httpRequest)
	if err != nil {
		wrappedErr := fmt.Errorf("perform request: %w", err)
		roundTripper.logRequestCompleted(httpRequest, nil, time.Since(startTime), wrappedErr)
		return nil, wrappedErr
	}

	response, err := roundTripper.buildTransportResponse(httpResponse)
	if err != nil {
		wrappedErr := fmt.Errorf("build response: %w", err)
		roundTripper.logRequestCompleted(httpRequest, httpResponse, time.Since(startTime), wrappedErr)
		return nil, wrappedErr
	}

	roundTripper.logRequestCompleted(httpRequest, httpResponse, time.Since(startTime), nil)

	return response, nil
}

func (roundTripper *HTTPRoundTripper) buildHTTPRequest(ctx context.Context, request transport.Request) (*http.Request, error) {
	requestURL, err := roundTripper.composeURL(request.Path, request.Query)
	if err != nil {
		return nil, fmt.Errorf("compose url: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, request.Method, requestURL.String(), request.Body)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	for headerKey, headerValue := range request.Header {
		httpRequest.Header.Set(headerKey, headerValue)
	}

	return httpRequest, nil
}

func (roundTripper *HTTPRoundTripper) composeURL(path string, query map[string]string) (*url.URL, error) {
	requestURL := *roundTripper.baseURL
	requestURL.Path = path

	if len(query) > 0 {
		urlValues := requestURL.Query()
		for queryKey, queryValue := range query {
			urlValues.Set(queryKey, queryValue)
		}
		requestURL.RawQuery = urlValues.Encode()
	}

	return &requestURL, nil
}

func (roundTripper *HTTPRoundTripper) buildTransportResponse(httpResponse *http.Response) (*transport.Response, error) {
	header := make(transport.Header, len(httpResponse.Header))
	for headerKey, headerValues := range httpResponse.Header {
		if len(headerValues) == 0 {
			continue
		}
		header[strings.ToLower(headerKey)] = strings.Join(headerValues, ",")
	}

	var body any

	switch header.Get(transport.HeaderContentType) {
	case transport.ApplicationJsonMediaType.String():
		if bodyBuffer, err := io.ReadAll(httpResponse.Body); err != nil {
			return nil, fmt.Errorf("read response body: %w", err)
		} else {
			body = bytes.NewBuffer(bodyBuffer)
		}
		_ = httpResponse.Body.Close()
	default:
		body = httpResponse.Body
	}
	
	return &transport.Response{
		StatusCode: httpResponse.StatusCode,
		Header:     header,
		Body:       body,
	}, nil
}

func (roundTripper *HTTPRoundTripper) logRequestCompleted(httpRequest *http.Request, httpResponse *http.Response, duration time.Duration, resultErr error) {
	if roundTripper.logger == nil {
		return
	}

	attributes := []slog.Attr{
		slog.String("method", httpRequest.Method),
		slog.String("url", httpRequest.URL.String()),
		slog.Duration("duration", duration),
	}

	if httpResponse != nil {
		attributes = append(attributes, slog.Int("statusCode", httpResponse.StatusCode))
	}

	arguments := make([]any, 0, len(attributes)+1)
	for _, attribute := range attributes {
		arguments = append(arguments, attribute)
	}

	if resultErr != nil {
		arguments = append(arguments, slog.String("error", resultErr.Error()))
		roundTripper.logger.Error("HTTP request failed.", arguments...)
		return
	}

	roundTripper.logger.Info("HTTP request completed.", arguments...)
}

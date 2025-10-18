package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	transportSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
)

func TestParseHeader(t *testing.T) {
	t.Run("returns empty header when no headers", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		request.Header = http.Header{}

		result := parseHeaders(request)

		assert.Empty(t, result)
	})

	t.Run("parses headers with lowercase keys", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer token")

		result := parseHeaders(request)

		assert.Equal(t, "application/json", result[transportSDK.HeaderContentType])
		assert.Equal(t, "Bearer token", result["authorization"])
	})

	t.Run("takes first value from multi-value headers", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		request.Header.Add("Accept", "text/html")
		request.Header.Add("Accept", "application/json")

		result := parseHeaders(request)

		assert.Equal(t, "text/html", result[transportSDK.HeaderAccept])
	})
}

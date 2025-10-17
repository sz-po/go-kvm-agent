package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHeader(t *testing.T) {
	t.Run("returns empty header when no headers", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		request.Header = http.Header{}

		result := parseHeader(request)

		assert.Empty(t, result)
	})

	t.Run("parses headers", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer token")

		result := parseHeader(request)

		assert.Equal(t, "application/json", result["Content-Type"])
		assert.Equal(t, "Bearer token", result["Authorization"])
	})

	t.Run("takes first value from multi-value headers", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		request.Header.Add("Accept", "text/html")
		request.Header.Add("Accept", "application/json")

		result := parseHeader(request)

		assert.Equal(t, "text/html", result["Accept"])
	})
}

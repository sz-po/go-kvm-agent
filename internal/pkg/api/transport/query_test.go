package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseQuery(t *testing.T) {
	t.Run("returns empty query when no query parameters", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/test", nil)

		result := parseQuery(request)

		assert.Empty(t, result)
	})

	t.Run("parses query parameters", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/test?key1=value1&key2=value2", nil)

		result := parseQuery(request)

		assert.Equal(t, "value1", result["key1"])
		assert.Equal(t, "value2", result["key2"])
	})

	t.Run("takes first value from multi-value parameters", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/test?key=first&key=second", nil)

		result := parseQuery(request)

		assert.Equal(t, "first", result["key"])
	})
}

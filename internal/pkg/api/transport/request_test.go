package transport

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestParseRequest(t *testing.T) {
	t.Run("parses full request with path, query, headers and body", func(t *testing.T) {
		router := chi.NewRouter()
		var capturedRequest *http.Request

		router.Get("/machine/{machineId}/peripheral/{peripheralId}", func(w http.ResponseWriter, r *http.Request) {
			capturedRequest = r
		})

		body := strings.NewReader(`{"key":"value"}`)
		request := httptest.NewRequest(http.MethodGet, "/machine/test-machine/peripheral/test-peripheral?query1=value1&query2=value2", body)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer token")

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		result := ParseRequest(capturedRequest)

		assert.Equal(t, "test-machine", result.Path["machineId"])
		assert.Equal(t, "test-peripheral", result.Path["peripheralId"])
		assert.Equal(t, "value1", result.Query["query1"])
		assert.Equal(t, "value2", result.Query["query2"])
		assert.Equal(t, "application/json", result.Header["Content-Type"])
		assert.Equal(t, "Bearer token", result.Header["Authorization"])
		assert.NotNil(t, result.Body)

		bodyContent, err := io.ReadAll(result.Body)
		assert.NoError(t, err)
		assert.Equal(t, `{"key":"value"}`, string(bodyContent))
	})

	t.Run("parses request with empty path and query", func(t *testing.T) {
		router := chi.NewRouter()
		var capturedRequest *http.Request

		router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			capturedRequest = r
		})

		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		result := ParseRequest(capturedRequest)

		assert.Empty(t, result.Path)
		assert.Empty(t, result.Query)
		assert.NotNil(t, result.Body)
	})

	t.Run("parses request without chi route context", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/test?query=value", nil)
		request.Header.Set("Custom-Header", "custom-value")

		result := ParseRequest(request)

		assert.Empty(t, result.Path)
		assert.Equal(t, "value", result.Query["query"])
		assert.Equal(t, "custom-value", result.Header["Custom-Header"])
		assert.NotNil(t, result.Body)
	})

	t.Run("takes first value from multi-value query parameters", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/test?multi=first&multi=second&multi=third", nil)

		result := ParseRequest(request)

		assert.Equal(t, "first", result.Query["multi"])
	})

	t.Run("takes first value from multi-value headers", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		request.Header.Add("Accept", "text/html")
		request.Header.Add("Accept", "application/json")
		request.Header.Add("Accept", "application/xml")

		result := ParseRequest(request)

		assert.Equal(t, "text/html", result.Header["Accept"])
	})

	t.Run("handles chi route context with multiple path parameters", func(t *testing.T) {
		router := chi.NewRouter()
		var capturedRequest *http.Request

		router.Get("/api/{version}/machine/{machineId}/peripheral/{peripheralId}/action/{actionId}", func(w http.ResponseWriter, r *http.Request) {
			capturedRequest = r
		})

		request := httptest.NewRequest(http.MethodGet, "/api/v1/machine/machine-123/peripheral/peripheral-456/action/start", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		result := ParseRequest(capturedRequest)

		assert.Equal(t, "v1", result.Path["version"])
		assert.Equal(t, "machine-123", result.Path["machineId"])
		assert.Equal(t, "peripheral-456", result.Path["peripheralId"])
		assert.Equal(t, "start", result.Path["actionId"])
	})
}

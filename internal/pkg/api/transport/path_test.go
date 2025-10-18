package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestParsePath(t *testing.T) {
	t.Run("returns empty path when no route context", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/test", nil)

		result := parsePathParams(request)

		assert.Empty(t, result)
	})

	t.Run("parses path parameters from chi route context", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/test", nil)

		routeContext := chi.NewRouteContext()
		routeContext.URLParams.Add("key1", "value1")
		routeContext.URLParams.Add("key2", "value2")

		ctx := context.WithValue(request.Context(), chi.RouteCtxKey, routeContext)
		request = request.WithContext(ctx)

		result := parsePathParams(request)

		assert.Equal(t, "value1", result["key1"])
		assert.Equal(t, "value2", result["key2"])
	})
}

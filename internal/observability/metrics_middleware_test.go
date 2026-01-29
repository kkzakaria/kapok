package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsMiddleware(t *testing.T) {
	reg := prometheus.NewRegistry()
	mc := NewMetricsCollector(reg)
	logger := zerolog.Nop()

	mw := NewMetricsMiddleware(mc, logger)

	handler := mw.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify metrics were recorded
	families, err := reg.Gather()
	require.NoError(t, err)

	found := false
	for _, f := range families {
		if f.GetName() == "kapok_http_requests_total" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected kapok_http_requests_total metric")
}

func TestMetricsMiddleware_CapturesStatusCode(t *testing.T) {
	reg := prometheus.NewRegistry()
	mc := NewMetricsCollector(reg)
	logger := zerolog.Nop()

	mw := NewMetricsMiddleware(mc, logger)

	handler := mw.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodPost, "/missing", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/api/test", "/api/test"},
		{"/api/tenants/123/resources", "/api/tenants/:id/resources"},
		{"/api/tenants/550e8400-e29b-41d4-a716-446655440000", "/api/tenants/:id"},
		{"/api/users/abcdef0123456789", "/api/users/:id"},
		{"/healthz", "/healthz"},
		{"/graphql", "/graphql"},
		{"/api/short", "/api/short"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizePath(tt.input))
		})
	}
}

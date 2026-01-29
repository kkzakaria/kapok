package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_WithoutTracing(t *testing.T) {
	cfg := Config{
		Enabled:        true,
		MetricsPort:    9090,
		TracingEnabled: false,
	}

	obs, err := New(context.Background(), cfg, zerolog.Nop())
	require.NoError(t, err)
	require.NotNil(t, obs)

	assert.NotNil(t, obs.Metrics)
	assert.NotNil(t, obs.Health)
	assert.Nil(t, obs.Tracing)
}

func TestObservability_MetricsHandler(t *testing.T) {
	cfg := Config{Enabled: true, TracingEnabled: false}
	obs, err := New(context.Background(), cfg, zerolog.Nop())
	require.NoError(t, err)

	handler := obs.MetricsHandler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "go_goroutines")
}

func TestObservability_HTTPMiddleware(t *testing.T) {
	cfg := Config{Enabled: true, TracingEnabled: false}
	obs, err := New(context.Background(), cfg, zerolog.Nop())
	require.NoError(t, err)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := obs.HTTPMiddleware(inner)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestObservability_Shutdown(t *testing.T) {
	cfg := Config{Enabled: true, TracingEnabled: false}
	obs, err := New(context.Background(), cfg, zerolog.Nop())
	require.NoError(t, err)

	err = obs.Shutdown(context.Background())
	assert.NoError(t, err)
}

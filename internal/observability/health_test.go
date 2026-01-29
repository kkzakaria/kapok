package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthChecker_Liveness(t *testing.T) {
	hc := NewHealthChecker(zerolog.Nop())
	handler := hc.LivenessHandler()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp["status"])
}

func TestHealthChecker_ReadinessAllHealthy(t *testing.T) {
	hc := NewHealthChecker(zerolog.Nop())
	hc.Register("db", func(ctx context.Context) error { return nil })

	handler := hc.ReadinessHandler()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHealthChecker_ReadinessUnhealthy(t *testing.T) {
	hc := NewHealthChecker(zerolog.Nop())
	hc.Register("db", func(ctx context.Context) error { return fmt.Errorf("connection refused") })

	handler := hc.ReadinessHandler()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
}

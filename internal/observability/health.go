package observability

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// CheckFunc is a function that performs a health check.
type CheckFunc func(ctx context.Context) error

// HealthChecker manages health check endpoints.
type HealthChecker struct {
	mu     sync.RWMutex
	checks map[string]CheckFunc
	logger zerolog.Logger
}

// NewHealthChecker creates a new HealthChecker.
func NewHealthChecker(logger zerolog.Logger) *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]CheckFunc),
		logger: logger,
	}
}

// Register adds a named health check.
func (hc *HealthChecker) Register(name string, check CheckFunc) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checks[name] = check
}

// LivenessHandler returns an HTTP handler for /healthz (always 200).
func (hc *HealthChecker) LivenessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
}

// ReadinessHandler returns an HTTP handler for /readyz.
func (hc *HealthChecker) ReadinessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		hc.mu.RLock()
		checks := make(map[string]CheckFunc, len(hc.checks))
		for k, v := range hc.checks {
			checks[k] = v
		}
		hc.mu.RUnlock()

		results := make(map[string]string)
		allHealthy := true
		for name, check := range checks {
			if err := check(ctx); err != nil {
				results[name] = err.Error()
				allHealthy = false
				hc.logger.Warn().Str("check", name).Err(err).Msg("health check failed")
			} else {
				results[name] = "ok"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if allHealthy {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": map[bool]string{true: "ok", false: "unavailable"}[allHealthy],
			"checks": results,
		})
	})
}

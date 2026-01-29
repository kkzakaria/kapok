package observability

import (
	"fmt"
	"net/http"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/kapok/kapok/internal/tenant"
	"github.com/rs/zerolog"
)

// MetricsMiddleware records HTTP request metrics.
type MetricsMiddleware struct {
	metrics *MetricsCollector
	logger  zerolog.Logger
}

// NewMetricsMiddleware creates a new MetricsMiddleware.
func NewMetricsMiddleware(metrics *MetricsCollector, logger zerolog.Logger) *MetricsMiddleware {
	return &MetricsMiddleware{metrics: metrics, logger: logger}
}

// Middleware returns an HTTP middleware that records request metrics.
// Uses httpsnoop to preserve http.Hijacker/http.Flusher interfaces.
func (m *MetricsMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		var statusCode int

		metrics := httpsnoop.CaptureMetrics(next, w, r)
		statusCode = metrics.Code

		duration := time.Since(start).Seconds()
		tenantID, _ := tenant.GetTenantID(r.Context())
		if tenantID == "" {
			tenantID = "unknown"
		}
		status := fmt.Sprintf("%d", statusCode)
		path := r.URL.Path

		m.metrics.HTTPRequestsTotal.WithLabelValues(tenantID, r.Method, path, status).Inc()
		m.metrics.HTTPRequestDuration.WithLabelValues(tenantID, r.Method, path).Observe(duration)
	})
}

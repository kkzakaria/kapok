package observability

import (
	"fmt"
	"net/http"
	"strings"
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

		metrics := httpsnoop.CaptureMetrics(next, w, r)

		duration := time.Since(start).Seconds()
		tenantID, _ := tenant.GetTenantID(r.Context())
		if tenantID == "" {
			tenantID = "unknown"
		}
		status := fmt.Sprintf("%d", metrics.Code)
		path := normalizePath(r.URL.Path)

		m.metrics.HTTPRequestsTotal.WithLabelValues(tenantID, r.Method, path, status).Inc()
		m.metrics.HTTPRequestDuration.WithLabelValues(tenantID, r.Method, path).Observe(duration)
	})
}

// normalizePath replaces dynamic path segments (UUIDs, numeric IDs) with
// placeholders to prevent high-cardinality label values in Prometheus.
func normalizePath(path string) string {
	segments := strings.Split(path, "/")
	for i, seg := range segments {
		if seg == "" {
			continue
		}
		if isDynamicSegment(seg) {
			segments[i] = ":id"
		}
	}
	return strings.Join(segments, "/")
}

// isDynamicSegment returns true if the segment looks like a dynamic ID
// (UUID, numeric, or hex string longer than 8 chars).
func isDynamicSegment(s string) bool {
	// Numeric IDs
	allDigits := true
	for _, c := range s {
		if c < '0' || c > '9' {
			allDigits = false
			break
		}
	}
	if allDigits && len(s) > 0 {
		return true
	}

	// UUIDs (with or without hyphens)
	if len(s) == 36 && s[8] == '-' && s[13] == '-' && s[18] == '-' && s[23] == '-' {
		return true
	}

	// Hex strings (e.g. short IDs) longer than 8 chars
	if len(s) > 8 {
		allHex := true
		for _, c := range s {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				allHex = false
				break
			}
		}
		if allHex {
			return true
		}
	}

	return false
}

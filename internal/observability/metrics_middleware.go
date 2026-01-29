package observability

import (
	"fmt"
	"net/http"
	"time"

	"github.com/kapok/kapok/internal/tenant"
	"github.com/rs/zerolog"
)

// responseWriter wraps http.ResponseWriter to capture the status code and bytes written.
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

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
func (m *MetricsMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		tenantID, _ := tenant.GetTenantID(r.Context())
		if tenantID == "" {
			tenantID = "unknown"
		}
		status := fmt.Sprintf("%d", rw.statusCode)

		m.metrics.HTTPRequestsTotal.WithLabelValues(tenantID, r.Method, r.URL.Path, status).Inc()
		m.metrics.HTTPRequestDuration.WithLabelValues(tenantID, r.Method, r.URL.Path).Observe(duration)
	})
}

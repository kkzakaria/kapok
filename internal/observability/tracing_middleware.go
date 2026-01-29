package observability

import (
	"net/http"

	"github.com/kapok/kapok/internal/tenant"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware wraps HTTP handlers with OpenTelemetry tracing.
type TracingMiddleware struct{}

// NewTracingMiddleware creates a new TracingMiddleware.
func NewTracingMiddleware() *TracingMiddleware {
	return &TracingMiddleware{}
}

// Middleware returns an HTTP middleware that adds tracing spans.
func (tm *TracingMiddleware) Middleware(next http.Handler) http.Handler {
	handler := otelhttp.NewHandler(next, "http.request",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return r.Method + " " + r.URL.Path
		}),
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)

		// Add tenant_id as span attribute if available
		tenantID, err := tenant.GetTenantID(r.Context())
		if err == nil {
			span := trace.SpanFromContext(r.Context())
			span.SetAttributes(attribute.String("tenant_id", tenantID))
		}
	})
}

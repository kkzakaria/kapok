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
	// Wrap the inner handler to set tenant_id on the span before it ends.
	// otelhttp.NewHandler creates and finishes the span around the inner handler,
	// so attributes must be set inside the wrapped handler, not after ServeHTTP returns.
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tenantID, err := tenant.GetTenantID(r.Context()); err == nil {
			span := trace.SpanFromContext(r.Context())
			span.SetAttributes(attribute.String("tenant_id", tenantID))
		}
		next.ServeHTTP(w, r)
	})
	return otelhttp.NewHandler(inner, "http.request",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return r.Method + " " + r.URL.Path
		}),
	)
}

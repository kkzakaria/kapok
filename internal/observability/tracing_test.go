package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestTracingProvider_Shutdown(t *testing.T) {
	// Test shutdown with nil provider
	tp := &TracingProvider{}
	err := tp.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestTracingProvider_InitAndShutdown(t *testing.T) {
	// Use an in-memory span exporter to avoid needing a real OTLP endpoint.
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(provider)

	tp := &TracingProvider{provider: provider}

	// Create a span to verify the provider works
	tracer := provider.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "test-span", spans[0].Name)

	// Shutdown should succeed
	err := tp.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestTracingMiddleware_WithInMemoryExporter(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(provider)
	defer provider.Shutdown(context.Background())

	tm := NewTracingMiddleware()
	handler := tm.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	spans := exporter.GetSpans()
	require.NotEmpty(t, spans)
	assert.Equal(t, "GET /api/test", spans[0].Name)
}

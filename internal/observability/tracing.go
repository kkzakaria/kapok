package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// TracingProvider manages OpenTelemetry tracing.
type TracingProvider struct {
	provider *sdktrace.TracerProvider
}

// NewTracingProvider initializes an OTel TracerProvider with OTLP/HTTP exporter.
func NewTracingProvider(ctx context.Context, serviceName, jaegerEndpoint string, sampleRate float64) (*TracingProvider, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(jaegerEndpoint),
		otlptracehttp.WithInsecure(),
	}
	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(sampleRate)),
	)
	otel.SetTracerProvider(tp)

	return &TracingProvider{provider: tp}, nil
}

// Shutdown gracefully shuts down the tracer provider.
func (tp *TracingProvider) Shutdown(ctx context.Context) error {
	if tp.provider != nil {
		return tp.provider.Shutdown(ctx)
	}
	return nil
}

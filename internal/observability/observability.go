package observability

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

// Config holds observability configuration.
type Config struct {
	Enabled        bool
	MetricsPort    int
	TracingEnabled bool
	SampleRate     float64
	JaegerEndpoint string
	ServiceName    string
}

// Observability is the main facade for all observability components.
type Observability struct {
	Metrics *MetricsCollector
	Tracing *TracingProvider
	Health  *HealthChecker
	logger  zerolog.Logger
	registry *prometheus.Registry
}

// New initializes all observability components.
func New(ctx context.Context, cfg Config, logger zerolog.Logger) (*Observability, error) {
	registry := prometheus.NewRegistry()
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	metrics := NewMetricsCollector(registry)
	health := NewHealthChecker(logger)

	obs := &Observability{
		Metrics:  metrics,
		Health:   health,
		logger:   logger,
		registry: registry,
	}

	if cfg.TracingEnabled {
		serviceName := cfg.ServiceName
		if serviceName == "" {
			serviceName = "kapok"
		}
		tp, err := NewTracingProvider(ctx, serviceName, cfg.JaegerEndpoint, cfg.SampleRate)
		if err != nil {
			logger.Warn().Err(err).Msg("failed to initialize tracing, continuing without it")
		} else {
			obs.Tracing = tp
		}
	}

	return obs, nil
}

// HTTPMiddleware returns the combined metrics + tracing middleware.
func (o *Observability) HTTPMiddleware(next http.Handler) http.Handler {
	metricsMiddleware := NewMetricsMiddleware(o.Metrics, o.logger)
	handler := metricsMiddleware.Middleware(next)

	if o.Tracing != nil {
		tracingMiddleware := NewTracingMiddleware()
		handler = tracingMiddleware.Middleware(handler)
	}

	return handler
}

// MetricsHandler returns the Prometheus metrics HTTP handler.
func (o *Observability) MetricsHandler() http.Handler {
	return promhttp.HandlerFor(o.registry, promhttp.HandlerOpts{})
}

// Shutdown gracefully shuts down all observability components.
func (o *Observability) Shutdown(ctx context.Context) error {
	if o.Tracing != nil {
		if err := o.Tracing.Shutdown(ctx); err != nil {
			o.logger.Error().Err(err).Msg("failed to shutdown tracing")
			return err
		}
	}
	return nil
}

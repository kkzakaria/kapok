# PR #24 Review: Epic 5 - Observability & Monitoring

## Summary

This PR adds a full observability stack to Kapok: Prometheus metrics, OpenTelemetry tracing, health checks, Grafana dashboards, alert rules, and Helm chart generation for deploying the stack on Kubernetes.

**Files changed:** 21 | **+1,303 / -51**

**Tests:** All pass (observability + config packages).

---

## Architecture Assessment

The design is well-structured with a clean facade pattern (`Observability` struct) that composes metrics, tracing, and health checking. The separation into individual files (metrics, tracing, health, middlewares) is appropriate.

---

## Issues Found

### Critical

1. **Tracing middleware sets span attributes after the span ends** (`internal/observability/tracing_middleware.go:31-34`)
   - `otelhttp.NewHandler` creates and finishes the span within `handler.ServeHTTP()`. After that call returns, `trace.SpanFromContext(r.Context())` returns the already-ended span. Calling `SetAttributes` on a finished span is a no-op.
   - **Fix:** Use `otelhttp.WithSpanOptions` or wrap the inner handler to set tenant_id *before* the span ends, e.g.:
     ```go
     handler := otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
         if tid, err := tenant.GetTenantID(r.Context()); err == nil {
             trace.SpanFromContext(r.Context()).SetAttributes(attribute.String("tenant_id", tid))
         }
         next.ServeHTTP(w, r)
     }), "http.request", ...)
     ```

2. **`responseWriter` breaks `http.Hijacker`/`http.Flusher` interfaces** (`internal/observability/metrics_middleware.go:14-33`)
   - Wrapping `http.ResponseWriter` without delegating `Hijack()` and `Flush()` will break WebSocket upgrades and SSE streaming. This is a common Go pitfall.
   - **Fix:** Implement optional interface delegation or use a library like `httpsnoop`.

### Medium

3. **High-cardinality `path` label on HTTP metrics** (`internal/observability/metrics_middleware.go:57`)
   - Using `r.URL.Path` directly means paths like `/api/tenants/abc123/resources/xyz` create unbounded label values, which will cause Prometheus memory issues in production.
   - **Fix:** Use route patterns (e.g. from chi's `RouteContext`) instead of raw paths.

4. **Config validation runs even when observability is disabled** (`pkg/config/config.go:133-138`)
   - A user setting `observability.enabled: false` with default zero values for `MetricsPort` (0) will fail validation. The validation should be gated on `Enabled`.

5. **Grafana password defaults to `"admin"` in CLI flags** (`cmd/kapok/cmd/deploy.go:43`)
   - This is a security concern. The default should be empty, forcing the user to set it explicitly, or generate a random password.

6. **Observability values.yaml is concatenated, not structured** (`internal/k8s/helm.go:97`)
   - `PrometheusValuesYAML + "\n" + LokiValuesYAML + ...` concatenates raw YAML strings. This is fragile — a missing newline or indentation issue would silently produce invalid YAML. Consider using a single template or marshaling a struct.

### Low

7. **`tracing_test.go` only tests nil shutdown** — no test for actual `NewTracingProvider` initialization. Understandable given it needs a real OTLP endpoint, but a test with an in-memory exporter would improve coverage.

8. **Dashboard JSON constants** (`internal/k8s/dashboards.go`) — These 164 lines of JSON constants would be better as embedded files (`//go:embed`) for maintainability.

9. **`TenantCPUUsage`, `TenantMemoryUsage`, `TenantStorageUsage` gauges** (`internal/observability/metrics.go:60-69`) are registered but never written to anywhere in this PR. There's no collector or periodic scraper populating them.

10. **`DBConnectionPoolExhausted` alert logic** (`internal/k8s/observability.go:193-194`) — `rate(...) == 0 and metric > 0` is fragile. A brief zero-rate during low traffic at night would fire this alert. Consider using connection pool-specific metrics instead.

---

## What's Good

- Clean facade pattern with `Observability.HTTPMiddleware()` composing both concerns
- Proper use of `prometheus.NewRegistry()` (not the global default) — testable and avoids conflicts
- Health check design with liveness/readiness separation following Kubernetes conventions
- Graceful degradation: tracing init failure logs a warning and continues without it
- Alert rules cover meaningful scenarios (error rate, latency, storage, DB)
- Per-tenant dashboard with template variables is a nice touch

---

## Verdict

**Request changes.** The tracing middleware bug (#1) means tenant attribution in traces is silently broken. The high-cardinality path label (#3) will cause production issues. Both should be fixed before merge. The other medium issues are worth addressing but not blocking.

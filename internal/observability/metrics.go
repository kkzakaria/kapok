package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsCollector holds all Prometheus metrics for the platform.
type MetricsCollector struct {
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	GraphQLQueriesTotal  *prometheus.CounterVec
	GraphQLQueryDuration *prometheus.HistogramVec
	GraphQLErrorsTotal   *prometheus.CounterVec
	DBQueriesTotal       *prometheus.CounterVec
	DBQueryDuration      *prometheus.HistogramVec
	TenantCPUUsage       *prometheus.GaugeVec
	TenantMemoryUsage    *prometheus.GaugeVec
	TenantStorageUsage   *prometheus.GaugeVec
	BackupsTotal         *prometheus.CounterVec
	BackupDuration       *prometheus.HistogramVec
	BackupSizeBytes      *prometheus.HistogramVec
	RestoresTotal        *prometheus.CounterVec
	LastBackupTimestamp  *prometheus.GaugeVec
}

// NewMetricsCollector creates and registers all Prometheus metrics.
func NewMetricsCollector(reg prometheus.Registerer) *MetricsCollector {
	factory := promauto.With(reg)
	return &MetricsCollector{
		HTTPRequestsTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "kapok_http_requests_total",
			Help: "Total number of HTTP requests",
		}, []string{"tenant_id", "method", "path", "status"}),
		HTTPRequestDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "kapok_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{"tenant_id", "method", "path"}),
		GraphQLQueriesTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "kapok_graphql_queries_total",
			Help: "Total number of GraphQL queries",
		}, []string{"tenant_id", "operation_type", "operation_name"}),
		GraphQLQueryDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "kapok_graphql_query_duration_seconds",
			Help:    "GraphQL query duration in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{"tenant_id", "operation_type", "operation_name"}),
		GraphQLErrorsTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "kapok_graphql_errors_total",
			Help: "Total number of GraphQL errors",
		}, []string{"tenant_id", "operation_type"}),
		DBQueriesTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "kapok_db_queries_total",
			Help: "Total number of database queries",
		}, []string{"operation"}),
		DBQueryDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "kapok_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{"operation"}),
		TenantCPUUsage: factory.NewGaugeVec(prometheus.GaugeOpts{
			Name: "kapok_tenant_cpu_usage",
			Help: "CPU usage per tenant",
		}, []string{"tenant_id"}),
		TenantMemoryUsage: factory.NewGaugeVec(prometheus.GaugeOpts{
			Name: "kapok_tenant_memory_usage",
			Help: "Memory usage per tenant",
		}, []string{"tenant_id"}),
		TenantStorageUsage: factory.NewGaugeVec(prometheus.GaugeOpts{
			Name: "kapok_tenant_storage_usage",
			Help: "Storage usage per tenant",
		}, []string{"tenant_id"}),
		BackupsTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "kapok_backups_total",
			Help: "Total number of backups performed",
		}, []string{"tenant_id", "status", "trigger"}),
		BackupDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "kapok_backup_duration_seconds",
			Help:    "Backup duration in seconds",
			Buckets: prometheus.ExponentialBuckets(1, 2, 12),
		}, []string{"tenant_id"}),
		BackupSizeBytes: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "kapok_backup_size_bytes",
			Help:    "Backup size in bytes",
			Buckets: prometheus.ExponentialBuckets(1024, 4, 10),
		}, []string{"tenant_id"}),
		RestoresTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "kapok_restores_total",
			Help: "Total number of restores performed",
		}, []string{"tenant_id", "status"}),
		LastBackupTimestamp: factory.NewGaugeVec(prometheus.GaugeOpts{
			Name: "kapok_last_backup_timestamp",
			Help: "Unix timestamp of last successful backup per tenant",
		}, []string{"tenant_id"}),
	}
}

// SetTenantResourceUsage updates the resource usage gauges for a tenant.
func (mc *MetricsCollector) SetTenantResourceUsage(tenantID string, cpu, memory, storage float64) {
	mc.TenantCPUUsage.WithLabelValues(tenantID).Set(cpu)
	mc.TenantMemoryUsage.WithLabelValues(tenantID).Set(memory)
	mc.TenantStorageUsage.WithLabelValues(tenantID).Set(storage)
}

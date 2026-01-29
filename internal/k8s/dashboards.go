package k8s

// Grafana dashboard JSON constants for ConfigMap-based provisioning.

// DashboardPlatformOverview is the Grafana dashboard for platform-wide metrics.
const DashboardPlatformOverview = `{
  "dashboard": {
    "title": "Kapok Platform Overview",
    "uid": "kapok-platform-overview",
    "tags": ["kapok"],
    "timezone": "browser",
    "panels": [
      {
        "title": "Request Rate",
        "type": "timeseries",
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0},
        "targets": [{"expr": "sum(rate(kapok_http_requests_total[5m]))"}]
      },
      {
        "title": "Error Rate",
        "type": "timeseries",
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0},
        "targets": [{"expr": "sum(rate(kapok_http_requests_total{status=~\"5..\"}[5m])) / sum(rate(kapok_http_requests_total[5m]))"}]
      },
      {
        "title": "Active Tenants",
        "type": "stat",
        "gridPos": {"h": 4, "w": 6, "x": 0, "y": 8},
        "targets": [{"expr": "count(count by (tenant_id) (kapok_http_requests_total))"}]
      },
      {
        "title": "P95 Latency",
        "type": "gauge",
        "gridPos": {"h": 4, "w": 6, "x": 6, "y": 8},
        "targets": [{"expr": "histogram_quantile(0.95, sum(rate(kapok_http_request_duration_seconds_bucket[5m])) by (le))"}]
      }
    ],
    "templating": {"list": []},
    "time": {"from": "now-1h", "to": "now"},
    "refresh": "10s"
  }
}`

// DashboardPerTenant is the Grafana dashboard for per-tenant metrics.
const DashboardPerTenant = `{
  "dashboard": {
    "title": "Kapok Per-Tenant Metrics",
    "uid": "kapok-per-tenant",
    "tags": ["kapok", "tenant"],
    "timezone": "browser",
    "templating": {
      "list": [
        {
          "name": "tenant_id",
          "type": "query",
          "datasource": "Prometheus",
          "query": "label_values(kapok_http_requests_total, tenant_id)",
          "refresh": 2,
          "sort": 1
        }
      ]
    },
    "panels": [
      {
        "title": "CPU Usage",
        "type": "timeseries",
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0},
        "targets": [{"expr": "kapok_tenant_cpu_usage{tenant_id=\"$tenant_id\"}"}]
      },
      {
        "title": "Memory Usage",
        "type": "timeseries",
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0},
        "targets": [{"expr": "kapok_tenant_memory_usage{tenant_id=\"$tenant_id\"}"}]
      },
      {
        "title": "Storage Usage",
        "type": "gauge",
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 8},
        "targets": [{"expr": "kapok_tenant_storage_usage{tenant_id=\"$tenant_id\"}"}]
      },
      {
        "title": "Request Rate",
        "type": "timeseries",
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 8},
        "targets": [{"expr": "sum(rate(kapok_http_requests_total{tenant_id=\"$tenant_id\"}[5m]))"}]
      }
    ],
    "time": {"from": "now-1h", "to": "now"},
    "refresh": "10s"
  }
}`

// DashboardGraphQL is the Grafana dashboard for GraphQL performance metrics.
const DashboardGraphQL = `{
  "dashboard": {
    "title": "Kapok GraphQL Performance",
    "uid": "kapok-graphql",
    "tags": ["kapok", "graphql"],
    "timezone": "browser",
    "panels": [
      {
        "title": "Query Rate by Operation",
        "type": "timeseries",
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0},
        "targets": [{"expr": "sum(rate(kapok_graphql_queries_total[5m])) by (operation_type, operation_name)"}]
      },
      {
        "title": "P95 Query Duration",
        "type": "timeseries",
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0},
        "targets": [{"expr": "histogram_quantile(0.95, sum(rate(kapok_graphql_query_duration_seconds_bucket[5m])) by (le, operation_name))"}]
      },
      {
        "title": "GraphQL Errors",
        "type": "timeseries",
        "gridPos": {"h": 8, "w": 24, "x": 0, "y": 8},
        "targets": [{"expr": "sum(rate(kapok_graphql_errors_total[5m])) by (tenant_id, operation_type)"}]
      }
    ],
    "templating": {"list": []},
    "time": {"from": "now-1h", "to": "now"},
    "refresh": "10s"
  }
}`

// DashboardInfrastructure is the Grafana dashboard for infrastructure health.
const DashboardInfrastructure = `{
  "dashboard": {
    "title": "Kapok Infrastructure Health",
    "uid": "kapok-infrastructure",
    "tags": ["kapok", "infrastructure"],
    "timezone": "browser",
    "panels": [
      {
        "title": "DB Query Rate",
        "type": "timeseries",
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0},
        "targets": [{"expr": "sum(rate(kapok_db_queries_total[5m])) by (operation)"}]
      },
      {
        "title": "DB Query Duration (P95)",
        "type": "timeseries",
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0},
        "targets": [{"expr": "histogram_quantile(0.95, sum(rate(kapok_db_query_duration_seconds_bucket[5m])) by (le))"}]
      },
      {
        "title": "Pod CPU Usage",
        "type": "timeseries",
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 8},
        "targets": [{"expr": "rate(container_cpu_usage_seconds_total{namespace=\"kapok\"}[5m])"}]
      },
      {
        "title": "Pod Memory Usage",
        "type": "timeseries",
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 8},
        "targets": [{"expr": "container_memory_usage_bytes{namespace=\"kapok\"}"}]
      }
    ],
    "templating": {"list": []},
    "time": {"from": "now-1h", "to": "now"},
    "refresh": "30s"
  }
}`

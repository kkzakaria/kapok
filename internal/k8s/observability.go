package k8s

// Observability Helm chart template constants.

// PrometheusValuesYAML contains kube-prometheus-stack subchart values.
const PrometheusValuesYAML = `prometheus:
  enabled: true
  server:
    retention: 30d
    global:
      scrape_interval: 15s
      evaluation_interval: 15s
    scrapeConfigs:
      - job_name: kapok-control-plane
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          - source_labels: [__meta_kubernetes_pod_label_app]
            regex: control-plane
            action: keep
          - source_labels: [__meta_kubernetes_pod_ip]
            target_label: __address__
            replacement: "$1:9090"
      - job_name: kapok-graphql-engine
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          - source_labels: [__meta_kubernetes_pod_label_app]
            regex: graphql-engine
            action: keep
          - source_labels: [__meta_kubernetes_pod_ip]
            target_label: __address__
            replacement: "$1:9090"
      - job_name: kapok-provisioner
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          - source_labels: [__meta_kubernetes_pod_label_app]
            regex: provisioner
            action: keep
          - source_labels: [__meta_kubernetes_pod_ip]
            target_label: __address__
            replacement: "$1:9090"
`

// GrafanaValuesYAML contains Grafana subchart values.
const GrafanaValuesYAML = `grafana:
  enabled: true
  adminPassword: "{{ .GrafanaPassword }}"
  datasources:
    datasources.yaml:
      apiVersion: 1
      datasources:
        - name: Prometheus
          type: prometheus
          url: http://prometheus-server:9090
          access: proxy
          isDefault: true
        - name: Loki
          type: loki
          url: http://loki:3100
          access: proxy
  dashboardProviders:
    dashboardproviders.yaml:
      apiVersion: 1
      providers:
        - name: kapok
          orgId: 1
          folder: Kapok
          type: file
          disableDeletion: false
          editable: true
          options:
            path: /var/lib/grafana/dashboards/kapok
`

// LokiValuesYAML contains Loki subchart values.
const LokiValuesYAML = `loki:
  enabled: true
  config:
    limits_config:
      retention_period: 168h
    schema_config:
      configs:
        - from: "2024-01-01"
          store: tsdb
          object_store: filesystem
          schema: v13
          index:
            prefix: index_
            period: 24h
    storage_config:
      filesystem:
        directory: /loki/chunks
`

// JaegerValuesYAML contains Jaeger all-in-one subchart values.
const JaegerValuesYAML = `jaeger:
  enabled: true
  allInOne:
    enabled: true
  collector:
    service:
      otlp:
        http:
          name: otlp-http
          port: 4318
        grpc:
          name: otlp-grpc
          port: 4317
  query:
    service:
      port: 16686
`

// AlertManagerConfigYAML contains AlertManager routing configuration.
const AlertManagerConfigYAML = `alertmanager:
  config:
    global:
      resolve_timeout: 5m
    route:
      group_by:
        - alertname
        - tenant_id
      group_wait: 10s
      group_interval: 10s
      repeat_interval: 1h
      receiver: default
      routes:
        - match:
            severity: critical
          receiver: pagerduty
        - match:
            severity: warning
          receiver: slack
        - match:
            severity: info
          receiver: default
    receivers:
      - name: default
      - name: pagerduty
        pagerduty_configs:
          - service_key: "{{ .PagerDutyKey }}"
      - name: slack
        slack_configs:
          - api_url: "{{ .SlackWebhook }}"
            channel: "#kapok-alerts"
            title: '{{ "{{ .GroupLabels.alertname }}" }}'
            text: '{{ "{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}" }}'
`

// AlertRulesYAML contains Prometheus alert rules.
const AlertRulesYAML = `apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-alert-rules
  namespace: {{ .Values.global.namespace }}
  labels:
    prometheus: kapok
data:
  kapok-alerts.yaml: |
    groups:
      - name: kapok.rules
        rules:
          - alert: HighErrorRate
            expr: rate(kapok_http_requests_total{status=~"5.."}[5m]) / rate(kapok_http_requests_total[5m]) > 0.05
            for: 5m
            labels:
              severity: critical
            annotations:
              summary: "High error rate detected"
              description: "More than 5% of requests are returning 5xx errors"
          - alert: HighLatency
            expr: histogram_quantile(0.95, rate(kapok_http_request_duration_seconds_bucket[5m])) > 2
            for: 5m
            labels:
              severity: warning
            annotations:
              summary: "High request latency"
              description: "95th percentile latency is above 2 seconds"
          - alert: TenantStorageFull
            expr: kapok_tenant_storage_usage > 0.9
            for: 10m
            labels:
              severity: warning
            annotations:
              summary: "Tenant storage nearly full"
              description: "Tenant {{ "{{ $labels.tenant_id }}" }} storage usage is above 90%"
          - alert: DBConnectionPoolExhausted
            expr: rate(kapok_db_queries_total[1m]) == 0 and kapok_db_queries_total > 0
            for: 2m
            labels:
              severity: critical
            annotations:
              summary: "Database connection pool may be exhausted"
              description: "No database queries processed in the last minute"
`

// ObservabilityValuesYAML is the combined values template for the observability subchart.
// Using a single constant avoids fragile string concatenation.
var ObservabilityValuesYAML = PrometheusValuesYAML + "\n" + LokiValuesYAML + "\n" + JaegerValuesYAML + "\n" + GrafanaValuesYAML + "\n" + AlertManagerConfigYAML

// ObservabilityChartYAML is the Chart.yaml for the observability subchart.
const ObservabilityChartYAML = `apiVersion: v2
name: observability
description: Kapok Observability Stack (Prometheus, Grafana, Loki, Jaeger)
type: application
version: 0.1.0
appVersion: "1.0.0"
dependencies:
  - name: kube-prometheus-stack
    version: ">=45.0.0"
    repository: https://prometheus-community.github.io/helm-charts
    condition: prometheus.enabled
  - name: loki
    version: ">=5.0.0"
    repository: https://grafana.github.io/helm-charts
    condition: loki.enabled
  - name: jaeger
    version: ">=0.71.0"
    repository: https://jaegertracing.github.io/helm-charts
    condition: jaeger.enabled
`

package k8s

// KEDAScaledObjectYAML is a Helm template written verbatim to disk via writeRaw.
// All {{ }} directives are standard Helm/Go template syntax, evaluated by Helm at install time.
const KEDAScaledObjectYAML = `{{- if .Values.global.keda.enabled }}
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: {{ .Release.Name }}-graphql-engine-keda
  namespace: {{ .Values.global.namespace }}
spec:
  scaleTargetRef:
    name: {{ .Release.Name }}-graphql-engine
  pollingInterval: 15
  cooldownPeriod: 300
  minReplicaCount: 2
  maxReplicaCount: 50
  triggers:
    - type: postgresql
      metadata:
        connectionFromEnv: KAPOK_DATABASE_URL
        query: "SELECT count(*) FROM pg_stat_activity WHERE state = 'active'"
        targetQueryValue: "80"
        activationTargetQueryValue: "40"
    - type: prometheus
      metadata:
        serverAddress: http://prometheus.monitoring:9090
        metricName: http_request_duration_seconds
        query: histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{service="graphql-engine"}[5m])) by (le))
        threshold: "0.2"
        activationThreshold: "0.1"
{{- end }}
`

package k8s

// Template constants for Helm chart generation.
// ValuesYAML and SubchartChartYAML use Go text/template (rendered with ChartConfig/subchartMeta).
// All other templates are Helm templates written verbatim — subchart-specific values
// (component name, path prefix) are injected via fmt.Sprintf with %s placeholders.

const ChartYAML = `apiVersion: v2
name: kapok-platform
description: Kapok BaaS Platform
type: application
version: 0.1.0
appVersion: "1.0.0"
dependencies:
  - name: control-plane
    version: 0.1.0
    repository: file://charts/control-plane
  - name: graphql-engine
    version: 0.1.0
    repository: file://charts/graphql-engine
  - name: provisioner
    version: 0.1.0
    repository: file://charts/provisioner
`

const ValuesYAML = `global:
  cloud: {{ .Cloud }}
  namespace: {{ .Namespace }}
  domain: {{ .Domain }}
  imageTag: {{ .ImageTag }}
  storageClass: {{ .StorageClass }}
  ingressClass: {{ .IngressClass }}
  tls:
    enabled: {{ .TLSEnabled }}
  hpa:
    enabled: {{ .HPAEnabled }}
  keda:
    enabled: {{ .KEDAEnabled }}
  secrets:
    databasePassword: ""
    jwtSecret: ""
    databaseURL: ""
    redisPassword: ""
  observability:
    enabled: {{ .ObservabilityEnabled }}
    metricsPort: 9090

control-plane:
  replicaCount: 2
  image:
    repository: kapok/control-plane
    tag: {{ .ImageTag }}
  resources:
    requests:
      cpu: 250m
      memory: 256Mi
    limits:
      cpu: "1"
      memory: 512Mi

graphql-engine:
  replicaCount: 2
  image:
    repository: kapok/graphql-engine
    tag: {{ .ImageTag }}
  resources:
    requests:
      cpu: 500m
      memory: 512Mi
    limits:
      cpu: "2"
      memory: 1Gi

provisioner:
  replicaCount: 1
  image:
    repository: kapok/provisioner
    tag: {{ .ImageTag }}
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 256Mi
`

const SubchartChartYAML = `apiVersion: v2
name: {{ .Name }}
description: {{ .Description }}
type: application
version: 0.1.0
appVersion: "1.0.0"
`

// DeploymentYAMLTmpl uses %s as a placeholder for the component name (replaced via strings.ReplaceAll).
// All {{ }} are literal Helm template directives.
const DeploymentYAMLTmpl = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}-%s
  namespace: {{ .Values.global.namespace }}
  labels:
    app: %s
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"
    spec:
      containers:
        - name: %s
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          ports:
            - containerPort: 8080
            - containerPort: 9090
              name: metrics
          resources:
            {{- toYaml .Values.resources | nindent 12 }}
          livenessProbe:
            httpGet:
              path: /healthz
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 15
          readinessProbe:
            httpGet:
              path: /readyz
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
          envFrom:
            - configMapRef:
                name: {{ .Release.Name }}-%s-config
            - secretRef:
                name: {{ .Release.Name }}-secrets
`

// ServiceYAMLTmpl uses %s as a placeholder for the component name.
const ServiceYAMLTmpl = `apiVersion: v1
kind: Service
metadata:
  name: {{ .Release.Name }}-%s
  namespace: {{ .Values.global.namespace }}
spec:
  selector:
    app: %s
  ports:
    - port: 80
      targetPort: 8080
      protocol: TCP
  type: ClusterIP
`

// IngressYAMLTmpl uses %s for the component name and %PATH% for the path prefix.
const IngressYAMLTmpl = `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ .Release.Name }}-%s
  namespace: {{ .Values.global.namespace }}
  annotations:
    kubernetes.io/ingress.class: {{ .Values.global.ingressClass }}
    {{- if .Values.global.tls.enabled }}
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    {{- end }}
spec:
  {{- if .Values.global.tls.enabled }}
  tls:
    - hosts:
        - {{ .Values.global.domain }}
      secretName: %s-tls
  {{- end }}
  rules:
    - host: {{ .Values.global.domain }}
      http:
        paths:
          - path: %PATH%
            pathType: Prefix
            backend:
              service:
                name: {{ .Release.Name }}-%s
                port:
                  number: 80
`

// configMapTemplates holds per-service ConfigMap templates with service-specific env vars.
var configMapTemplates = map[string]string{
	"control-plane": `apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-control-plane-config
  namespace: {{ .Values.global.namespace }}
data:
  KAPOK_SERVER_HOST: "0.0.0.0"
  KAPOK_SERVER_PORT: "8080"
  KAPOK_LOG_LEVEL: "info"
  KAPOK_LOG_FORMAT: "json"
  KAPOK_SERVICE_ROLE: "control-plane"
  KAPOK_OBSERVABILITY_ENABLED: "true"
  KAPOK_OBSERVABILITY_METRICS_PORT: "9090"
  KAPOK_OBSERVABILITY_TRACING_ENABLED: "true"
  KAPOK_OBSERVABILITY_TRACING_SAMPLE_RATE: "0.1"
  KAPOK_OBSERVABILITY_JAEGER_ENDPOINT: "jaeger-collector:4318"
`,
	"graphql-engine": `apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-graphql-engine-config
  namespace: {{ .Values.global.namespace }}
data:
  KAPOK_SERVER_HOST: "0.0.0.0"
  KAPOK_SERVER_PORT: "8080"
  KAPOK_LOG_LEVEL: "info"
  KAPOK_LOG_FORMAT: "json"
  KAPOK_SERVICE_ROLE: "graphql-engine"
  KAPOK_GRAPHQL_INTROSPECTION: "true"
  KAPOK_GRAPHQL_CACHE_TTL: "300"
  KAPOK_OBSERVABILITY_ENABLED: "true"
  KAPOK_OBSERVABILITY_METRICS_PORT: "9090"
  KAPOK_OBSERVABILITY_TRACING_ENABLED: "true"
  KAPOK_OBSERVABILITY_TRACING_SAMPLE_RATE: "0.1"
  KAPOK_OBSERVABILITY_JAEGER_ENDPOINT: "jaeger-collector:4318"
`,
	"provisioner": `apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-provisioner-config
  namespace: {{ .Values.global.namespace }}
data:
  KAPOK_SERVER_HOST: "0.0.0.0"
  KAPOK_SERVER_PORT: "8080"
  KAPOK_LOG_LEVEL: "info"
  KAPOK_LOG_FORMAT: "json"
  KAPOK_SERVICE_ROLE: "provisioner"
  KAPOK_PROVISIONER_SCHEMA_PREFIX: "tenant_"
  KAPOK_OBSERVABILITY_ENABLED: "true"
  KAPOK_OBSERVABILITY_METRICS_PORT: "9090"
  KAPOK_OBSERVABILITY_TRACING_ENABLED: "true"
  KAPOK_OBSERVABILITY_TRACING_SAMPLE_RATE: "0.1"
  KAPOK_OBSERVABILITY_JAEGER_ENDPOINT: "jaeger-collector:4318"
`,
}

// SecretYAML is a Helm template for the shared platform secrets.
// Values are expected to be provided via --set or a values override at deploy time.
const SecretYAML = `apiVersion: v1
kind: Secret
metadata:
  name: {{ .Release.Name }}-secrets
  namespace: {{ .Values.global.namespace }}
type: Opaque
stringData:
  KAPOK_DATABASE_PASSWORD: {{ .Values.global.secrets.databasePassword | default "" | quote }}
  KAPOK_JWT_SECRET: {{ .Values.global.secrets.jwtSecret | default "" | quote }}
  KAPOK_DATABASE_URL: {{ .Values.global.secrets.databaseURL | default "" | quote }}
  KAPOK_REDIS_PASSWORD: {{ .Values.global.secrets.redisPassword | default "" | quote }}
`

const NamespaceYAML = `apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Values.global.namespace }}
  labels:
    app.kubernetes.io/managed-by: kapok
`

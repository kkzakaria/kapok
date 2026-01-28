package k8s

const CertManagerYAML = `{{- if (index .Values.global.tls "enabled") }}
apiVersion: v1
kind: Namespace
metadata:
  name: cert-manager
  labels:
    app.kubernetes.io/managed-by: kapok
{{- end }}
`

const ClusterIssuerYAML = `{{- if (index .Values.global.tls "enabled") }}
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: {{ "{{ .Values.global.tls.email | default \"admin@example.com\" }}" }}
    privateKeySecretRef:
      name: letsencrypt-prod-key
    solvers:
      - http01:
          ingress:
            class: {{ "{{ .Values.global.ingressClass }}" }}
{{- end }}
`

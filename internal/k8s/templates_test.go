package k8s

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateConstants_NotEmpty(t *testing.T) {
	templates := map[string]string{
		"ChartYAML":            ChartYAML,
		"NamespaceYAML":        NamespaceYAML,
		"SecretYAML":           SecretYAML,
		"DeploymentYAMLTmpl":   DeploymentYAMLTmpl,
		"ServiceYAMLTmpl":      ServiceYAMLTmpl,
		"IngressYAMLTmpl":      IngressYAMLTmpl,
		"HPAYAMLTmpl":          HPAYAMLTmpl,
		"CertManagerYAML":      CertManagerYAML,
		"ClusterIssuerYAML":    ClusterIssuerYAML,
		"KEDAScaledObjectYAML": KEDAScaledObjectYAML,
	}

	for name, tmpl := range templates {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, tmpl)
		})
	}
}

func TestValuesTemplate_Renders(t *testing.T) {
	tmpl, err := template.New("values").Parse(ValuesYAML)
	require.NoError(t, err)

	cfg := ChartConfig{
		Cloud:        CloudAWS,
		Namespace:    "test-ns",
		Domain:       "test.com",
		ImageTag:     "v1",
		StorageClass: "gp3",
		IngressClass: "alb",
		TLSEnabled:   true,
		HPAEnabled:   true,
		KEDAEnabled:  false,
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, cfg)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "cloud: aws")
	assert.Contains(t, output, "namespace: test-ns")
	assert.Contains(t, output, "domain: test.com")
	assert.Contains(t, output, "storageClass: gp3")
	assert.Contains(t, output, "enabled: true")
}

func TestSubchartChartTemplate_Renders(t *testing.T) {
	tmpl, err := template.New("subchart").Parse(SubchartChartYAML)
	require.NoError(t, err)

	data := subchartMeta{Name: "control-plane", Description: "Kapok Control Plane"}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "name: control-plane")
	assert.Contains(t, output, "description: Kapok Control Plane")
}

func TestTemplates_ReplaceAll(t *testing.T) {
	name := "control-plane"
	replace := func(tmpl string) string {
		return strings.ReplaceAll(tmpl, "%s", name)
	}

	deployment := replace(DeploymentYAMLTmpl)
	assert.Contains(t, deployment, "name: {{ .Release.Name }}-control-plane")
	assert.Contains(t, deployment, "app: control-plane")

	service := replace(ServiceYAMLTmpl)
	assert.Contains(t, service, "name: {{ .Release.Name }}-control-plane")

	ingress := strings.ReplaceAll(strings.Replace(IngressYAMLTmpl, "%PATH%", "/api", 1), "%s", name)
	assert.Contains(t, ingress, "path: /api")
	assert.Contains(t, ingress, "secretName: control-plane-tls")

	hpa := replace(HPAYAMLTmpl)
	assert.Contains(t, hpa, "name: {{ .Release.Name }}-control-plane")
	assert.Contains(t, hpa, "averageUtilization: 70")
}

func TestConfigMapTemplates_PerService(t *testing.T) {
	for _, name := range []string{"control-plane", "graphql-engine", "provisioner"} {
		t.Run(name, func(t *testing.T) {
			tmpl, ok := configMapTemplates[name]
			assert.True(t, ok, "configmap template should exist for %s", name)
			assert.Contains(t, tmpl, "KAPOK_SERVICE_ROLE")
			assert.Contains(t, tmpl, name)
		})
	}
}

func TestSecretTemplate_NotEmpty(t *testing.T) {
	assert.Contains(t, SecretYAML, "KAPOK_DATABASE_PASSWORD")
	assert.Contains(t, SecretYAML, "KAPOK_JWT_SECRET")
}

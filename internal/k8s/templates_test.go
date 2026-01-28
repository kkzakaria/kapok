package k8s

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateConstants_NotEmpty(t *testing.T) {
	templates := map[string]string{
		"ChartYAML":            ChartYAML,
		"NamespaceYAML":        NamespaceYAML,
		"DeploymentYAMLFmt":    DeploymentYAMLFmt,
		"ServiceYAMLFmt":       ServiceYAMLFmt,
		"IngressYAMLFmt":       IngressYAMLFmt,
		"ConfigMapYAMLFmt":     ConfigMapYAMLFmt,
		"HPAYAMLFmt":           HPAYAMLFmt,
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

func TestFmtTemplates_Sprintf(t *testing.T) {
	// Verify fmt.Sprintf works correctly with the Fmt templates
	name := "control-plane"
	path := "/api"

	deployment := fmt.Sprintf(DeploymentYAMLFmt, name, name, name, name, name, name)
	assert.Contains(t, deployment, "name: {{ .Release.Name }}-control-plane")
	assert.Contains(t, deployment, "app: control-plane")

	service := fmt.Sprintf(ServiceYAMLFmt, name, name)
	assert.Contains(t, service, "name: {{ .Release.Name }}-control-plane")

	ingress := fmt.Sprintf(IngressYAMLFmt, name, name, path, name)
	assert.Contains(t, ingress, "path: /api")
	assert.Contains(t, ingress, "secretName: control-plane-tls")

	configmap := fmt.Sprintf(ConfigMapYAMLFmt, name)
	assert.Contains(t, configmap, "name: {{ .Release.Name }}-control-plane-config")

	hpa := fmt.Sprintf(HPAYAMLFmt, name, name)
	assert.Contains(t, hpa, "name: {{ .Release.Name }}-control-plane")
	assert.Contains(t, hpa, "averageUtilization: 70")
}

package k8s

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestHelmChartGenerator_GenerateCharts_Basic(t *testing.T) {
	dir := t.TempDir()
	gen := &HelmChartGenerator{}

	cfg := ChartConfig{
		ReleaseName: "kapok",
		Namespace:   "kapok",
		Cloud:       CloudGeneric,
		Domain:      "kapok.local",
		TLSEnabled:  false,
		HPAEnabled:  false,
		KEDAEnabled: false,
		ImageTag:    "v1.0.0",
	}

	err := gen.GenerateCharts(dir, cfg)
	require.NoError(t, err)

	root := filepath.Join(dir, "kapok-platform")

	// Verify umbrella chart files
	assert.FileExists(t, filepath.Join(root, "Chart.yaml"))
	assert.FileExists(t, filepath.Join(root, "values.yaml"))
	assert.FileExists(t, filepath.Join(root, "templates", "namespace.yaml"))
	assert.FileExists(t, filepath.Join(root, "templates", "secrets.yaml"))

	// TLS files should not exist
	assert.NoFileExists(t, filepath.Join(root, "templates", "cert-manager.yaml"))
	assert.NoFileExists(t, filepath.Join(root, "templates", "cluster-issuer.yaml"))

	// KEDA should not exist
	assert.NoFileExists(t, filepath.Join(root, "templates", "keda-scaled-object.yaml"))

	// Verify subcharts
	for _, sc := range []string{"control-plane", "graphql-engine", "provisioner"} {
		scDir := filepath.Join(root, "charts", sc)
		assert.FileExists(t, filepath.Join(scDir, "Chart.yaml"))
		assert.FileExists(t, filepath.Join(scDir, "templates", "deployment.yaml"))
		assert.FileExists(t, filepath.Join(scDir, "templates", "service.yaml"))
		assert.FileExists(t, filepath.Join(scDir, "templates", "ingress.yaml"))
		assert.FileExists(t, filepath.Join(scDir, "templates", "configmap.yaml"))
		// HPA should not exist
		assert.NoFileExists(t, filepath.Join(scDir, "templates", "hpa.yaml"))
	}
}

func TestHelmChartGenerator_WithAllFeatures(t *testing.T) {
	dir := t.TempDir()
	gen := &HelmChartGenerator{}

	cfg := ChartConfig{
		ReleaseName: "kapok",
		Namespace:   "production",
		Cloud:       CloudAWS,
		Domain:      "app.example.com",
		TLSEnabled:  true,
		HPAEnabled:  true,
		KEDAEnabled: true,
		ImageTag:    "v2.0.0",
	}

	err := gen.GenerateCharts(dir, cfg)
	require.NoError(t, err)

	root := filepath.Join(dir, "kapok-platform")

	// TLS files should exist
	assert.FileExists(t, filepath.Join(root, "templates", "cert-manager.yaml"))
	assert.FileExists(t, filepath.Join(root, "templates", "cluster-issuer.yaml"))

	// KEDA should exist
	assert.FileExists(t, filepath.Join(root, "templates", "keda-scaled-object.yaml"))

	// HPA should exist in subcharts
	for _, sc := range []string{"control-plane", "graphql-engine", "provisioner"} {
		assert.FileExists(t, filepath.Join(root, "charts", sc, "templates", "hpa.yaml"))
	}
}

func TestHelmChartGenerator_ValuesYAMLParseable(t *testing.T) {
	dir := t.TempDir()
	gen := &HelmChartGenerator{}

	cfg := ChartConfig{
		ReleaseName: "kapok",
		Namespace:   "kapok",
		Cloud:       CloudAWS,
		Domain:      "test.example.com",
		TLSEnabled:  true,
		HPAEnabled:  true,
		KEDAEnabled: false,
		ImageTag:    "v1.0.0",
	}

	err := gen.GenerateCharts(dir, cfg)
	require.NoError(t, err)

	valuesPath := filepath.Join(dir, "kapok-platform", "values.yaml")
	data, err := os.ReadFile(valuesPath)
	require.NoError(t, err)

	var values map[string]interface{}
	err = yaml.Unmarshal(data, &values)
	require.NoError(t, err)

	global := values["global"].(map[string]interface{})
	assert.Equal(t, "aws", global["cloud"])
	assert.Equal(t, "test.example.com", global["domain"])
	assert.Equal(t, true, global["tls"].(map[string]interface{})["enabled"])
}

func TestHelmChartGenerator_CloudSpecificValues(t *testing.T) {
	dir := t.TempDir()
	gen := &HelmChartGenerator{}

	cfg := ChartConfig{
		ReleaseName: "kapok",
		Namespace:   "kapok",
		Cloud:       CloudGCP,
		Domain:      "kapok.local",
		ImageTag:    "latest",
	}

	err := gen.GenerateCharts(dir, cfg)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "kapok-platform", "values.yaml"))
	require.NoError(t, err)

	var values map[string]interface{}
	require.NoError(t, yaml.Unmarshal(data, &values))

	global := values["global"].(map[string]interface{})
	assert.Equal(t, "standard-rwo", global["storageClass"])
	assert.Equal(t, "gce", global["ingressClass"])
}

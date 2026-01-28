package k8s

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// ChartConfig holds all configuration for Helm chart generation.
type ChartConfig struct {
	ReleaseName  string
	Namespace    string
	Cloud        CloudProvider
	Domain       string
	TLSEnabled   bool
	HPAEnabled   bool
	KEDAEnabled  bool
	ImageTag     string
	StorageClass string
	IngressClass string
}

// subchartMeta describes a subchart to generate.
type subchartMeta struct {
	Name        string
	Description string
	PathPrefix  string
}

var subcharts = []subchartMeta{
	{Name: "control-plane", Description: "Kapok Control Plane", PathPrefix: "/api"},
	{Name: "graphql-engine", Description: "Kapok GraphQL Engine", PathPrefix: "/graphql"},
	{Name: "provisioner", Description: "Kapok Tenant Provisioner", PathPrefix: "/provision"},
}

// HelmChartGenerator generates a Helm umbrella chart to disk.
type HelmChartGenerator struct{}

// GenerateCharts writes the full chart tree to outputDir.
func (g *HelmChartGenerator) GenerateCharts(outputDir string, cfg ChartConfig) error {
	cloudCfg := CloudConfigFor(cfg.Cloud)
	if cfg.StorageClass == "" {
		cfg.StorageClass = cloudCfg.StorageClass
	}
	if cfg.IngressClass == "" {
		cfg.IngressClass = cloudCfg.IngressClass
	}

	root := filepath.Join(outputDir, "kapok-platform")

	// Umbrella chart
	if err := writeRaw(filepath.Join(root, "Chart.yaml"), ChartYAML); err != nil {
		return fmt.Errorf("failed to write Chart.yaml: %w", err)
	}
	if err := writeGoTemplate(filepath.Join(root, "values.yaml"), ValuesYAML, cfg); err != nil {
		return fmt.Errorf("failed to write values.yaml: %w", err)
	}

	// Umbrella templates (these are Helm templates, written verbatim)
	tmplDir := filepath.Join(root, "templates")
	if err := writeRaw(filepath.Join(tmplDir, "namespace.yaml"), NamespaceYAML); err != nil {
		return fmt.Errorf("failed to write namespace.yaml: %w", err)
	}
	if cfg.TLSEnabled {
		if err := writeRaw(filepath.Join(tmplDir, "cert-manager.yaml"), CertManagerYAML); err != nil {
			return fmt.Errorf("failed to write cert-manager.yaml: %w", err)
		}
		if err := writeRaw(filepath.Join(tmplDir, "cluster-issuer.yaml"), ClusterIssuerYAML); err != nil {
			return fmt.Errorf("failed to write cluster-issuer.yaml: %w", err)
		}
	}
	if cfg.KEDAEnabled {
		if err := writeRaw(filepath.Join(tmplDir, "keda-scaled-object.yaml"), KEDAScaledObjectYAML); err != nil {
			return fmt.Errorf("failed to write keda-scaled-object.yaml: %w", err)
		}
	}

	// Subcharts
	for _, sc := range subcharts {
		scDir := filepath.Join(root, "charts", sc.Name)
		scTmplDir := filepath.Join(scDir, "templates")

		if err := writeGoTemplate(filepath.Join(scDir, "Chart.yaml"), SubchartChartYAML, sc); err != nil {
			return fmt.Errorf("failed to write %s/Chart.yaml: %w", sc.Name, err)
		}

		n := sc.Name
		if err := writeRaw(filepath.Join(scTmplDir, "deployment.yaml"),
			fmt.Sprintf(DeploymentYAMLFmt, n, n, n, n, n, n)); err != nil {
			return fmt.Errorf("failed to write %s deployment: %w", n, err)
		}
		if err := writeRaw(filepath.Join(scTmplDir, "service.yaml"),
			fmt.Sprintf(ServiceYAMLFmt, n, n)); err != nil {
			return fmt.Errorf("failed to write %s service: %w", n, err)
		}
		if err := writeRaw(filepath.Join(scTmplDir, "ingress.yaml"),
			fmt.Sprintf(IngressYAMLFmt, n, n, sc.PathPrefix, n)); err != nil {
			return fmt.Errorf("failed to write %s ingress: %w", n, err)
		}
		if err := writeRaw(filepath.Join(scTmplDir, "configmap.yaml"),
			fmt.Sprintf(ConfigMapYAMLFmt, n)); err != nil {
			return fmt.Errorf("failed to write %s configmap: %w", n, err)
		}
		if cfg.HPAEnabled {
			if err := writeRaw(filepath.Join(scTmplDir, "hpa.yaml"),
				fmt.Sprintf(HPAYAMLFmt, n, n)); err != nil {
				return fmt.Errorf("failed to write %s hpa: %w", n, err)
			}
		}
	}

	return nil
}

// writeRaw writes content to path as-is.
func writeRaw(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// writeGoTemplate renders a Go text/template and writes the result.
func writeGoTemplate(path, tmplStr string, data interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	t, err := template.New("").Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

package deploy

import (
	"fmt"
	"os"

	"github.com/kapok/kapok/internal/k8s"
	"github.com/rs/zerolog/log"
)

// Options holds deploy command options.
type Options struct {
	Cloud     string
	Namespace string
	Domain    string
	TLS       bool
	HPA       bool
	KEDA      bool
	ImageTag  string
	OutputDir string
	DryRun    bool
	Timeout   string
}

// Deployer orchestrates detect → generate → install → monitor.
type Deployer struct {
	Detector  k8s.CloudDetector
	Generator *k8s.HelmChartGenerator
	Runner    CommandRunner
}

// Run executes the full deployment pipeline.
func (d *Deployer) Run(opts Options) error {
	// 1. Detect cloud provider
	cloud := k8s.CloudProvider(opts.Cloud)
	if cloud == "" {
		cloud = d.Detector.Detect()
	}
	log.Info().Str("cloud", string(cloud)).Msg("detected cloud provider")

	// 2. Build chart config
	cloudCfg := k8s.CloudConfigFor(cloud)
	chartCfg := k8s.ChartConfig{
		ReleaseName:  "kapok",
		Namespace:    opts.Namespace,
		Cloud:        cloud,
		Domain:       opts.Domain,
		TLSEnabled:   opts.TLS,
		HPAEnabled:   opts.HPA,
		KEDAEnabled:  opts.KEDA,
		ImageTag:     opts.ImageTag,
		StorageClass: cloudCfg.StorageClass,
		IngressClass: cloudCfg.IngressClass,
	}

	// 3. Generate charts
	outputDir := opts.OutputDir
	autoCreated := false
	if outputDir == "" {
		var err error
		outputDir, err = os.MkdirTemp("", "kapok-charts-*")
		if err != nil {
			return fmt.Errorf("failed to create temp dir: %w", err)
		}
		autoCreated = true
	}

	log.Info().Str("output_dir", outputDir).Msg("generating Helm charts")
	if err := d.Generator.GenerateCharts(outputDir, chartCfg); err != nil {
		return fmt.Errorf("failed to generate charts: %w", err)
	}

	if opts.DryRun {
		log.Info().Str("output_dir", outputDir).Msg("dry run complete, charts generated")
		return nil
	}

	// 4. Helm install/upgrade
	if autoCreated {
		defer os.RemoveAll(outputDir)
	}
	log.Info().Msg("deploying with Helm")
	chartPath := outputDir + "/kapok-platform"
	timeout := opts.Timeout
	if timeout == "" {
		timeout = "10m"
	}
	output, err := d.Runner.Run("helm", "upgrade", "--install",
		"kapok", chartPath,
		"--namespace", opts.Namespace,
		"--create-namespace",
		"--wait",
		"--timeout", timeout,
	)
	if err != nil {
		return fmt.Errorf("helm deploy failed: %w", err)
	}
	log.Info().Str("output", output).Msg("deployment complete")

	return nil
}

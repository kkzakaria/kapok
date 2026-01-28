package cmd

import (
	"fmt"

	"github.com/kapok/kapok/internal/deploy"
	"github.com/kapok/kapok/internal/k8s"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy Kapok to Kubernetes",
	Long: `Deploy Kapok to Kubernetes cluster with one command.

Auto-detects cloud provider and generates optimized Helm charts:
  • AWS EKS
  • GCP GKE
  • Azure AKS

Examples:
  kapok deploy                              # Auto-detect and deploy
  kapok deploy --cloud aws --domain app.io  # Deploy to AWS with custom domain
  kapok deploy --dry-run --output-dir ./out # Generate charts only
  kapok deploy --tls --keda                 # Enable TLS and KEDA autoscaling`,
	RunE: runDeploy,
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().String("cloud", "", "Cloud provider (aws|gcp|azure|generic)")
	deployCmd.Flags().String("namespace", "kapok", "Kubernetes namespace")
	deployCmd.Flags().String("domain", "kapok.local", "Domain for ingress")
	deployCmd.Flags().Bool("tls", false, "Enable TLS with cert-manager")
	deployCmd.Flags().Bool("hpa", true, "Enable Horizontal Pod Autoscaler")
	deployCmd.Flags().Bool("keda", false, "Enable KEDA event-driven autoscaling")
	deployCmd.Flags().String("image-tag", "latest", "Docker image tag")
	deployCmd.Flags().String("output-dir", "", "Output directory for generated charts")
	deployCmd.Flags().Bool("dry-run", false, "Generate charts without deploying")
	deployCmd.Flags().String("context", "", "Kubeconfig context name for cloud detection")
}

func runDeploy(cmd *cobra.Command, args []string) error {
	cloud, _ := cmd.Flags().GetString("cloud")
	namespace, _ := cmd.Flags().GetString("namespace")
	domain, _ := cmd.Flags().GetString("domain")
	tls, _ := cmd.Flags().GetBool("tls")
	hpa, _ := cmd.Flags().GetBool("hpa")
	keda, _ := cmd.Flags().GetBool("keda")
	imageTag, _ := cmd.Flags().GetString("image-tag")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	kubeContext, _ := cmd.Flags().GetString("context")

	deployer := &deploy.Deployer{
		Detector:  &k8s.KubeconfigDetector{ContextName: kubeContext},
		Generator: &k8s.HelmChartGenerator{},
		Runner:    &deploy.ExecRunner{},
	}

	opts := deploy.Options{
		Cloud:     cloud,
		Namespace: namespace,
		Domain:    domain,
		TLS:       tls,
		HPA:       hpa,
		KEDA:      keda,
		ImageTag:  imageTag,
		OutputDir: outputDir,
		DryRun:    dryRun,
	}

	if err := deployer.Run(opts); err != nil {
		return fmt.Errorf("deploy failed: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Deployment complete.")
	return nil
}

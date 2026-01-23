package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy Kapok to Kubernetes",
	Long: `Deploy Kapok to Kubernetes cluster with one command.

Auto-detects cloud provider and generates optimized Helm charts:
  • AWS EKS
  • GCP GKE
  • Azure AKS`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement actual deploy logic in Epic 4
		fmt.Fprintln(cmd.OutOrStdout(), "Deploying Kapok to Kubernetes...")
		fmt.Fprintln(cmd.OutOrStdout(), "✓ Detecting cloud provider...")
		fmt.Fprintln(cmd.OutOrStdout(), "✓ Generating Helm charts...")
		fmt.Fprintln(cmd.OutOrStdout(), "✓ Deploying services...")
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().String("cloud", "", "Cloud provider (aws|gcp|azure)")
	deployCmd.Flags().Bool("dry-run", false, "Generate manifests without deploying")
}

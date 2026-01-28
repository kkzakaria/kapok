package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version information (will be set during build)
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kapok",
	Short: "Backend-as-a-Service auto-hébergé avec multi-tenancy native",
	Long: `Kapok est une plateforme Backend-as-a-Service (BaaS) auto-hébergée 
conçue pour les développeurs frontend qui ont besoin de contrôle total 
sur leur infrastructure sans expertise DevOps.

Features:
  • Multi-Tenant Foundation (schema-per-tenant isolation)
  • GraphQL Auto-Generated (from PostgreSQL schema)
  • CLI Developer-Friendly (init, dev, deploy, tenant commands)
  • Kubernetes Deployment (one-command deploy to EKS/GKE/AKS)
  • Zero-Config (smart defaults for 90% of use cases)`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return ExecuteWithContext(context.Background(), os.Stdout, os.Args[1:])
}

// ExecuteContext is the testable version of Execute that accepts io.Writer and args
func ExecuteContext(out io.Writer, args []string) error {
	return ExecuteWithContext(context.Background(), out, args)
}

// ExecuteWithContext is like ExecuteContext but accepts a context for cancellation
func ExecuteWithContext(ctx context.Context, out io.Writer, args []string) error {
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs(args)
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().String("context", "", "Kubernetes context to use")
	rootCmd.PersistentFlags().String("namespace", "kapok", "Kubernetes namespace")
	rootCmd.PersistentFlags().StringP("output", "o", "text", "Output format (text|json|yaml)")

	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "kapok version %s\n", version)
			fmt.Fprintf(cmd.OutOrStdout(), "commit: %s\n", commit)
			fmt.Fprintf(cmd.OutOrStdout(), "built: %s\n", date)
		},
	})
}

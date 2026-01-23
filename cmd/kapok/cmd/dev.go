package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start local development environment",
	Long: `Start local development environment with PostgreSQL and GraphQL Playground.

Launches:
  • PostgreSQL container (via Docker)
  • GraphQL engine in development mode
  • Auto-reload on code changes`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement actual dev logic in Story 1.6
		fmt.Fprintln(cmd.OutOrStdout(), "Starting Kapok development environment...")
		fmt.Fprintln(cmd.OutOrStdout(), "✓ PostgreSQL starting on localhost:5432")
		fmt.Fprintln(cmd.OutOrStdout(), "✓ GraphQL Playground available at http://localhost:8080/playground")
	},
}

func init() {
	rootCmd.AddCommand(devCmd)

	devCmd.Flags().Int("port", 8080, "GraphQL server port")
	devCmd.Flags().Bool("no-postgres", false, "Skip PostgreSQL container")
}

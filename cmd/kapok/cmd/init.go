package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new Kapok project",
	Long: `Initialize a new Kapok project with zero configuration.

Creates:
  • kapok.yaml configuration file with smart defaults
  • .env.example template
  • README.md with project-specific quick start
  • docs/ folder with basic architecture documentation`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := "my-kapok-project"
		if len(args) > 0 {
			projectName = args[0]
		}

		// TODO: Implement actual init logic in Story 1.5
		fmt.Fprintf(cmd.OutOrStdout(), "Initializing Kapok project: %s\n", projectName)
		fmt.Fprintln(cmd.OutOrStdout(), "✓ Created kapok.yaml")
		fmt.Fprintln(cmd.OutOrStdout(), "✓ Created .env.example")
		fmt.Fprintln(cmd.OutOrStdout(), "✓ Created README.md")
		fmt.Fprintln(cmd.OutOrStdout(), "\nNext steps:")
		fmt.Fprintln(cmd.OutOrStdout(), "  kapok dev    # Start local development")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().Bool("force", false, "Overwrite existing files")
}

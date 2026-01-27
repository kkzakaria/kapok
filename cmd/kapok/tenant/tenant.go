package tenant

import (
	"github.com/spf13/cobra"
)

// NewTenantCommand creates the tenant root command
func NewTenantCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tenant",
		Short: "Manage tenants",
		Long:  "Commands to create, list, and delete tenants in the Kapok platform",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(NewCreateCommand())
	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewDeleteCommand())

	return cmd
}

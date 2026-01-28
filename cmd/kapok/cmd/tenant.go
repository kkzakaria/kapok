package cmd

import (
	tenantcmd "github.com/kapok/kapok/cmd/kapok/tenant"
)

func init() {
	// Add the tenant command from the tenant package
	rootCmd.AddCommand(tenantcmd.NewTenantCommand())
}

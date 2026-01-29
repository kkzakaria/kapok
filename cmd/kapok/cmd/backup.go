package cmd

import (
	backupcmd "github.com/kapok/kapok/cmd/kapok/backup"
)

func init() {
	rootCmd.AddCommand(backupcmd.NewBackupCommand())
}

package backup

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the backup delete command.
func NewDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete BACKUP_ID",
		Short: "Delete a backup",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Printf("Are you sure you want to delete backup %s? Use --force to confirm.\n", args[0])
				return nil
			}
			return runDelete(args[0])
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation")
	return cmd
}

func runDelete(backupID string) error {
	ctx := context.Background()

	svc, db, err := newBackupService(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := svc.DeleteBackup(ctx, backupID); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	fmt.Printf("Backup %s deleted.\n", backupID)
	return nil
}

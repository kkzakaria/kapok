package backup

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// NewRestoreCommand creates the backup restore command.
func NewRestoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore BACKUP_ID",
		Short: "Restore a backup",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRestore(args[0])
		},
	}
	return cmd
}

func runRestore(backupID string) error {
	ctx := context.Background()

	svc, db, err := newBackupService(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := svc.RestoreBackup(ctx, backupID); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	fmt.Printf("\nRestore completed successfully for backup %s\n", backupID)
	return nil
}

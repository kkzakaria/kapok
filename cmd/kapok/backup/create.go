package backup

import (
	"context"
	"fmt"

	bk "github.com/kapok/kapok/internal/backup"
	"github.com/spf13/cobra"
)

// NewCreateCommand creates the backup create command.
func NewCreateCommand() *cobra.Command {
	var tenantID string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a backup for a tenant",
		RunE: func(cmd *cobra.Command, args []string) error {
			if tenantID == "" {
				return fmt.Errorf("--tenant-id is required")
			}
			return runCreate(tenantID)
		},
	}

	cmd.Flags().StringVar(&tenantID, "tenant-id", "", "Tenant ID to backup")
	cmd.MarkFlagRequired("tenant-id")
	return cmd
}

func runCreate(tenantID string) error {
	ctx := context.Background()

	svc, db, err := newBackupService(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	// Look up schema name
	var schemaName string
	if err := db.QueryRowContext(ctx, `SELECT schema_name FROM tenants WHERE id = $1`, tenantID).Scan(&schemaName); err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	b, err := svc.CreateBackup(ctx, tenantID, schemaName, bk.TriggerManual)
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	fmt.Printf("\nBackup triggered successfully!\n\n")
	fmt.Printf("  ID:     %s\n", b.ID)
	fmt.Printf("  Status: %s\n", b.Status)
	fmt.Printf("  Path:   %s\n\n", b.StoragePath)
	return nil
}

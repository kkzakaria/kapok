package backup

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"

	bk "github.com/kapok/kapok/internal/backup"
	"github.com/kapok/kapok/internal/backup/storage"
	"github.com/kapok/kapok/internal/database"
	"github.com/rs/zerolog"
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
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	ctx := context.Background()

	dbConfig, err := loadDBConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	db, err := database.NewDB(ctx, dbConfig, logger)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Look up schema name
	var schemaName string
	if err := db.QueryRowContext(ctx, `SELECT schema_name FROM tenants WHERE id = $1`, tenantID).Scan(&schemaName); err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	storagePath := os.Getenv("KAPOK_BACKUP_STORAGE_PATH")
	if storagePath == "" {
		storagePath = "./backups"
	}

	store, err := storage.NewFilesystemStore(storagePath)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	var encKey []byte
	if keyHex := os.Getenv("KAPOK_BACKUP_ENCRYPTION_KEY"); keyHex != "" {
		encKey, err = hex.DecodeString(keyHex)
		if err != nil || len(encKey) != 32 {
			return fmt.Errorf("KAPOK_BACKUP_ENCRYPTION_KEY must be 64 hex chars (32 bytes)")
		}
	}

	svc := bk.NewService(db, store, encKey, 30, logger)
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

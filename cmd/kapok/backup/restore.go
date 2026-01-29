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
	if err := svc.RestoreBackup(ctx, backupID); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	fmt.Printf("\nRestore completed successfully for backup %s\n", backupID)
	return nil
}

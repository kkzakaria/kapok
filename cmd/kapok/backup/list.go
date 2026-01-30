package backup

import (
	"context"
	"fmt"
	"os"

	bk "github.com/kapok/kapok/internal/backup"
	"github.com/kapok/kapok/internal/database"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

// NewListCommand creates the backup list command.
func NewListCommand() *cobra.Command {
	var tenantID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List backups",
		RunE: func(cmd *cobra.Command, args []string) error {
			if tenantID == "" {
				return fmt.Errorf("--tenant-id is required")
			}
			return runList(tenantID)
		},
	}

	cmd.Flags().StringVar(&tenantID, "tenant-id", "", "Tenant ID to list backups for")
	cmd.MarkFlagRequired("tenant-id")
	return cmd
}

func runList(tenantID string) error {
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

	repo := bk.NewRepository(db)
	backups, err := repo.ListByTenant(ctx, tenantID, 100, 0)
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	if len(backups) == 0 {
		fmt.Println("No backups found.")
		return nil
	}

	fmt.Printf("%-36s  %-10s  %-10s  %-12s  %s\n", "ID", "STATUS", "TYPE", "SIZE", "CREATED")
	fmt.Println("------------------------------------  ----------  ----------  ------------  -------------------")
	for _, b := range backups {
		fmt.Printf("%-36s  %-10s  %-10s  %12d  %s\n",
			b.ID, b.Status, b.Type, b.SizeBytes, b.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

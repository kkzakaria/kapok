package tenant

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kapok/kapok/internal/database"
	"github.com/kapok/kapok/internal/tenant"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	force      bool
	hardDelete bool
)

// NewDeleteCommand creates the tenant delete command
func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete TENANT_ID",
		Short: "Delete a tenant",
		Long:  "Deletes a tenant by ID. By default performs a soft delete (status = deleted). Use --hard to permanently delete the tenant and its schema.",
		Args:  cobra.ExactArgs(1),
		RunE:  runDelete,
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&hardDelete, "hard", false, "Permanently delete tenant and drop schema (WARNING: irreversible)")

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) error {
	tenantID := args[0]

	// Initialize logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Get database configuration
	dbConfig, err := loadDBConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Connect to database
	ctx := context.Background()
	db, err := database.NewDB(ctx, dbConfig, logger)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Create provisioner
	provisioner := tenant.NewProvisioner(db, logger)

	// Get tenant details
	existingTenant, err := provisioner.GetTenantByID(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to find tenant: %w", err)
	}

	// Confirmation prompt (unless --force)
	if !force {
		if hardDelete {
			fmt.Printf("\n⚠️  WARNING: Hard delete will permanently delete tenant '%s' and DROP its schema!\n", existingTenant.Name)
			fmt.Printf("   This action is IRREVERSIBLE and all data will be lost.\n\n")
		} else {
			fmt.Printf("\nAbout to soft delete tenant '%s' (ID: %s)\n", existingTenant.Name, tenantID)
			fmt.Printf("The tenant will be marked as deleted but data will be preserved.\n\n")
		}

		fmt.Print("Are you sure? (yes/no): ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" && response != "y" {
			fmt.Println("❌ Delete cancelled")
			return nil
		}
	}

	// Perform delete
	if hardDelete {
		logger.Warn().
			Str("tenant_id", tenantID).
			Str("tenant_name", existingTenant.Name).
			Msg("performing hard delete")

		if err := provisioner.HardDeleteTenant(ctx, tenantID); err != nil {
			return fmt.Errorf("failed to hard delete tenant: %w", err)
		}

		fmt.Printf("\n✅ Tenant '%s' permanently deleted (schema dropped)\n\n", existingTenant.Name)
	} else {
		logger.Info().
			Str("tenant_id", tenantID).
			Str("tenant_name", existingTenant.Name).
			Msg("performing soft delete")

		if err := provisioner.DeleteTenant(ctx, tenantID); err != nil {
			return fmt.Errorf("failed to delete tenant: %w", err)
		}

		fmt.Printf("\n✅ Tenant '%s' soft deleted (status = deleted)\n", existingTenant.Name)
		fmt.Printf("   Schema '%s' preserved. Use --hard to permanently delete.\n\n", existingTenant.SchemaName)
	}

	return nil
}

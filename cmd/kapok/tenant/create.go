package tenant

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kapok/kapok/internal/database"
	"github.com/kapok/kapok/internal/tenant"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

// NewCreateCommand creates the tenant create command
func NewCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a new tenant",
		Long:  "Creates a new tenant with the specified name and provisions its database schema",
		Args:  cobra.ExactArgs(1),
		RunE:  runCreate,
	}

	return cmd
}

func runCreate(cmd *cobra.Command, args []string) error {
	tenantName := args[0]

	// Initialize logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Validate tenant name
	if err := tenant.ValidateName(tenantName); err != nil {
		return fmt.Errorf("invalid tenant name: %w", err)
	}

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

	// Create tenant
	logger.Info().Str("name", tenantName).Msg("creating tenant")
	start := time.Now()

	newTenant, err := provisioner.CreateTenant(ctx, tenantName)
	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	duration := time.Since(start)

	// Display success message
	fmt.Printf("\n✅ Tenant created successfully!\n\n")
	fmt.Printf("  ID:          %s\n", newTenant.ID)
	fmt.Printf("  Name:        %s\n", newTenant.Name)
	fmt.Printf("  Schema:      %s\n", newTenant.SchemaName)
	fmt.Printf("  Status:      %s\n", newTenant.Status)
	fmt.Printf("  Created:     %s\n", newTenant.CreatedAt.Format(time.RFC3339))
	fmt.Printf("  Duration:    %s\n\n", duration.Round(time.Millisecond))

	if duration > 30*time.Second {
		logger.Warn().
			Dur("duration", duration).
			Msg("tenant provisioning exceeded 30 second target")
	}

	return nil
}

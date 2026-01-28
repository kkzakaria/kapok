package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kapok/kapok/internal/database"
	"github.com/kapok/kapok/internal/tenant"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	outputFormat string
	limit        int
	offset       int
	status       string
)

// NewListCommand creates the tenant list command
func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all tenants",
		Long:  "Lists all tenants in the system with optional filtering and pagination",
		RunE:  runList,
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table or json")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of tenants to display")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of tenants to skip")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status (active, provisioning, suspended, deleted)")

	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	// Initialize logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Get database configuration
	dbConfig := database.Config{
		Host:     viper.GetString("database.host"),
		Port:     viper.GetInt("database.port"),
		Database: viper.GetString("database.name"),
		User:     viper.GetString("database.user"),
		Password: viper.GetString("database.password"),
		SSLMode:  viper.GetString("database.sslmode"),
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

	// List tenants
	var tenantStatus tenant.TenantStatus
	if status != "" {
		tenantStatus = tenant.TenantStatus(status)
	}

	tenants, err := provisioner.ListTenants(ctx, tenantStatus, limit, offset)
	if err != nil {
		return fmt.Errorf("failed to list tenants: %w", err)
	}

	// Display output
	switch outputFormat {
	case "json":
		return displayJSON(tenants)
	case "table":
		return displayTable(tenants)
	default:
		return fmt.Errorf("invalid output format: %s (use 'table' or 'json')", outputFormat)
	}
}

func displayTable(tenants []*tenant.Tenant) error {
	if len(tenants) == 0 {
		fmt.Println("No tenants found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()

	// Header
	fmt.Fprintln(w, "ID\tNAME\tSCHEMA\tSTATUS\tCREATED")
	fmt.Fprintln(w, "──\t────\t──────\t──────\t───────")

	// Rows
	for _, t := range tenants {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			t.ID[:8]+"...", // Show first 8 chars of UUID
			t.Name,
			t.SchemaName,
			t.Status,
			t.CreatedAt.Format("2006-01-02 15:04"),
		)
	}

	fmt.Fprintf(w, "\nTotal: %d tenant(s)\n", len(tenants))

	return nil
}

func displayJSON(tenants []*tenant.Tenant) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(tenants)
}

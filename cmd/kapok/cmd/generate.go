package cmd

import (
	"database/sql"
	"fmt"

	"github.com/kapok/kapok/pkg/config"
	"github.com/kapok/kapok/pkg/codegen"
	"github.com/kapok/kapok/pkg/codegen/typescript"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	generateOutputDir string
	generateSchema    string
	generateProjectName string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate SDK and other artifacts",
	Long: `Generate SDK clients, types, and other artifacts from your database schema.
	
Example:
  kapok generate sdk                         # Generate TypeScript SDK
  kapok generate sdk --output-dir ./client   # Generate to custom directory
  kapok generate sdk --schema tenant_123    # Generate for specific schema`,
}

// sdkCmd represents the sdk subcommand
var sdkCmd = &cobra.Command{
	Use:   "sdk",
	Short: "Generate TypeScript SDK from database schema",
	Long: `Generate a type-safe TypeScript SDK from your PostgreSQL database schema.

The SDK includes:
  - TypeScript interfaces for all tables
  - CRUD functions for each entity
  - Main KapokClient class
  - package.json and tsconfig.json

Example:
  kapok generate sdk
  kapok generate sdk --output-dir ./client/sdk
  kapok generate sdk --schema public --project-name my-api-client`,
	RunE: runGenerateSDK,
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.AddCommand(sdkCmd)

	// SDK command flags
	sdkCmd.Flags().StringVar(&generateOutputDir, "output-dir", "./sdk/typescript", "Output directory for generated SDK")
	sdkCmd.Flags().StringVarP(&generateSchema, "schema", "s", "public", "PostgreSQL schema name")
	sdkCmd.Flags().StringVar(&generateProjectName, "project-name", "kapok-sdk", "NPM package name for the SDK")
}

func runGenerateSDK(cmd *cobra.Command, args []string) error {
	log.Info().Msg("Starting SDK generation...")

	// Load configuration with defaults
	cfg := config.Defaults()
	
	// Allow override from environment variables
	// In production, you would use a proper config loader here

	// Connect to database
	db, err := connectToDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Verify connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().
		Str("schema", generateSchema).
		Str("output_dir", generateOutputDir).
		Msg("Introspecting database schema...")

	// Introspect schema
	introspector := codegen.NewSchemaIntrospector(db)
	schema, err := introspector.IntrospectSchema(generateSchema)
	if err != nil {
		return fmt.Errorf("failed to introspect schema: %w", err)
	}

	if len(schema.Tables) == 0 {
		log.Warn().
			Str("schema", generateSchema).
			Msg("No tables found in schema")
		return fmt.Errorf("no tables found in schema '%s'", generateSchema)
	}

	log.Info().
		Int("table_count", len(schema.Tables)).
		Msg("Schema introspection complete")

	// Generate SDK
	log.Info().Msg("Generating TypeScript SDK...")
	
	clientGen := typescript.NewClientGenerator()
	if err := clientGen.WriteSDK(generateOutputDir, schema, generateProjectName); err != nil {
		return fmt.Errorf("failed to write SDK: %w", err)
	}

	log.Info().
		Str("output_dir", generateOutputDir).
		Int("tables", len(schema.Tables)).
		Msg("✓ SDK generation complete!")

	// Print next steps
	fmt.Println("\n📦 Next steps:")
	fmt.Printf("  cd %s\n", generateOutputDir)
	fmt.Println("  npm install")
	fmt.Println("  npm run build")
	fmt.Println("\n💡 Import the SDK in your app:")
	fmt.Printf("  import { KapokClient } from '%s';\n", generateProjectName)

	return nil
}

// connectToDatabase creates a database connection from config
func connectToDatabase(cfg *config.Config) (*sql.DB, error) {
	// Build connection string
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Database,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	return db, nil
}

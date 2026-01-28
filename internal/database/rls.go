package database

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
)

// RLSManager handles Row-Level Security policy management
type RLSManager struct {
	db     *DB
	logger zerolog.Logger
}

// NewRLSManager creates a new RLS manager
func NewRLSManager(db *DB, logger zerolog.Logger) *RLSManager {
	return &RLSManager{
		db:     db,
		logger: logger,
	}
}

// EnableRLSForTable enables Row-Level Security on a table
func (r *RLSManager) EnableRLSForTable(ctx context.Context, schemaName, tableName string) error {
	r.logger.Info().
		Str("schema", schemaName).
		Str("table", tableName).
		Msg("enabling RLS on table")

	// Validate inputs to prevent SQL injection
	if !isValidSchemaName(schemaName) {
		return fmt.Errorf("invalid schema name: %s", schemaName)
	}
	if !isValidTableName(tableName) {
		return fmt.Errorf("invalid table name: %s", tableName)
	}

	// Enable RLS
	query := fmt.Sprintf("ALTER TABLE %s.%s ENABLE ROW LEVEL SECURITY", schemaName, tableName)
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to enable RLS on %s.%s: %w", schemaName, tableName, err)
	}

	return nil
}

// CreateTenantIsolationPolicy creates a RLS policy for tenant isolation
func (r *RLSManager) CreateTenantIsolationPolicy(ctx context.Context, schemaName, tableName string) error {
	r.logger.Info().
		Str("schema", schemaName).
		Str("table", tableName).
		Msg("creating tenant isolation policy")

	// Validate inputs
	if !isValidSchemaName(schemaName) {
		return fmt.Errorf("invalid schema name: %s", schemaName)
	}
	if !isValidTableName(tableName) {
		return fmt.Errorf("invalid table name: %s", tableName)
	}

	// Create policy that enforces tenant_id match
	policyName := fmt.Sprintf("tenant_isolation_%s", tableName)
	query := fmt.Sprintf(`
		CREATE POLICY IF NOT EXISTS %s ON %s.%s
		USING (tenant_id = current_setting('app.tenant_id', true)::uuid)
	`, policyName, schemaName, tableName)

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create RLS policy on %s.%s: %w", schemaName, tableName, err)
	}

	r.logger.Info().
		Str("policy", policyName).
		Str("schema", schemaName).
		Str("table", tableName).
		Msg("tenant isolation policy created")

	return nil
}

// ApplyRLSPolicies applies RLS policies to all tables in a schema
func (r *RLSManager) ApplyRLSPolicies(ctx context.Context, schemaName string) error {
	r.logger.Info().
		Str("schema", schemaName).
		Msg("applying RLS policies to all tables in schema")

	// Validate schema name
	if !isValidSchemaName(schemaName) {
		return fmt.Errorf("invalid schema name: %s", schemaName)
	}

	// Get all tables in the schema
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = $1 
		AND table_type = 'BASE TABLE'
	`
	rows, err := r.db.QueryContext(ctx, query, schemaName)
	if err != nil {
		return fmt.Errorf("failed to get tables in schema %s: %w", schemaName, err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating tables: %w", err)
	}

	// Apply RLS to each table
	for _, tableName := range tables {
		// Check if table has tenant_id column
		hasTenantID, err := r.tableHasTenantIDColumn(ctx, schemaName, tableName)
		if err != nil {
			r.logger.Warn().
				Err(err).
				Str("table", tableName).
				Msg("failed to check if table has tenant_id column, skipping")
			continue
		}

		if !hasTenantID {
			r.logger.Debug().
				Str("table", tableName).
				Msg("table does not have tenant_id column, skipping RLS")
			continue
		}

		// Enable RLS
		if err := r.EnableRLSForTable(ctx, schemaName, tableName); err != nil {
			return fmt.Errorf("failed to enable RLS on %s.%s: %w", schemaName, tableName, err)
		}

		// Create tenant isolation policy
		if err := r.CreateTenantIsolationPolicy(ctx, schemaName, tableName); err != nil {
			return fmt.Errorf("failed to create policy on %s.%s: %w", schemaName, tableName, err)
		}
	}

	r.logger.Info().
		Str("schema", schemaName).
		Int("tables", len(tables)).
		Msg("RLS policies applied successfully")

	return nil
}

// tableHasTenantIDColumn checks if a table has a tenant_id column
func (r *RLSManager) tableHasTenantIDColumn(ctx context.Context, schemaName, tableName string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_schema = $1 
			AND table_name = $2 
			AND column_name = 'tenant_id'
		)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, schemaName, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check tenant_id column: %w", err)
	}
	return exists, nil
}

// VerifyRLSEnabled verifies that RLS is enabled on all tables with tenant_id
func (r *RLSManager) VerifyRLSEnabled(ctx context.Context, schemaName string) ([]string, error) {
	r.logger.Info().
		Str("schema", schemaName).
		Msg("verifying RLS is enabled on all tenant tables")

	// Query to find tables with tenant_id but without RLS
	query := `
		SELECT t.tablename
		FROM pg_tables t
		LEFT JOIN pg_class c ON t.tablename = c.relname
		WHERE t.schemaname = $1
		AND EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = t.schemaname
			AND table_name = t.tablename
			AND column_name = 'tenant_id'
		)
		AND (c.relrowsecurity = false OR c.relrowsecurity IS NULL)
	`

	rows, err := r.db.QueryContext(ctx, query, schemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to verify RLS: %w", err)
	}
	defer rows.Close()

	var tablesWithoutRLS []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tablesWithoutRLS = append(tablesWithoutRLS, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	if len(tablesWithoutRLS) > 0 {
		r.logger.Warn().
			Strs("tables", tablesWithoutRLS).
			Msg("tables with tenant_id but without RLS enabled")
	} else {
		r.logger.Info().
			Str("schema", schemaName).
			Msg("all tenant tables have RLS enabled")
	}

	return tablesWithoutRLS, nil
}

// isValidTableName validates table name to prevent SQL injection
func isValidTableName(tableName string) bool {
	// Table name should only contain alphanumeric, underscore
	for _, c := range tableName {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return len(tableName) > 0 && len(tableName) <= 63 // PostgreSQL max identifier length
}

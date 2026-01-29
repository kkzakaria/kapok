package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

// Migrator handles database migrations
type Migrator struct {
	db     *DB
	logger zerolog.Logger
}

// NewMigrator creates a new database migrator
func NewMigrator(db *DB, logger zerolog.Logger) *Migrator {
	return &Migrator{
		db:     db,
		logger: logger,
	}
}

// CreateControlDatabase creates the control database schema
func (m *Migrator) CreateControlDatabase(ctx context.Context) error {
	m.logger.Info().Msg("creating control database schema")

	// Begin transaction
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create tenants table
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS tenants (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(50) UNIQUE NOT NULL,
			schema_name VARCHAR(100) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tenants table: %w", err)
	}

	// Create index on status
	_, err = tx.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tenants status index: %w", err)
	}

	// Create casbin_rule table for RBAC
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS casbin_rule (
			id SERIAL PRIMARY KEY,
			ptype VARCHAR(10),
			v0 VARCHAR(256),
			v1 VARCHAR(256),
			v2 VARCHAR(256),
			v3 VARCHAR(256),
			v4 VARCHAR(256),
			v5 VARCHAR(256)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create casbin_rule table: %w", err)
	}

	// Create index on casbin_rule for performance
	_, err = tx.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_casbin_rule ON casbin_rule(ptype, v0, v1, v2, v3)
	`)
	if err != nil {
		return fmt.Errorf("failed to create casbin_rule index: %w", err)
	}

	// Create audit_log table
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS audit_log (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID,
			user_id VARCHAR(256),
			action VARCHAR(100) NOT NULL,
			resource VARCHAR(256),
			timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
			metadata JSONB
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create audit_log table: %w", err)
	}

	// Create indexes on audit_log
	_, err = tx.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_audit_log_tenant ON audit_log(tenant_id, timestamp)
	`)
	if err != nil {
		return fmt.Errorf("failed to create audit_log tenant index: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_audit_log_timestamp ON audit_log(timestamp)
	`)
	if err != nil {
		return fmt.Errorf("failed to create audit_log timestamp index: %w", err)
	}

	// Extend tenants table with control-plane columns
	tenantExtensions := []string{
		"ALTER TABLE tenants ADD COLUMN IF NOT EXISTS slug VARCHAR(100)",
		"ALTER TABLE tenants ADD COLUMN IF NOT EXISTS isolation_level VARCHAR(20) DEFAULT 'schema'",
		"ALTER TABLE tenants ADD COLUMN IF NOT EXISTS storage_used_bytes BIGINT DEFAULT 0",
		"ALTER TABLE tenants ADD COLUMN IF NOT EXISTS last_activity TIMESTAMP",
	}
	for _, stmt := range tenantExtensions {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to extend tenants table: %w", err)
		}
	}

	// Create users table for authentication
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(256) UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			roles TEXT NOT NULL DEFAULT 'user',
			tenant_id UUID,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create backups table
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS backups (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL REFERENCES tenants(id),
			schema_name VARCHAR(100) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			type VARCHAR(20) NOT NULL DEFAULT 'schema',
			trigger VARCHAR(20) NOT NULL DEFAULT 'manual',
			storage_path TEXT NOT NULL DEFAULT '',
			size_bytes BIGINT NOT NULL DEFAULT 0,
			checksum VARCHAR(64) NOT NULL DEFAULT '',
			encrypted BOOLEAN NOT NULL DEFAULT false,
			compressed BOOLEAN NOT NULL DEFAULT false,
			error_message TEXT NOT NULL DEFAULT '',
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			expires_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create backups table: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_backups_tenant ON backups(tenant_id, created_at DESC)
	`)
	if err != nil {
		return fmt.Errorf("failed to create backups tenant index: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_backups_expires ON backups(expires_at) WHERE expires_at IS NOT NULL
	`)
	if err != nil {
		return fmt.Errorf("failed to create backups expires index: %w", err)
	}

	// Create backup_schedules table
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS backup_schedules (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL REFERENCES tenants(id),
			cron_expr VARCHAR(100) NOT NULL,
			enabled BOOLEAN NOT NULL DEFAULT true,
			retention_days INT NOT NULL DEFAULT 30,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create backup_schedules table: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	m.logger.Info().Msg("control database schema created successfully")
	return nil
}

// CreateTenantSchema creates a new schema for a tenant
func (m *Migrator) CreateTenantSchema(ctx context.Context, schemaName string) error {
	m.logger.Info().Str("schema", schemaName).Msg("creating tenant schema")

	// Validate schema name (security: prevent SQL injection)
	if !isValidSchemaName(schemaName) {
		return fmt.Errorf("invalid schema name: %s", schemaName)
	}

	// Create schema
	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)
	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create schema %s: %w", schemaName, err)
	}

	m.logger.Info().Str("schema", schemaName).Msg("tenant schema created successfully")
	return nil
}

// DropTenantSchema drops a tenant schema (hard delete)
func (m *Migrator) DropTenantSchema(ctx context.Context, schemaName string) error {
	m.logger.Warn().Str("schema", schemaName).Msg("dropping tenant schema")

	// Validate schema name (security: prevent SQL injection)
	if !isValidSchemaName(schemaName) {
		return fmt.Errorf("invalid schema name: %s", schemaName)
	}

	// Drop schema
	query := fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)
	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop schema %s: %w", schemaName, err)
	}

	m.logger.Warn().Str("schema", schemaName).Msg("tenant schema dropped")
	return nil
}

// SchemaExists checks if a schema exists
func (m *Migrator) SchemaExists(ctx context.Context, schemaName string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM information_schema.schemata
			WHERE schema_name = $1
		)
	`
	err := m.db.QueryRowContext(ctx, query, schemaName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check schema existence: %w", err)
	}
	return exists, nil
}

// ExecuteMigration executes a migration SQL script within a transaction
func (m *Migrator) ExecuteMigration(ctx context.Context, migrationSQL string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin migration transaction: %w", err)
	}
	defer tx.Rollback()

	// Split migration into individual statements (simple approach)
	statements := splitSQLStatements(migrationSQL)

	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		m.logger.Debug().Int("statement", i+1).Msg("executing migration statement")
		_, err := tx.ExecContext(ctx, stmt)
		if err != nil {
			return fmt.Errorf("failed to execute migration statement %d: %w", i+1, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	m.logger.Info().Msg("migration executed successfully")
	return nil
}

// isValidSchemaName validates schema name to prevent SQL injection
func isValidSchemaName(schemaName string) bool {
	// Schema name must start with "tenant_" and contain only safe characters
	if !strings.HasPrefix(schemaName, "tenant_") {
		return false
	}
	// Check for only alphanumeric, underscore, and hyphen
	for _, c := range schemaName {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

// splitSQLStatements splits a SQL script into individual statements
func splitSQLStatements(sql string) []string {
	// Simple split by semicolon (doesn't handle all edge cases, but good enough for now)
	return strings.Split(sql, ";")
}

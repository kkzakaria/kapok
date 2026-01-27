package tenant

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kapok/kapok/internal/database"
	"github.com/rs/zerolog"
)

// Provisioner handles tenant provisioning operations
type Provisioner struct {
	db      *database.DB
	migrator *database.Migrator
	rls     *database.RLSManager
	logger  zerolog.Logger
}

// NewProvisioner creates a new tenant provisioner
func NewProvisioner(db *database.DB, logger zerolog.Logger) *Provisioner {
	return &Provisioner{
		db:      db,
		migrator: database.NewMigrator(db, logger),
		rls:     database.NewRLSManager(db, logger),
		logger:  logger,
	}
}

// CreateTenant provisions a new tenant with schema isolation
func (p *Provisioner) CreateTenant(ctx context.Context, name string) (*Tenant, error) {
	start := time.Now()
	
	p.logger.Info().
		Str("name", name).
		Msg("starting tenant provisioning")

	// Validate tenant name
	if err := ValidateName(name); err != nil {
		return nil, fmt.Errorf("invalid tenant name: %w", err)
	}

	// Generate tenant ID
	tenantID := uuid.New().String()
	schemaName := GenerateSchemaName(tenantID)

	// Create tenant object
	tenant := &Tenant{
		ID:         tenantID,
		Name:       name,
		SchemaName: schemaName,
		Status:     StatusProvisioning,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Begin transaction
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert tenant metadata
	query := `
		INSERT INTO tenants (id, name, schema_name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.ExecContext(ctx, query, 
		tenant.ID, 
		tenant.Name, 
		tenant.SchemaName, 
		tenant.Status,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert tenant metadata: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit tenant metadata: %w", err)
	}

	// Create tenant schema (outside transaction for DDL)
	if err := p.migrator.CreateTenantSchema(ctx, schemaName); err != nil {
		// Rollback: delete tenant metadata
		p.deleteTenantMetadata(ctx, tenantID)
		return nil, fmt.Errorf("failed to create tenant schema: %w", err)
	}

	// Update status to active
	tenant.Status = StatusActive
	if err := p.updateTenantStatus(ctx, tenantID, StatusActive); err != nil {
		p.logger.Warn().
			Err(err).
			Str("tenant_id", tenantID).
			Msg("failed to update tenant status to active")
	}

	duration := time.Since(start)
	p.logger.Info().
		Str("tenant_id", tenantID).
		Str("name", name).
		Dur("duration_ms", duration).
		Msg("tenant provisioned successfully")

	// Log to audit trail
	p.logAudit(ctx, tenantID, "tenant.create", fmt.Sprintf("tenant:%s", tenantID))

	return tenant, nil
}

// ListTenants retrieves all tenants with optional filtering
func (p *Provisioner) ListTenants(ctx context.Context, status TenantStatus, limit, offset int) ([]*Tenant, error) {
	p.logger.Debug().
		Str("status", string(status)).
		Int("limit", limit).
		Int("offset", offset).
		Msg("listing tenants")

	// Build query
	query := `
		SELECT id, name, schema_name, status, created_at, updated_at
		FROM tenants
	`
	args := []interface{}{}
	argIndex := 1

	// Add status filter if provided
	if status != "" {
		query += fmt.Sprintf(" WHERE status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	// Add ordering
	query += " ORDER BY created_at DESC"

	// Add pagination
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
		argIndex++
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, offset)
	}

	// Execute query
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tenants: %w", err)
	}
	defer rows.Close()

	// Scan results
	var tenants []*Tenant
	for rows.Next() {
		var tenant Tenant
		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.SchemaName,
			&tenant.Status,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenants = append(tenants, &tenant)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tenants: %w", err)
	}

	p.logger.Debug().
		Int("count", len(tenants)).
		Msg("tenants retrieved")

	return tenants, nil
}

// GetTenantByID retrieves a tenant by ID
func (p *Provisioner) GetTenantByID(ctx context.Context, id string) (*Tenant, error) {
	query := `
		SELECT id, name, schema_name, status, created_at, updated_at
		FROM tenants
		WHERE id = $1
	`

	var tenant Tenant
	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.SchemaName,
		&tenant.Status,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return &tenant, nil
}

// GetTenantByName retrieves a tenant by name
func (p *Provisioner) GetTenantByName(ctx context.Context, name string) (*Tenant, error) {
	query := `
		SELECT id, name, schema_name, status, created_at, updated_at
		FROM tenants
		WHERE name = $1
	`

	var tenant Tenant
	err := p.db.QueryRowContext(ctx, query, name).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.SchemaName,
		&tenant.Status,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found: %s", name)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return &tenant, nil
}

// DeleteTenant soft-deletes a tenant (preserves schema for recovery)
func (p *Provisioner) DeleteTenant(ctx context.Context, id string) error {
	p.logger.Info().
		Str("tenant_id", id).
		Msg("deleting tenant")

	// Verify tenant exists
	tenant, err := p.GetTenantByID(ctx, id)
	if err != nil {
		return err
	}

	// Update status to deleted (soft delete)
	if err := p.updateTenantStatus(ctx, id, StatusDeleted); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	p.logger.Info().
		Str("tenant_id", id).
		Str("name", tenant.Name).
		Msg("tenant deleted (soft delete)")

	// Log to audit trail
	p.logAudit(ctx, id, "tenant.delete", fmt.Sprintf("tenant:%s", id))

	return nil
}

// HardDeleteTenant permanently removes a tenant and drops its schema
func (p *Provisioner) HardDeleteTenant(ctx context.Context, id string) error {
	p.logger.Warn().
		Str("tenant_id", id).
		Msg("hard deleting tenant (permanent)")

	// Get tenant
	tenant, err := p.GetTenantByID(ctx, id)
	if err != nil {
		return err
	}

	// Drop schema
	if err := p.migrator.DropTenantSchema(ctx, tenant.SchemaName); err != nil {
		return fmt.Errorf("failed to drop tenant schema: %w", err)
	}

	// Delete metadata
	if err := p.deleteTenantMetadata(ctx, id); err != nil {
		return fmt.Errorf("failed to delete tenant metadata: %w", err)
	}

	p.logger.Warn().
		Str("tenant_id", id).
		Str("name", tenant.Name).
		Msg("tenant hard deleted")

	// Log to audit trail
	p.logAudit(ctx, id, "tenant.hard_delete", fmt.Sprintf("tenant:%s", id))

	return nil
}

// updateTenantStatus updates the status of a tenant
func (p *Provisioner) updateTenantStatus(ctx context.Context, id string, status TenantStatus) error {
	query := `
		UPDATE tenants
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := p.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update tenant status: %w", err)
	}
	return nil
}

// deleteTenantMetadata removes tenant metadata from the database
func (p *Provisioner) deleteTenantMetadata(ctx context.Context, id string) error {
	query := "DELETE FROM tenants WHERE id = $1"
	_, err := p.db.ExecContext(ctx, query, id)
	return err
}

// logAudit logs an audit event
func (p *Provisioner) logAudit(ctx context.Context, tenantID, action, resource string) {
	query := `
		INSERT INTO audit_log (tenant_id, action, resource, timestamp)
		VALUES ($1, $2, $3, $4)
	`
	_, err := p.db.ExecContext(ctx, query, tenantID, action, resource, time.Now())
	if err != nil {
		p.logger.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Str("action", action).
			Msg("failed to log audit event")
	}
}

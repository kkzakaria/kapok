package tenant

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kapok/kapok/internal/database"
	"github.com/rs/zerolog"
)

var nonAlphanumDash = regexp.MustCompile(`[^a-z0-9-]+`)
var multiDash = regexp.MustCompile(`-{2,}`)

// slugify converts a name to a URL-safe slug.
func slugify(name string) string {
	s := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	s = nonAlphanumDash.ReplaceAllString(s, "")
	s = multiDash.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// tenantSelectColumns is the common column list for tenant queries
const tenantSelectColumns = `id, name, schema_name, status,
	COALESCE(slug, ''), COALESCE(isolation_level, 'schema'),
	COALESCE(storage_used_bytes, 0), last_activity,
	parent_id, COALESCE(hierarchy_level, 'organization'), COALESCE(path, ''),
	COALESCE(database_name, ''), COALESCE(database_host, ''), COALESCE(database_port, 0),
	COALESCE(tier, 'standard'),
	created_at, updated_at`

// scanTenant scans a tenant row into a Tenant struct
func scanTenant(scanner interface{ Scan(dest ...interface{}) error }) (*Tenant, error) {
	var t Tenant
	err := scanner.Scan(
		&t.ID, &t.Name, &t.SchemaName, &t.Status,
		&t.Slug, &t.IsolationLevel,
		&t.StorageUsedBytes, &t.LastActivity,
		&t.ParentID, &t.HierarchyLevel, &t.Path,
		&t.DatabaseName, &t.DatabaseHost, &t.DatabasePort,
		&t.Tier,
		&t.CreatedAt, &t.UpdatedAt,
	)
	return &t, err
}

// Provisioner handles tenant provisioning operations
type Provisioner struct {
	db       *database.DB
	migrator *database.Migrator
	rls      *database.RLSManager
	logger   zerolog.Logger
}

// NewProvisioner creates a new tenant provisioner
func NewProvisioner(db *database.DB, logger zerolog.Logger) *Provisioner {
	return &Provisioner{
		db:       db,
		migrator: database.NewMigrator(db, logger),
		rls:      database.NewRLSManager(db, logger),
		logger:   logger,
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

	// Generate slug from name: lowercase, replace spaces with dashes, strip non-alphanumeric
	slug := slugify(name)

	// Create tenant object
	tenant := &Tenant{
		ID:             tenantID,
		Name:           name,
		SchemaName:     schemaName,
		Status:         StatusProvisioning,
		Slug:           slug,
		IsolationLevel: "schema",
		HierarchyLevel: HierarchyOrganization,
		Path:           "/" + tenantID,
		Tier:           "standard",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Begin transaction
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert tenant metadata
	query := `
		INSERT INTO tenants (id, name, schema_name, status, slug, isolation_level,
			hierarchy_level, path, tier, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = tx.ExecContext(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.SchemaName,
		tenant.Status,
		tenant.Slug,
		tenant.IsolationLevel,
		tenant.HierarchyLevel,
		tenant.Path,
		tenant.Tier,
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
		if rollbackErr := p.deleteTenantMetadata(ctx, tenantID); rollbackErr != nil {
			p.logger.Error().
				Err(rollbackErr).
				Str("tenant_id", tenantID).
				Msg("failed to rollback tenant metadata after schema creation failure")
		}
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

// CreateChildTenant provisions a new child tenant under the given parent
func (p *Provisioner) CreateChildTenant(ctx context.Context, parentID, name string) (*Tenant, error) {
	parent, err := p.GetTenantByID(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent tenant: %w", err)
	}

	// Determine child hierarchy level
	var childLevel HierarchyLevel
	switch parent.HierarchyLevel {
	case HierarchyOrganization:
		childLevel = HierarchyProject
	case HierarchyProject:
		childLevel = HierarchyTeam
	default:
		return nil, fmt.Errorf("cannot create child under %s level tenant (max depth %d)", parent.HierarchyLevel, MaxHierarchyDepth)
	}

	if err := ValidateName(name); err != nil {
		return nil, fmt.Errorf("invalid tenant name: %w", err)
	}

	tenantID := uuid.New().String()
	schemaName := GenerateSchemaName(tenantID)
	slug := slugify(name)
	path := parent.Path + "/" + tenantID

	tenant := &Tenant{
		ID:             tenantID,
		Name:           name,
		SchemaName:     schemaName,
		Status:         StatusProvisioning,
		Slug:           slug,
		IsolationLevel: parent.IsolationLevel,
		ParentID:       &parentID,
		HierarchyLevel: childLevel,
		Path:           path,
		Tier:           parent.Tier,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO tenants (id, name, schema_name, status, slug, isolation_level,
			parent_id, hierarchy_level, path, tier, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err = tx.ExecContext(ctx, query,
		tenant.ID, tenant.Name, tenant.SchemaName, tenant.Status,
		tenant.Slug, tenant.IsolationLevel,
		tenant.ParentID, tenant.HierarchyLevel, tenant.Path, tenant.Tier,
		tenant.CreatedAt, tenant.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert child tenant metadata: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit child tenant metadata: %w", err)
	}

	// Create tenant schema
	if err := p.migrator.CreateTenantSchema(ctx, schemaName); err != nil {
		if rollbackErr := p.deleteTenantMetadata(ctx, tenantID); rollbackErr != nil {
			p.logger.Error().Err(rollbackErr).Str("tenant_id", tenantID).
				Msg("failed to rollback child tenant metadata")
		}
		return nil, fmt.Errorf("failed to create child tenant schema: %w", err)
	}

	tenant.Status = StatusActive
	if err := p.updateTenantStatus(ctx, tenantID, StatusActive); err != nil {
		p.logger.Warn().Err(err).Str("tenant_id", tenantID).
			Msg("failed to update child tenant status to active")
	}

	p.logAudit(ctx, tenantID, "tenant.create_child", fmt.Sprintf("parent:%s,tenant:%s", parentID, tenantID))

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
	query := fmt.Sprintf("SELECT %s FROM tenants", tenantSelectColumns)
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
		t, err := scanTenant(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenants = append(tenants, t)
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
	query := fmt.Sprintf("SELECT %s FROM tenants WHERE id = $1", tenantSelectColumns)

	t, err := scanTenant(p.db.QueryRowContext(ctx, query, id))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return t, nil
}

// GetTenantByName retrieves a tenant by name
func (p *Provisioner) GetTenantByName(ctx context.Context, name string) (*Tenant, error) {
	query := fmt.Sprintf("SELECT %s FROM tenants WHERE name = $1", tenantSelectColumns)

	t, err := scanTenant(p.db.QueryRowContext(ctx, query, name))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found: %s", name)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return t, nil
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

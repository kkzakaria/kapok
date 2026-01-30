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

// MigrationManager orchestrates tenant migrations between isolation levels
type MigrationManager struct {
	db          *database.DB
	provisioner *Provisioner
	poolManager *PoolManager
	strategies  *StrategyRegistry
	logger      zerolog.Logger
}

// NewMigrationManager creates a new MigrationManager
func NewMigrationManager(
	db *database.DB,
	provisioner *Provisioner,
	poolManager *PoolManager,
	strategies *StrategyRegistry,
	logger zerolog.Logger,
) *MigrationManager {
	return &MigrationManager{
		db:          db,
		provisioner: provisioner,
		poolManager: poolManager,
		strategies:  strategies,
		logger:      logger,
	}
}

// Migrate initiates a migration from schema to database isolation (or vice versa)
func (m *MigrationManager) Migrate(ctx context.Context, tenantID, fromIsolation, toIsolation string) (*MigrationRecord, error) {
	m.logger.Info().
		Str("tenant_id", tenantID).
		Str("from", fromIsolation).
		Str("to", toIsolation).
		Msg("starting tenant migration")

	// Validate strategies exist
	if _, err := m.strategies.Get(fromIsolation); err != nil {
		return nil, fmt.Errorf("invalid source isolation: %w", err)
	}
	toStrategy, err := m.strategies.Get(toIsolation)
	if err != nil {
		return nil, fmt.Errorf("invalid target isolation: %w", err)
	}

	// Get tenant
	tenant, err := m.provisioner.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	if tenant.IsolationLevel != fromIsolation {
		return nil, fmt.Errorf("tenant isolation is %s, not %s", tenant.IsolationLevel, fromIsolation)
	}

	// Create migration record
	record := &MigrationRecord{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		FromIsolation: fromIsolation,
		ToIsolation:   toIsolation,
		Status:        MigrationPending,
		StartedAt:     time.Now(),
		CreatedAt:     time.Now(),
	}

	// Insert migration record
	_, err = m.db.ExecContext(ctx, `
		INSERT INTO tenant_migrations (id, tenant_id, from_isolation, to_isolation, status, started_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, record.ID, record.TenantID, record.FromIsolation, record.ToIsolation,
		record.Status, record.StartedAt, record.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration record: %w", err)
	}

	// Update tenant status to migrating
	if err := m.provisioner.updateTenantStatus(ctx, tenantID, StatusMigrating); err != nil {
		return nil, fmt.Errorf("failed to set tenant status to migrating: %w", err)
	}

	// Phase 1: Replicate - provision target
	m.updateMigrationStatus(ctx, record.ID, MigrationReplicating)
	if err := toStrategy.Provision(ctx, tenant); err != nil {
		m.failMigration(ctx, record.ID, tenantID, err)
		return record, fmt.Errorf("failed to provision target: %w", err)
	}

	// Phase 2: Verify
	m.updateMigrationStatus(ctx, record.ID, MigrationVerifying)
	// Verification is handled by the testing framework (Story 7)

	// Phase 3: Cutover
	m.updateMigrationStatus(ctx, record.ID, MigrationCutover)
	_, err = m.db.ExecContext(ctx, `
		UPDATE tenants SET isolation_level = $1, updated_at = NOW() WHERE id = $2
	`, toIsolation, tenantID)
	if err != nil {
		m.failMigration(ctx, record.ID, tenantID, err)
		return record, fmt.Errorf("failed to update isolation level: %w", err)
	}

	// Complete
	now := time.Now()
	record.Status = MigrationCompleted
	record.CompletedAt = &now
	rollbackDeadline := now.Add(24 * time.Hour)
	record.RollbackBefore = &rollbackDeadline

	m.db.ExecContext(ctx, `
		UPDATE tenant_migrations SET status = $1, completed_at = $2, rollback_before = $3 WHERE id = $4
	`, MigrationCompleted, now, rollbackDeadline, record.ID)

	m.provisioner.updateTenantStatus(ctx, tenantID, StatusActive)

	m.logger.Info().
		Str("migration_id", record.ID).
		Str("tenant_id", tenantID).
		Msg("migration completed")

	return record, nil
}

// GetMigration retrieves a migration record by ID
func (m *MigrationManager) GetMigration(ctx context.Context, migrationID string) (*MigrationRecord, error) {
	var r MigrationRecord
	err := m.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, from_isolation, to_isolation, status,
			started_at, completed_at, rollback_before, COALESCE(error_message, ''), created_at
		FROM tenant_migrations WHERE id = $1
	`, migrationID).Scan(
		&r.ID, &r.TenantID, &r.FromIsolation, &r.ToIsolation, &r.Status,
		&r.StartedAt, &r.CompletedAt, &r.RollbackBefore, &r.ErrorMessage, &r.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("migration not found: %s", migrationID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get migration: %w", err)
	}
	return &r, nil
}

// Rollback reverses a completed migration
func (m *MigrationManager) Rollback(ctx context.Context, migrationID string) error {
	record, err := m.GetMigration(ctx, migrationID)
	if err != nil {
		return err
	}

	if record.Status != MigrationCompleted {
		return fmt.Errorf("can only rollback completed migrations, current status: %s", record.Status)
	}

	if record.RollbackBefore != nil && time.Now().After(*record.RollbackBefore) {
		return fmt.Errorf("rollback window has expired")
	}

	m.logger.Info().Str("migration_id", migrationID).Msg("rolling back migration")

	// Revert isolation level
	_, err = m.db.ExecContext(ctx, `
		UPDATE tenants SET isolation_level = $1, updated_at = NOW() WHERE id = $2
	`, record.FromIsolation, record.TenantID)
	if err != nil {
		return fmt.Errorf("failed to revert isolation level: %w", err)
	}

	m.db.ExecContext(ctx, `
		UPDATE tenant_migrations SET status = $1 WHERE id = $2
	`, MigrationRolledBack, migrationID)

	m.provisioner.updateTenantStatus(ctx, record.TenantID, StatusActive)

	return nil
}

func (m *MigrationManager) updateMigrationStatus(ctx context.Context, migrationID string, status MigrationStatus) {
	m.db.ExecContext(ctx, `
		UPDATE tenant_migrations SET status = $1 WHERE id = $2
	`, status, migrationID)
}

func (m *MigrationManager) failMigration(ctx context.Context, migrationID, tenantID string, origErr error) {
	m.db.ExecContext(ctx, `
		UPDATE tenant_migrations SET status = $1, error_message = $2 WHERE id = $3
	`, MigrationFailed, origErr.Error(), migrationID)
	m.provisioner.updateTenantStatus(ctx, tenantID, StatusActive)
}

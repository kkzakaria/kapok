package tenant

import (
	"context"
	"fmt"

	"github.com/kapok/kapok/internal/database"
	"github.com/rs/zerolog"
)

// UsageCollector collects per-tenant resource usage metrics from PostgreSQL
type UsageCollector struct {
	db     *database.DB
	logger zerolog.Logger
}

// NewUsageCollector creates a new UsageCollector
func NewUsageCollector(db *database.DB, logger zerolog.Logger) *UsageCollector {
	return &UsageCollector{
		db:     db,
		logger: logger,
	}
}

// CollectUsage gathers current resource usage for a single tenant
func (u *UsageCollector) CollectUsage(ctx context.Context, tenantID string) (*TenantUsage, error) {
	usage := &TenantUsage{TenantID: tenantID}

	// Get storage usage from tenants table
	err := u.db.QueryRowContext(ctx,
		"SELECT COALESCE(storage_used_bytes, 0) FROM tenants WHERE id = $1", tenantID,
	).Scan(&usage.StorageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage usage: %w", err)
	}

	// Get active connections for the tenant's schema
	var schemaName string
	err = u.db.QueryRowContext(ctx,
		"SELECT schema_name FROM tenants WHERE id = $1", tenantID,
	).Scan(&schemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema name: %w", err)
	}

	// Count active connections (approximate via pg_stat_activity)
	err = u.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM pg_stat_activity
		WHERE datname = current_database()
		AND query LIKE '%' || $1 || '%'
		AND state = 'active'
	`, schemaName).Scan(&usage.ConnectionCount)
	if err != nil {
		u.logger.Debug().Err(err).Str("tenant_id", tenantID).Msg("failed to count connections")
		// Non-fatal: default to 0
	}

	// QPS is computed externally (from metrics/counters); return 0 as default
	usage.QPS = 0

	return usage, nil
}

// CollectAllUsage gathers usage for all active tenants
func (u *UsageCollector) CollectAllUsage(ctx context.Context) ([]*TenantUsage, error) {
	rows, err := u.db.QueryContext(ctx,
		"SELECT id FROM tenants WHERE status = 'active'")
	if err != nil {
		return nil, fmt.Errorf("failed to list active tenants: %w", err)
	}
	defer rows.Close()

	var usages []*TenantUsage
	for rows.Next() {
		var tenantID string
		if err := rows.Scan(&tenantID); err != nil {
			continue
		}
		usage, err := u.CollectUsage(ctx, tenantID)
		if err != nil {
			u.logger.Warn().Err(err).Str("tenant_id", tenantID).Msg("failed to collect usage")
			continue
		}
		usages = append(usages, usage)
	}

	return usages, rows.Err()
}

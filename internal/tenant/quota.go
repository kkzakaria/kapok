package tenant

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kapok/kapok/internal/database"
	"github.com/rs/zerolog"
)

// QuotaManager manages tenant resource quotas
type QuotaManager struct {
	db           *database.DB
	hierarchyMgr *HierarchyManager
	logger       zerolog.Logger
}

// NewQuotaManager creates a new QuotaManager
func NewQuotaManager(db *database.DB, hierarchyMgr *HierarchyManager, logger zerolog.Logger) *QuotaManager {
	return &QuotaManager{
		db:           db,
		hierarchyMgr: hierarchyMgr,
		logger:       logger,
	}
}

// SetQuota creates or updates quotas for a tenant
func (q *QuotaManager) SetQuota(ctx context.Context, quota *Quota) error {
	query := `
		INSERT INTO tenant_quotas (tenant_id, max_storage_bytes, max_connections, max_qps, max_children, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tenant_id) DO UPDATE SET
			max_storage_bytes = EXCLUDED.max_storage_bytes,
			max_connections = EXCLUDED.max_connections,
			max_qps = EXCLUDED.max_qps,
			max_children = EXCLUDED.max_children,
			updated_at = EXCLUDED.updated_at
	`
	now := time.Now()
	_, err := q.db.ExecContext(ctx, query,
		quota.TenantID, quota.MaxStorageBytes, quota.MaxConnections,
		quota.MaxQPS, quota.MaxChildren, now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to set quota: %w", err)
	}

	q.logger.Info().
		Str("tenant_id", quota.TenantID).
		Int64("max_storage_bytes", quota.MaxStorageBytes).
		Int("max_connections", quota.MaxConnections).
		Msg("quota set")

	return nil
}

// GetQuota retrieves quotas for a tenant
func (q *QuotaManager) GetQuota(ctx context.Context, tenantID string) (*Quota, error) {
	query := `
		SELECT tenant_id, max_storage_bytes, max_connections, max_qps, max_children, created_at, updated_at
		FROM tenant_quotas
		WHERE tenant_id = $1
	`
	var quota Quota
	err := q.db.QueryRowContext(ctx, query, tenantID).Scan(
		&quota.TenantID, &quota.MaxStorageBytes, &quota.MaxConnections,
		&quota.MaxQPS, &quota.MaxChildren, &quota.CreatedAt, &quota.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no quota found for tenant: %s", tenantID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}
	return &quota, nil
}

// CheckChildQuota verifies that the parent tenant has capacity for another child
func (q *QuotaManager) CheckChildQuota(ctx context.Context, parentID string) error {
	quota, err := q.GetQuota(ctx, parentID)
	if err != nil {
		// No quota means no limit
		return nil
	}

	if quota.MaxChildren <= 0 {
		return nil
	}

	childCount, err := q.hierarchyMgr.CountChildren(ctx, parentID)
	if err != nil {
		return fmt.Errorf("failed to count children: %w", err)
	}

	if childCount >= quota.MaxChildren {
		return fmt.Errorf("parent tenant %s has reached max children quota (%d)", parentID, quota.MaxChildren)
	}

	return nil
}

// EnforceQuota checks current usage against quota and returns violations
func (q *QuotaManager) EnforceQuota(ctx context.Context, tenantID string, usage *TenantUsage) ([]ThresholdViolation, error) {
	quota, err := q.GetQuota(ctx, tenantID)
	if err != nil {
		return nil, nil // no quota = no enforcement
	}

	var violations []ThresholdViolation

	if quota.MaxStorageBytes > 0 && usage.StorageBytes > quota.MaxStorageBytes {
		violations = append(violations, ThresholdViolation{
			TenantID:   tenantID,
			Metric:     "storage_bytes",
			Current:    float64(usage.StorageBytes),
			Threshold:  float64(quota.MaxStorageBytes),
			Percentage: float64(usage.StorageBytes) / float64(quota.MaxStorageBytes),
		})
	}

	if quota.MaxConnections > 0 && usage.ConnectionCount > quota.MaxConnections {
		violations = append(violations, ThresholdViolation{
			TenantID:   tenantID,
			Metric:     "connections",
			Current:    float64(usage.ConnectionCount),
			Threshold:  float64(quota.MaxConnections),
			Percentage: float64(usage.ConnectionCount) / float64(quota.MaxConnections),
		})
	}

	if quota.MaxQPS > 0 && usage.QPS > quota.MaxQPS {
		violations = append(violations, ThresholdViolation{
			TenantID:   tenantID,
			Metric:     "qps",
			Current:    usage.QPS,
			Threshold:  quota.MaxQPS,
			Percentage: usage.QPS / quota.MaxQPS,
		})
	}

	return violations, nil
}

// RollupUsage aggregates usage from all children of a tenant
func (q *QuotaManager) RollupUsage(ctx context.Context, tenantID string, collector *UsageCollector) (*TenantUsage, error) {
	descendants, err := q.hierarchyMgr.GetSubtree(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subtree: %w", err)
	}

	total := &TenantUsage{TenantID: tenantID}

	// Include self
	selfUsage, err := collector.CollectUsage(ctx, tenantID)
	if err == nil {
		total.StorageBytes += selfUsage.StorageBytes
		total.ConnectionCount += selfUsage.ConnectionCount
		total.QPS += selfUsage.QPS
	}

	for _, d := range descendants {
		usage, err := collector.CollectUsage(ctx, d.ID)
		if err != nil {
			q.logger.Warn().Err(err).Str("tenant_id", d.ID).Msg("failed to collect child usage")
			continue
		}
		total.StorageBytes += usage.StorageBytes
		total.ConnectionCount += usage.ConnectionCount
		total.QPS += usage.QPS
	}

	return total, nil
}

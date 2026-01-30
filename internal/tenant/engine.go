package tenant

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kapok/kapok/internal/database"
	"github.com/rs/zerolog"
)

// DecisionEngine evaluates usage against thresholds and triggers migrations
type DecisionEngine struct {
	db               *database.DB
	usageCollector   *UsageCollector
	thresholdManager *ThresholdManager
	migrationManager *MigrationManager
	provisioner      *Provisioner
	logger           zerolog.Logger

	mu             sync.Mutex
	cooldowns      map[string]time.Time // tenantID -> earliest next migration
	cooldownPeriod time.Duration
	approvalRequired bool
	pendingApprovals map[string]string // tenantID -> recommended toIsolation
}

// NewDecisionEngine creates a new DecisionEngine
func NewDecisionEngine(
	db *database.DB,
	usageCollector *UsageCollector,
	thresholdManager *ThresholdManager,
	migrationManager *MigrationManager,
	provisioner *Provisioner,
	logger zerolog.Logger,
	cooldownPeriod time.Duration,
	approvalRequired bool,
) *DecisionEngine {
	return &DecisionEngine{
		db:               db,
		usageCollector:   usageCollector,
		thresholdManager: thresholdManager,
		migrationManager: migrationManager,
		provisioner:      provisioner,
		logger:           logger,
		cooldowns:        make(map[string]time.Time),
		cooldownPeriod:   cooldownPeriod,
		approvalRequired: approvalRequired,
		pendingApprovals: make(map[string]string),
	}
}

// EvaluateAll checks all active tenants and triggers migrations as needed
func (e *DecisionEngine) EvaluateAll(ctx context.Context) ([]ThresholdViolation, error) {
	usages, err := e.usageCollector.CollectAllUsage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to collect usage: %w", err)
	}

	var allViolations []ThresholdViolation
	for _, usage := range usages {
		tenant, err := e.provisioner.GetTenantByID(ctx, usage.TenantID)
		if err != nil {
			continue
		}

		violations := e.thresholdManager.Evaluate(usage, tenant.Tier)
		if len(violations) == 0 {
			continue
		}

		allViolations = append(allViolations, violations...)
		e.handleViolations(ctx, tenant, violations)
	}

	return allViolations, nil
}

func (e *DecisionEngine) handleViolations(ctx context.Context, tenant *Tenant, violations []ThresholdViolation) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check cooldown
	if cooldownUntil, ok := e.cooldowns[tenant.ID]; ok {
		if time.Now().Before(cooldownUntil) {
			e.logger.Debug().
				Str("tenant_id", tenant.ID).
				Time("cooldown_until", cooldownUntil).
				Msg("tenant in cooldown, skipping migration trigger")
			return
		}
	}

	// Only trigger migration for schema -> database
	if tenant.IsolationLevel != "schema" {
		return
	}

	toIsolation := "database"

	if e.approvalRequired {
		e.pendingApprovals[tenant.ID] = toIsolation
		e.logger.Info().
			Str("tenant_id", tenant.ID).
			Str("recommended", toIsolation).
			Int("violations", len(violations)).
			Msg("migration approval required")
		return
	}

	// Auto-migrate
	e.cooldowns[tenant.ID] = time.Now().Add(e.cooldownPeriod)
	go func() {
		_, err := e.migrationManager.Migrate(ctx, tenant.ID, "schema", toIsolation)
		if err != nil {
			e.logger.Error().Err(err).Str("tenant_id", tenant.ID).Msg("auto-migration failed")
		}
	}()
}

// ApproveMigration approves a pending migration for a tenant
func (e *DecisionEngine) ApproveMigration(ctx context.Context, tenantID string) (*MigrationRecord, error) {
	e.mu.Lock()
	toIsolation, ok := e.pendingApprovals[tenantID]
	if !ok {
		e.mu.Unlock()
		return nil, fmt.Errorf("no pending migration approval for tenant %s", tenantID)
	}
	delete(e.pendingApprovals, tenantID)
	e.cooldowns[tenantID] = time.Now().Add(e.cooldownPeriod)
	e.mu.Unlock()

	tenant, err := e.provisioner.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return e.migrationManager.Migrate(ctx, tenantID, tenant.IsolationLevel, toIsolation)
}

// GetPendingApprovals returns tenants awaiting migration approval
func (e *DecisionEngine) GetPendingApprovals() map[string]string {
	e.mu.Lock()
	defer e.mu.Unlock()
	result := make(map[string]string, len(e.pendingApprovals))
	for k, v := range e.pendingApprovals {
		result[k] = v
	}
	return result
}

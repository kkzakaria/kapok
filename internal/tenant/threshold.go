package tenant

import (
	"sync"

	"github.com/rs/zerolog"
)

// ThresholdManager evaluates tenant usage against configurable tier thresholds
type ThresholdManager struct {
	mu     sync.RWMutex
	tiers  map[string]TierThresholds
	logger zerolog.Logger
}

// NewThresholdManager creates a new ThresholdManager with the given tier configurations
func NewThresholdManager(tiers []TierThresholds, logger zerolog.Logger) *ThresholdManager {
	m := &ThresholdManager{
		tiers:  make(map[string]TierThresholds),
		logger: logger,
	}
	for _, t := range tiers {
		m.tiers[t.Tier] = t
	}
	return m
}

// Evaluate checks usage against the threshold for the given tier and returns violations
func (tm *ThresholdManager) Evaluate(usage *TenantUsage, tier string) []ThresholdViolation {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	t, ok := tm.tiers[tier]
	if !ok {
		tm.logger.Warn().Str("tier", tier).Msg("unknown tier, skipping threshold evaluation")
		return nil
	}

	var violations []ThresholdViolation

	if t.MaxStorageBytes > 0 {
		pct := float64(usage.StorageBytes) / float64(t.MaxStorageBytes)
		if pct >= t.MigrationTrigger {
			violations = append(violations, ThresholdViolation{
				TenantID:   usage.TenantID,
				Metric:     "storage_bytes",
				Current:    float64(usage.StorageBytes),
				Threshold:  float64(t.MaxStorageBytes),
				Percentage: pct,
			})
		}
	}

	if t.MaxConnections > 0 {
		pct := float64(usage.ConnectionCount) / float64(t.MaxConnections)
		if pct >= t.MigrationTrigger {
			violations = append(violations, ThresholdViolation{
				TenantID:   usage.TenantID,
				Metric:     "connections",
				Current:    float64(usage.ConnectionCount),
				Threshold:  float64(t.MaxConnections),
				Percentage: pct,
			})
		}
	}

	if t.MaxQPS > 0 {
		pct := usage.QPS / t.MaxQPS
		if pct >= t.MigrationTrigger {
			violations = append(violations, ThresholdViolation{
				TenantID:   usage.TenantID,
				Metric:     "qps",
				Current:    usage.QPS,
				Threshold:  t.MaxQPS,
				Percentage: pct,
			})
		}
	}

	return violations
}

// SetTier adds or updates a tier threshold configuration
func (tm *ThresholdManager) SetTier(t TierThresholds) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.tiers[t.Tier] = t
}

// GetTier returns the threshold configuration for a tier
func (tm *ThresholdManager) GetTier(tier string) (TierThresholds, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	t, ok := tm.tiers[tier]
	return t, ok
}

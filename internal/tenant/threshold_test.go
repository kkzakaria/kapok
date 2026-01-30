package tenant

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestThresholdManagerEvaluate(t *testing.T) {
	logger := zerolog.Nop()
	tiers := []TierThresholds{
		{Tier: "standard", MaxStorageBytes: 10000, MaxConnections: 100, MaxQPS: 1000, MigrationTrigger: 0.8},
	}
	tm := NewThresholdManager(tiers, logger)

	tests := []struct {
		name       string
		usage      TenantUsage
		tier       string
		violations int
	}{
		{
			name:       "below threshold",
			usage:      TenantUsage{StorageBytes: 5000, ConnectionCount: 50, QPS: 500},
			tier:       "standard",
			violations: 0,
		},
		{
			name:       "at threshold",
			usage:      TenantUsage{StorageBytes: 8000, ConnectionCount: 80, QPS: 800},
			tier:       "standard",
			violations: 3,
		},
		{
			name:       "above threshold",
			usage:      TenantUsage{StorageBytes: 9500, ConnectionCount: 95, QPS: 950},
			tier:       "standard",
			violations: 3,
		},
		{
			name:       "unknown tier",
			usage:      TenantUsage{StorageBytes: 9999},
			tier:       "unknown",
			violations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := tm.Evaluate(&tt.usage, tt.tier)
			assert.Len(t, violations, tt.violations)
		})
	}
}

func TestThresholdManagerSetGetTier(t *testing.T) {
	logger := zerolog.Nop()
	tm := NewThresholdManager(nil, logger)

	_, ok := tm.GetTier("custom")
	assert.False(t, ok)

	tm.SetTier(TierThresholds{Tier: "custom", MaxStorageBytes: 5000})
	tier, ok := tm.GetTier("custom")
	assert.True(t, ok)
	assert.Equal(t, int64(5000), tier.MaxStorageBytes)
}

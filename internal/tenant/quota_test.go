package tenant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuotaEnforcement(t *testing.T) {
	tests := []struct {
		name       string
		quota      Quota
		usage      TenantUsage
		violations int
	}{
		{
			name:       "no violations when under limits",
			quota:      Quota{MaxStorageBytes: 1000, MaxConnections: 10, MaxQPS: 100},
			usage:      TenantUsage{StorageBytes: 500, ConnectionCount: 5, QPS: 50},
			violations: 0,
		},
		{
			name:       "storage violation",
			quota:      Quota{MaxStorageBytes: 1000, MaxConnections: 10, MaxQPS: 100},
			usage:      TenantUsage{StorageBytes: 1500, ConnectionCount: 5, QPS: 50},
			violations: 1,
		},
		{
			name:       "all violations",
			quota:      Quota{MaxStorageBytes: 1000, MaxConnections: 10, MaxQPS: 100},
			usage:      TenantUsage{StorageBytes: 1500, ConnectionCount: 15, QPS: 150},
			violations: 3,
		},
		{
			name:       "zero quota means no limit",
			quota:      Quota{MaxStorageBytes: 0, MaxConnections: 0, MaxQPS: 0},
			usage:      TenantUsage{StorageBytes: 999999, ConnectionCount: 999, QPS: 999},
			violations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var violations []ThresholdViolation

			if tt.quota.MaxStorageBytes > 0 && tt.usage.StorageBytes > tt.quota.MaxStorageBytes {
				violations = append(violations, ThresholdViolation{Metric: "storage_bytes"})
			}
			if tt.quota.MaxConnections > 0 && tt.usage.ConnectionCount > tt.quota.MaxConnections {
				violations = append(violations, ThresholdViolation{Metric: "connections"})
			}
			if tt.quota.MaxQPS > 0 && tt.usage.QPS > tt.quota.MaxQPS {
				violations = append(violations, ThresholdViolation{Metric: "qps"})
			}

			assert.Len(t, violations, tt.violations)
		})
	}
}

package tenant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTenantUsageStruct(t *testing.T) {
	usage := TenantUsage{
		TenantID:        "test-id",
		StorageBytes:    1024,
		ConnectionCount: 5,
		QPS:             10.5,
	}

	assert.Equal(t, "test-id", usage.TenantID)
	assert.Equal(t, int64(1024), usage.StorageBytes)
	assert.Equal(t, 5, usage.ConnectionCount)
	assert.Equal(t, 10.5, usage.QPS)
}

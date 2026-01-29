package observability

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetricsCollector(t *testing.T) {
	reg := prometheus.NewRegistry()
	mc := NewMetricsCollector(reg)
	require.NotNil(t, mc)

	assert.NotNil(t, mc.HTTPRequestsTotal)
	assert.NotNil(t, mc.HTTPRequestDuration)
	assert.NotNil(t, mc.GraphQLQueriesTotal)
	assert.NotNil(t, mc.GraphQLQueryDuration)
	assert.NotNil(t, mc.GraphQLErrorsTotal)
	assert.NotNil(t, mc.DBQueriesTotal)
	assert.NotNil(t, mc.DBQueryDuration)
	assert.NotNil(t, mc.TenantCPUUsage)
	assert.NotNil(t, mc.TenantMemoryUsage)
	assert.NotNil(t, mc.TenantStorageUsage)
}

func TestMetricsCollector_IncrementCounter(t *testing.T) {
	reg := prometheus.NewRegistry()
	mc := NewMetricsCollector(reg)

	mc.HTTPRequestsTotal.WithLabelValues("tenant-1", "GET", "/api/test", "200").Inc()
	mc.DBQueriesTotal.WithLabelValues("SELECT").Inc()

	// Verify metrics are gathered without error
	families, err := reg.Gather()
	require.NoError(t, err)
	assert.NotEmpty(t, families)
}

func TestMetricsCollector_SetTenantResourceUsage(t *testing.T) {
	reg := prometheus.NewRegistry()
	mc := NewMetricsCollector(reg)

	mc.SetTenantResourceUsage("tenant-1", 0.75, 512.0, 0.85)

	families, err := reg.Gather()
	require.NoError(t, err)

	found := map[string]bool{}
	for _, f := range families {
		switch f.GetName() {
		case "kapok_tenant_cpu_usage", "kapok_tenant_memory_usage", "kapok_tenant_storage_usage":
			found[f.GetName()] = true
		}
	}
	assert.True(t, found["kapok_tenant_cpu_usage"])
	assert.True(t, found["kapok_tenant_memory_usage"])
	assert.True(t, found["kapok_tenant_storage_usage"])
}

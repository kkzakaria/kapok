package tenant

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithTenantID_GetTenantID(t *testing.T) {
	ctx := context.Background()
	tenantID := "test-tenant-123"

	// Add tenant ID to context
	ctx = WithTenantID(ctx, tenantID)

	// Retrieve tenant ID
	retrieved, err := GetTenantID(ctx)
	require.NoError(t, err)
	assert.Equal(t, tenantID, retrieved)
}

func TestGetTenantID_NotFound(t *testing.T) {
	ctx := context.Background()

	// Try to get tenant ID from empty context
	_, err := GetTenantID(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant ID not found in context")
}

func TestGetTenantID_EmptyString(t *testing.T) {
	ctx := context.Background()
	ctx = WithTenantID(ctx, "")

	// Empty string should be treated as not found
	_, err := GetTenantID(ctx)
	assert.Error(t, err)
}

func TestMustGetTenantID_Success(t *testing.T) {
	ctx := context.Background()
	tenantID := "test-tenant-123"
	ctx = WithTenantID(ctx, tenantID)

	// Should not panic
	retrieved := MustGetTenantID(ctx)
	assert.Equal(t, tenantID, retrieved)
}

func TestMustGetTenantID_Panic(t *testing.T) {
	ctx := context.Background()

	// Should panic when tenant ID not found
	assert.Panics(t, func() {
		MustGetTenantID(ctx)
	})
}

func TestWithTenant_GetTenant(t *testing.T) {
	ctx := context.Background()
	tenant := &Tenant{
		ID:         "test-tenant-123",
		Name:       "test-tenant",
		SchemaName: "tenant_test-tenant-123",
		Status:     StatusActive,
	}

	// Add tenant to context
	ctx = WithTenant(ctx, tenant)

	// Retrieve tenant
	retrieved, err := GetTenant(ctx)
	require.NoError(t, err)
	assert.Equal(t, tenant, retrieved)

	// Tenant ID should also be set
	tenantID, err := GetTenantID(ctx)
	require.NoError(t, err)
	assert.Equal(t, tenant.ID, tenantID)
}

func TestGetTenant_NotFound(t *testing.T) {
	ctx := context.Background()

	// Try to get tenant from empty context
	_, err := GetTenant(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant not found in context")
}

func TestWithTenant_Nil(t *testing.T) {
	ctx := context.Background()

	// Add nil tenant
	ctx = WithTenant(ctx, nil)

	// Should fail to retrieve
	_, err := GetTenant(ctx)
	assert.Error(t, err)
}

func TestHasTenantID(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() context.Context
		expected bool
	}{
		{
			name: "has tenant ID",
			setup: func() context.Context {
				return WithTenantID(context.Background(), "test-123")
			},
			expected: true,
		},
		{
			name: "no tenant ID",
			setup: func() context.Context {
				return context.Background()
			},
			expected: false,
		},
		{
			name: "empty tenant ID",
			setup: func() context.Context {
				return WithTenantID(context.Background(), "")
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			result := HasTenantID(ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContextKeyIsolation(t *testing.T) {
	ctx := context.Background()

	// Add tenant ID
	ctx = WithTenantID(ctx, "tenant-id-123")

	// Add tenant object
	tenant := &Tenant{
		ID:         "tenant-obj-456", // Different ID
		Name:       "test",
		SchemaName: "tenant_test",
		Status:     StatusActive,
	}
	ctx = WithTenant(ctx, tenant)

	// Tenant object's ID should override the direct tenant ID
	tenantID, err := GetTenantID(ctx)
	require.NoError(t, err)
	assert.Equal(t, tenant.ID, tenantID, "WithTenant should update tenant ID")

	// Tenant object should be retrievable
	retrievedTenant, err := GetTenant(ctx)
	require.NoError(t, err)
	assert.Equal(t, tenant, retrievedTenant)
}

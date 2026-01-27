package tenant

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: These are unit tests that test the provisioner logic
// Integration tests with real PostgreSQL will be in provisioner_integration_test.go

func TestValidateNameInProvisioner(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{
			name:      "valid name",
			input:     "acme-corp",
			wantError: false,
		},
		{
			name:      "invalid - too short",
			input:     "ab",
			wantError: true,
		},
		{
			name:      "invalid - special chars",
			input:     "tenant@name",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTenantLifecycle(t *testing.T) {
	// This test validates the tenant struct lifecycle
	tenant := &Tenant{
		ID:         "test-id-123",
		Name:       "test-tenant",
		SchemaName: "tenant_test-id-123",
		Status:     StatusProvisioning,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Validate initial state
	assert.Equal(t, StatusProvisioning, tenant.Status)
	require.NoError(t, tenant.Validate())

	// Transition to active
	tenant.Status = StatusActive
	require.NoError(t, tenant.Validate())

	// Transition to deleted
	tenant.Status = StatusDeleted
	require.NoError(t, tenant.Validate())
}

func TestTenantStatusTransitions(t *testing.T) {
	validStatuses := []TenantStatus{
		StatusActive,
		StatusProvisioning,
		StatusSuspended,
		StatusDeleted,
	}

	tenant := &Tenant{
		ID:         "test-id",
		Name:       "test",
		SchemaName: "tenant_test",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Test all valid status transitions
	for _, status := range validStatuses {
		t.Run(string(status), func(t *testing.T) {
			tenant.Status = status
			assert.NoError(t, tenant.Validate())
		})
	}

	// Test invalid status
	tenant.Status = "invalid_status"
	assert.Error(t, tenant.Validate())
}

func TestGenerateSchemaNameForProvisioner(t *testing.T) {
	tests := []struct {
		name     string
		tenantID string
		expected string
	}{
		{
			name:     "UUID format",
			tenantID: "550e8400-e29b-41d4-a716-446655440000",
			expected: "tenant_550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:     "simple ID",
			tenantID: "acme-123",
			expected: "tenant_acme-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSchemaName(tt.tenantID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Mock tests for provisioner logic (without database)
func TestProvisionerValidation(t *testing.T) {
	ctx := context.Background()
	
	// Test that invalid tenant names are rejected early
	invalidNames := []string{
		"ab",           // too short
		"a",            // too short
		"tenant@name",  // invalid chars
		"tenant name",  // spaces
		"",             // empty
	}

	for _, name := range invalidNames {
		t.Run("invalid_"+name, func(t *testing.T) {
			err := ValidateName(name)
			assert.Error(t, err, "should reject invalid name: %s", name)
		})
	}

	// Test that valid names pass
	validNames := []string{
		"acme",
		"acme-corp",
		"acme_corp",
		"acme123",
		"test-tenant-name",
	}

	for _, name := range validNames {
		t.Run("valid_"+name, func(t *testing.T) {
			err := ValidateName(name)
			assert.NoError(t, err, "should accept valid name: %s", name)
		})
	}

	_ = ctx // Will be used in integration tests
}

func TestTenantTimestamps(t *testing.T) {
	now := time.Now()
	tenant := &Tenant{
		ID:         "test-id",
		Name:       "test",
		SchemaName: "tenant_test",
		Status:     StatusActive,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	assert.False(t, tenant.CreatedAt.IsZero())
	assert.False(t, tenant.UpdatedAt.IsZero())
	assert.True(t, tenant.CreatedAt.Equal(tenant.UpdatedAt))
}

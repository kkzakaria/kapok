package tenant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid name with alphanumeric",
			input:     "acme-corp",
			wantError: false,
		},
		{
			name:      "valid name with underscores",
			input:     "test_tenant_123",
			wantError: false,
		},
		{
			name:      "valid name minimum length",
			input:     "abc",
			wantError: false,
		},
		{
			name:      "too short",
			input:     "ab",
			wantError: true,
			errorMsg:  "tenant name must be at least 3 characters",
		},
		{
			name:      "too long",
			input:     "this-is-a-very-long-tenant-name-exceeding-the-maximum-allowed-length",
			wantError: true,
			errorMsg:  "tenant name cannot exceed 50 characters",
		},
		{
			name:      "invalid characters - spaces",
			input:     "tenant name",
			wantError: true,
			errorMsg:  "tenant name can only contain alphanumeric characters, hyphens, and underscores",
		},
		{
			name:      "invalid characters - special chars",
			input:     "tenant@name!",
			wantError: true,
			errorMsg:  "tenant name can only contain alphanumeric characters, hyphens, and underscores",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			
			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerateSchemaName(t *testing.T) {
	tests := []struct {
		name     string
		tenantID string
		expected string
	}{
		{
			name:     "basic UUID",
			tenantID: "123e4567-e89b-12d3-a456-426614174000",
			expected: "tenant_123e4567_e89b_12d3_a456_426614174000",
		},
		{
			name:     "simple ID",
			tenantID: "tenant-123",
			expected: "tenant_tenant_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSchemaName(tt.tenantID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTenant_Validate(t *testing.T) {
	tests := []struct {
		name      string
		tenant    Tenant
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid tenant",
			tenant: Tenant{
				ID:         "123e4567-e89b-12d3-a456-426614174000",
				Name:       "acme-corp",
				SchemaName: "tenant_123e4567_e89b_12d3_a456_426614174000",
				Status:     StatusActive,
			},
			wantError: false,
		},
		{
			name: "missing ID",
			tenant: Tenant{
				Name:       "acme-corp",
				SchemaName: "tenant_123",
				Status:     StatusActive,
			},
			wantError: true,
			errorMsg:  "tenant ID is required",
		},
		{
			name: "missing schema name",
			tenant: Tenant{
				ID:     "123e4567-e89b-12d3-a456-426614174000",
				Name:   "acme-corp",
				Status: StatusActive,
			},
			wantError: true,
			errorMsg:  "schema name is required",
		},
		{
			name: "invalid status",
			tenant: Tenant{
				ID:         "123e4567-e89b-12d3-a456-426614174000",
				Name:       "acme-corp",
				SchemaName: "tenant_123",
				Status:     "invalid",
			},
			wantError: true,
			errorMsg:  "invalid tenant status",
		},
		{
			name: "invalid name",
			tenant: Tenant{
				ID:         "123e4567-e89b-12d3-a456-426614174000",
				Name:       "ab", // too short
				SchemaName: "tenant_123",
				Status:     StatusActive,
			},
			wantError: true,
			errorMsg:  "tenant name must be at least",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tenant.Validate()
			
			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTenantStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		status   TenantStatus
		expected string
	}{
		{
			name:     "active status",
			status:   StatusActive,
			expected: "active",
		},
		{
			name:     "provisioning status",
			status:   StatusProvisioning,
			expected: "provisioning",
		},
		{
			name:     "suspended status",
			status:   StatusSuspended,
			expected: "suspended",
		},
		{
			name:     "deleted status",
			status:   StatusDeleted,
			expected: "deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

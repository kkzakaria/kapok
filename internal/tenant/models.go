package tenant

import (
	"fmt"
	"regexp"
	"time"
)

// TenantStatus represents the current status of a tenant
type TenantStatus string

const (
	StatusActive       TenantStatus = "active"
	StatusProvisioning TenantStatus = "provisioning"
	StatusSuspended    TenantStatus = "suspended"
	StatusDeleted      TenantStatus = "deleted"
)

// String returns the string representation of TenantStatus
func (s TenantStatus) String() string {
	return string(s)
}

// Tenant represents a multi-tenant entity in the system
type Tenant struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	SchemaName string       `json:"schema_name"`
	Status     TenantStatus `json:"status"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

// Validation constants
const (
	MaxTenantNameLength = 50
	MinTenantNameLength = 3
)

var (
	// tenantNameRegex allows alphanumeric characters, hyphens, and underscores
	tenantNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// ValidateName validates tenant name according to business rules
func ValidateName(name string) error {
	if len(name) < MinTenantNameLength {
		return fmt.Errorf("tenant name must be at least %d characters", MinTenantNameLength)
	}
	
	if len(name) > MaxTenantNameLength {
		return fmt.Errorf("tenant name cannot exceed %d characters", MaxTenantNameLength)
	}
	
	if !tenantNameRegex.MatchString(name) {
		return fmt.Errorf("tenant name can only contain alphanumeric characters, hyphens, and underscores")
	}
	
	return nil
}

// GenerateSchemaName creates a PostgreSQL schema name from tenant ID
func GenerateSchemaName(tenantID string) string {
	return fmt.Sprintf("tenant_%s", tenantID)
}

// Validate validates the tenant struct
func (t *Tenant) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	
	if err := ValidateName(t.Name); err != nil {
		return err
	}
	
	if t.SchemaName == "" {
		return fmt.Errorf("schema name is required")
	}
	
	// Validate status
	switch t.Status {
	case StatusActive, StatusProvisioning, StatusSuspended, StatusDeleted:
		// Valid status
	default:
		return fmt.Errorf("invalid tenant status: %s", t.Status)
	}
	
	return nil
}

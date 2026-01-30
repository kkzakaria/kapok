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
	StatusMigrating    TenantStatus = "migrating"
)

// HierarchyLevel represents the level of a tenant in the hierarchy
type HierarchyLevel string

const (
	HierarchyOrganization HierarchyLevel = "organization"
	HierarchyProject      HierarchyLevel = "project"
	HierarchyTeam         HierarchyLevel = "team"
)

// MaxHierarchyDepth is the maximum depth of the tenant hierarchy
const MaxHierarchyDepth = 3

// MigrationStatus represents the status of a tenant migration
type MigrationStatus string

const (
	MigrationPending     MigrationStatus = "pending"
	MigrationReplicating MigrationStatus = "replicating"
	MigrationVerifying   MigrationStatus = "verifying"
	MigrationCutover     MigrationStatus = "cutover"
	MigrationCompleted   MigrationStatus = "completed"
	MigrationRolledBack  MigrationStatus = "rolled_back"
	MigrationFailed      MigrationStatus = "failed"
)

// String returns the string representation of TenantStatus
func (s TenantStatus) String() string {
	return string(s)
}

// Tenant represents a multi-tenant entity in the system
type Tenant struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	SchemaName       string         `json:"schema_name"`
	Status           TenantStatus   `json:"status"`
	Slug             string         `json:"slug"`
	IsolationLevel   string         `json:"isolation_level"`
	StorageUsedBytes int64          `json:"storage_used_bytes"`
	LastActivity     *time.Time     `json:"last_activity"`
	ParentID         *string        `json:"parent_id,omitempty"`
	HierarchyLevel   HierarchyLevel `json:"hierarchy_level"`
	Path             string         `json:"path"`
	DatabaseName     string         `json:"database_name,omitempty"`
	DatabaseHost     string         `json:"database_host,omitempty"`
	DatabasePort     int            `json:"database_port,omitempty"`
	Tier             string         `json:"tier"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// MigrationRecord tracks a tenant migration from one isolation level to another
type MigrationRecord struct {
	ID             string          `json:"id"`
	TenantID       string          `json:"tenant_id"`
	FromIsolation  string          `json:"from_isolation"`
	ToIsolation    string          `json:"to_isolation"`
	Status         MigrationStatus `json:"status"`
	StartedAt      time.Time       `json:"started_at"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
	RollbackBefore *time.Time      `json:"rollback_before,omitempty"`
	ErrorMessage   string          `json:"error_message,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

// Quota represents resource quotas for a tenant
type Quota struct {
	TenantID        string    `json:"tenant_id"`
	MaxStorageBytes int64     `json:"max_storage_bytes"`
	MaxConnections  int       `json:"max_connections"`
	MaxQPS          float64   `json:"max_qps"`
	MaxChildren     int       `json:"max_children"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TenantUsage represents current resource usage for a tenant
type TenantUsage struct {
	TenantID        string  `json:"tenant_id"`
	StorageBytes    int64   `json:"storage_bytes"`
	ConnectionCount int     `json:"connection_count"`
	QPS             float64 `json:"qps"`
}

// TierThresholds defines threshold values for a specific tier
type TierThresholds struct {
	Tier             string  `json:"tier"`
	MaxStorageBytes  int64   `json:"max_storage_bytes"`
	MaxConnections   int     `json:"max_connections"`
	MaxQPS           float64 `json:"max_qps"`
	MigrationTrigger float64 `json:"migration_trigger"` // percentage 0-1
}

// ThresholdViolation represents a threshold that has been exceeded
type ThresholdViolation struct {
	TenantID   string  `json:"tenant_id"`
	Metric     string  `json:"metric"`
	Current    float64 `json:"current"`
	Threshold  float64 `json:"threshold"`
	Percentage float64 `json:"percentage"`
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
// PostgreSQL identifiers cannot contain hyphens, so we replace them with underscores
func GenerateSchemaName(tenantID string) string {
	// Replace hyphens from UUID with underscores for PostgreSQL compatibility
	sanitized := regexp.MustCompile(`-`).ReplaceAllString(tenantID, "_")
	return fmt.Sprintf("tenant_%s", sanitized)
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
	case StatusActive, StatusProvisioning, StatusSuspended, StatusDeleted, StatusMigrating:
		// Valid status
	default:
		return fmt.Errorf("invalid tenant status: %s", t.Status)
	}

	return nil
}

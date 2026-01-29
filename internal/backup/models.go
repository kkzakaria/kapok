package backup

import (
	"fmt"
	"time"
)

// Backup status constants
const (
	StatusPending    = "pending"
	StatusRunning    = "running"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusDeleting   = "deleting"
	StatusRestoring  = "restoring"
)

// Backup type constants
const (
	TypeFull   = "full"
	TypeSchema = "schema"
)

// Trigger constants
const (
	TriggerManual    = "manual"
	TriggerScheduled = "scheduled"
	TriggerAPI       = "api"
)

// Backup represents a single backup record.
type Backup struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	SchemaName    string     `json:"schema_name"`
	Status        string     `json:"status"`
	Type          string     `json:"type"`
	Trigger       string     `json:"trigger"`
	StoragePath   string     `json:"storage_path"`
	SizeBytes     int64      `json:"size_bytes"`
	Checksum      string     `json:"checksum"`
	Encrypted     bool       `json:"encrypted"`
	Compressed    bool       `json:"compressed"`
	ErrorMessage  string     `json:"error_message,omitempty"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// BackupSchedule represents a cron-based backup schedule.
type BackupSchedule struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	CronExpr   string    `json:"cron_expr"`
	Enabled    bool      `json:"enabled"`
	RetentionDays int   `json:"retention_days"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Validate validates the backup model.
func (b *Backup) Validate() error {
	if b.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if b.SchemaName == "" {
		return fmt.Errorf("schema_name is required")
	}
	validStatuses := map[string]bool{
		StatusPending: true, StatusRunning: true, StatusCompleted: true,
		StatusFailed: true, StatusDeleting: true, StatusRestoring: true,
	}
	if b.Status != "" && !validStatuses[b.Status] {
		return fmt.Errorf("invalid status: %s", b.Status)
	}
	validTypes := map[string]bool{TypeFull: true, TypeSchema: true}
	if b.Type != "" && !validTypes[b.Type] {
		return fmt.Errorf("invalid type: %s", b.Type)
	}
	validTriggers := map[string]bool{TriggerManual: true, TriggerScheduled: true, TriggerAPI: true}
	if b.Trigger != "" && !validTriggers[b.Trigger] {
		return fmt.Errorf("invalid trigger: %s", b.Trigger)
	}
	return nil
}

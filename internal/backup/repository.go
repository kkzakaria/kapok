package backup

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kapok/kapok/internal/database"
)

// Repository handles CRUD operations for backups.
type Repository struct {
	db *database.DB
}

// NewRepository creates a new backup repository.
func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new backup record.
func (r *Repository) Create(ctx context.Context, b *Backup) error {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO backups (tenant_id, schema_name, status, type, trigger, storage_path, encrypted, compressed, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`, b.TenantID, b.SchemaName, b.Status, b.Type, b.Trigger, b.StoragePath, b.Encrypted, b.Compressed, b.ExpiresAt,
	).Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	return nil
}

// GetByID retrieves a backup by ID.
func (r *Repository) GetByID(ctx context.Context, id string) (*Backup, error) {
	b := &Backup{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, schema_name, status, type, trigger, storage_path,
			   size_bytes, checksum, encrypted, compressed, error_message,
			   started_at, completed_at, expires_at, created_at, updated_at
		FROM backups WHERE id = $1
	`, id).Scan(
		&b.ID, &b.TenantID, &b.SchemaName, &b.Status, &b.Type, &b.Trigger, &b.StoragePath,
		&b.SizeBytes, &b.Checksum, &b.Encrypted, &b.Compressed, &b.ErrorMessage,
		&b.StartedAt, &b.CompletedAt, &b.ExpiresAt, &b.CreatedAt, &b.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("backup not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get backup: %w", err)
	}
	return b, nil
}

// ListByTenant returns backups for a given tenant, ordered by creation time desc.
func (r *Repository) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*Backup, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, schema_name, status, type, trigger, storage_path,
			   size_bytes, checksum, encrypted, compressed, error_message,
			   started_at, completed_at, expires_at, created_at, updated_at
		FROM backups WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}
	defer rows.Close()
	return scanBackups(rows)
}

// UpdateStatus sets the status and optionally the error message.
func (r *Repository) UpdateStatus(ctx context.Context, id, status, errMsg string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE backups SET status = $1, error_message = $2, updated_at = NOW() WHERE id = $3
	`, status, errMsg, id)
	if err != nil {
		return fmt.Errorf("failed to update backup status: %w", err)
	}
	return nil
}

// UpdateCompleted marks a backup as completed with size, checksum, and timestamps.
func (r *Repository) UpdateCompleted(ctx context.Context, id string, sizeBytes int64, checksum string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE backups SET status = $1, size_bytes = $2, checksum = $3,
			   completed_at = $4, updated_at = $4 WHERE id = $5
	`, StatusCompleted, sizeBytes, checksum, now, id)
	if err != nil {
		return fmt.Errorf("failed to update completed backup: %w", err)
	}
	return nil
}

// Delete removes a backup record.
func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM backups WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}
	return nil
}

// ListExpired returns backups whose expires_at is in the past.
func (r *Repository) ListExpired(ctx context.Context) ([]*Backup, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, schema_name, status, type, trigger, storage_path,
			   size_bytes, checksum, encrypted, compressed, error_message,
			   started_at, completed_at, expires_at, created_at, updated_at
		FROM backups WHERE expires_at IS NOT NULL AND expires_at < NOW() AND status = $1
	`, StatusCompleted)
	if err != nil {
		return nil, fmt.Errorf("failed to list expired backups: %w", err)
	}
	defer rows.Close()
	return scanBackups(rows)
}

func scanBackups(rows *sql.Rows) ([]*Backup, error) {
	var backups []*Backup
	for rows.Next() {
		b := &Backup{}
		if err := rows.Scan(
			&b.ID, &b.TenantID, &b.SchemaName, &b.Status, &b.Type, &b.Trigger, &b.StoragePath,
			&b.SizeBytes, &b.Checksum, &b.Encrypted, &b.Compressed, &b.ErrorMessage,
			&b.StartedAt, &b.CompletedAt, &b.ExpiresAt, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan backup: %w", err)
		}
		backups = append(backups, b)
	}
	return backups, rows.Err()
}

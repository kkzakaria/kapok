package backup

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/kapok/kapok/internal/backup/storage"
	"github.com/kapok/kapok/internal/database"
	"github.com/rs/zerolog"
)

// Service is the core backup orchestrator.
type Service struct {
	repo          *Repository
	store         storage.Store
	db            *database.DB
	encryptionKey []byte // 32 bytes for AES-256; nil means no encryption
	logger        zerolog.Logger
	retentionDays int
}

// NewService creates a new backup service.
func NewService(db *database.DB, store storage.Store, encryptionKey []byte, retentionDays int, logger zerolog.Logger) *Service {
	return &Service{
		repo:          NewRepository(db),
		store:         store,
		db:            db,
		encryptionKey: encryptionKey,
		logger:        logger,
		retentionDays: retentionDays,
	}
}

// GetRepository exposes the repository for API handlers.
func (s *Service) GetRepository() *Repository {
	return s.repo
}

// CreateBackup runs pg_dump → compress → encrypt → upload for a tenant schema.
func (s *Service) CreateBackup(ctx context.Context, tenantID, schemaName, trigger string) (*Backup, error) {
	var expiresAt *time.Time
	if s.retentionDays > 0 {
		t := time.Now().Add(time.Duration(s.retentionDays) * 24 * time.Hour)
		expiresAt = &t
	}

	b := &Backup{
		TenantID:   tenantID,
		SchemaName: schemaName,
		Status:     StatusPending,
		Type:       TypeSchema,
		Trigger:    trigger,
		Encrypted:  len(s.encryptionKey) == 32,
		Compressed: true,
		ExpiresAt:  expiresAt,
	}

	storagePath := fmt.Sprintf("backups/%s/%s.sql.gz", tenantID, time.Now().UTC().Format("20060102T150405Z"))
	if b.Encrypted {
		storagePath += ".enc"
	}
	b.StoragePath = storagePath

	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}

	// Run async
	go s.executeBackup(b)
	return b, nil
}

func (s *Service) executeBackup(b *Backup) {
	ctx := context.Background()
	now := time.Now()
	b.StartedAt = &now

	if err := s.repo.UpdateStatus(ctx, b.ID, StatusRunning, ""); err != nil {
		s.logger.Error().Err(err).Str("backup_id", b.ID).Msg("failed to set running status")
		return
	}

	// pg_dump
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		s.db.Config().Host, s.db.Config().Port, s.db.Config().User,
		s.db.Config().Password, s.db.Config().Database, s.db.Config().SSLMode)

	cmd := exec.CommandContext(ctx, "pg_dump", "--schema="+b.SchemaName, "--no-owner", "--no-acl", connStr)
	dumpOut, err := cmd.Output()
	if err != nil {
		s.failBackup(ctx, b, fmt.Sprintf("pg_dump failed: %v", err))
		return
	}

	// Compress
	var compressed bytes.Buffer
	if err := Compress(&compressed, bytes.NewReader(dumpOut)); err != nil {
		s.failBackup(ctx, b, fmt.Sprintf("compression failed: %v", err))
		return
	}

	// Checksum (on compressed data before encryption)
	checksum, err := Checksum(bytes.NewReader(compressed.Bytes()))
	if err != nil {
		s.failBackup(ctx, b, fmt.Sprintf("checksum failed: %v", err))
		return
	}

	// Encrypt
	var uploadData bytes.Buffer
	if b.Encrypted {
		if err := Encrypt(&uploadData, bytes.NewReader(compressed.Bytes()), s.encryptionKey); err != nil {
			s.failBackup(ctx, b, fmt.Sprintf("encryption failed: %v", err))
			return
		}
	} else {
		uploadData = compressed
	}

	// Upload
	if err := s.store.Upload(ctx, b.StoragePath, bytes.NewReader(uploadData.Bytes())); err != nil {
		s.failBackup(ctx, b, fmt.Sprintf("upload failed: %v", err))
		return
	}

	if err := s.repo.UpdateCompleted(ctx, b.ID, int64(uploadData.Len()), checksum); err != nil {
		s.logger.Error().Err(err).Str("backup_id", b.ID).Msg("failed to mark backup completed")
		return
	}

	s.logger.Info().
		Str("backup_id", b.ID).
		Str("tenant_id", b.TenantID).
		Int("size_bytes", uploadData.Len()).
		Msg("backup completed")
}

func (s *Service) failBackup(ctx context.Context, b *Backup, errMsg string) {
	s.logger.Error().Str("backup_id", b.ID).Msg(errMsg)
	if err := s.repo.UpdateStatus(ctx, b.ID, StatusFailed, errMsg); err != nil {
		s.logger.Error().Err(err).Msg("failed to update backup failure status")
	}
}

// RestoreBackup downloads → decrypts → decompresses → pg_restore for a backup.
func (s *Service) RestoreBackup(ctx context.Context, backupID string) error {
	b, err := s.repo.GetByID(ctx, backupID)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(ctx, b.ID, StatusRestoring, ""); err != nil {
		return err
	}

	// Download
	rc, err := s.store.Download(ctx, b.StoragePath)
	if err != nil {
		s.failBackup(ctx, b, fmt.Sprintf("download failed: %v", err))
		return err
	}
	defer rc.Close()

	var raw bytes.Buffer
	if _, err := raw.ReadFrom(rc); err != nil {
		s.failBackup(ctx, b, fmt.Sprintf("read failed: %v", err))
		return err
	}

	// Decrypt
	var decrypted bytes.Buffer
	if b.Encrypted {
		if err := Decrypt(&decrypted, bytes.NewReader(raw.Bytes()), s.encryptionKey); err != nil {
			s.failBackup(ctx, b, fmt.Sprintf("decryption failed: %v", err))
			return err
		}
	} else {
		decrypted = raw
	}

	// Decompress
	var sqlData bytes.Buffer
	if err := Decompress(&sqlData, bytes.NewReader(decrypted.Bytes())); err != nil {
		s.failBackup(ctx, b, fmt.Sprintf("decompression failed: %v", err))
		return err
	}

	// pg_restore via psql (schema-level SQL dump)
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		s.db.Config().Host, s.db.Config().Port, s.db.Config().User,
		s.db.Config().Password, s.db.Config().Database, s.db.Config().SSLMode)

	cmd := exec.CommandContext(ctx, "psql", connStr)
	cmd.Stdin = bytes.NewReader(sqlData.Bytes())
	if out, err := cmd.CombinedOutput(); err != nil {
		errMsg := fmt.Sprintf("psql restore failed: %v: %s", err, string(out))
		s.failBackup(ctx, b, errMsg)
		return fmt.Errorf("%s", errMsg)
	}

	if err := s.repo.UpdateStatus(ctx, b.ID, StatusCompleted, ""); err != nil {
		return err
	}

	s.logger.Info().Str("backup_id", b.ID).Str("tenant_id", b.TenantID).Msg("restore completed")
	return nil
}

// DeleteBackup removes a backup from storage and the database.
func (s *Service) DeleteBackup(ctx context.Context, backupID string) error {
	b, err := s.repo.GetByID(ctx, backupID)
	if err != nil {
		return err
	}

	if err := s.store.Delete(ctx, b.StoragePath); err != nil {
		s.logger.Warn().Err(err).Str("path", b.StoragePath).Msg("failed to delete backup file (may not exist)")
	}

	return s.repo.Delete(ctx, b.ID)
}

// CleanupExpired removes all expired backups.
func (s *Service) CleanupExpired(ctx context.Context) error {
	expired, err := s.repo.ListExpired(ctx)
	if err != nil {
		return err
	}
	for _, b := range expired {
		if err := s.DeleteBackup(ctx, b.ID); err != nil {
			s.logger.Error().Err(err).Str("backup_id", b.ID).Msg("failed to cleanup expired backup")
		}
	}
	s.logger.Info().Int("count", len(expired)).Msg("expired backups cleaned up")
	return nil
}

// BackupAllTenants triggers a backup for every active tenant.
func (s *Service) BackupAllTenants(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `SELECT id, schema_name FROM tenants WHERE status = 'active'`)
	if err != nil {
		return fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()

	var failures int
	for rows.Next() {
		var id, schema string
		if err := rows.Scan(&id, &schema); err != nil {
			failures++
			continue
		}
		if _, err := s.CreateBackup(ctx, id, schema, TriggerScheduled); err != nil {
			s.logger.Error().Err(err).Str("tenant_id", id).Msg("scheduled backup failed")
			failures++
		}
	}
	if failures > 0 {
		return fmt.Errorf("%d tenant backups failed", failures)
	}
	return rows.Err()
}

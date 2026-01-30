package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kapok/kapok/internal/backup"
)

// TriggerBackup starts a new backup for a tenant.
func TriggerBackup(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := chi.URLParam(r, "id")

		// Look up tenant to get schema name
		var schemaName string
		err := deps.DB.QueryRowContext(r.Context(),
			`SELECT schema_name FROM tenants WHERE id = $1`, tenantID,
		).Scan(&schemaName)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				errorResponse(w, http.StatusNotFound, "tenant not found")
				return
			}
			errorResponse(w, http.StatusInternalServerError, "failed to look up tenant")
			return
		}

		b, err := deps.BackupService.CreateBackup(r.Context(), tenantID, schemaName, backup.TriggerAPI)
		if err != nil {
			deps.Logger.Error().Err(err).Str("tenant_id", tenantID).Msg("failed to trigger backup")
			errorResponse(w, http.StatusInternalServerError, "failed to trigger backup")
			return
		}

		writeJSON(w, http.StatusAccepted, b)
	}
}

// ListBackups returns all backups for a tenant.
func ListBackups(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := chi.URLParam(r, "id")
		backups, err := deps.BackupService.GetRepository().ListByTenant(r.Context(), tenantID, 100, 0)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "failed to list backups")
			return
		}
		if backups == nil {
			backups = []*backup.Backup{}
		}
		writeJSON(w, http.StatusOK, backups)
	}
}

// GetBackup returns a single backup by ID.
func GetBackup(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		backupID := chi.URLParam(r, "backupId")
		b, err := deps.BackupService.GetRepository().GetByID(r.Context(), backupID)
		if err != nil {
			if errors.Is(err, backup.ErrBackupNotFound) {
				errorResponse(w, http.StatusNotFound, "backup not found")
				return
			}
			errorResponse(w, http.StatusInternalServerError, "failed to get backup")
			return
		}
		writeJSON(w, http.StatusOK, b)
	}
}

// RestoreBackup triggers a restore from a backup.
func RestoreBackup(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		backupID := chi.URLParam(r, "backupId")
		if err := deps.BackupService.RestoreBackup(r.Context(), backupID); err != nil {
			if errors.Is(err, backup.ErrBackupNotFound) {
				errorResponse(w, http.StatusNotFound, "backup not found")
				return
			}
			deps.Logger.Error().Err(err).Str("backup_id", backupID).Msg("restore failed")
			errorResponse(w, http.StatusInternalServerError, "restore failed")
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]string{"status": "restoring"})
	}
}

// DeleteBackup deletes a backup.
func DeleteBackup(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		backupID := chi.URLParam(r, "backupId")
		if err := deps.BackupService.DeleteBackup(r.Context(), backupID); err != nil {
			if errors.Is(err, backup.ErrBackupNotFound) {
				errorResponse(w, http.StatusNotFound, "backup not found")
				return
			}
			errorResponse(w, http.StatusInternalServerError, "failed to delete backup")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}

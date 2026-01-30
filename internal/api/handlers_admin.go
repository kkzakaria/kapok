package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/kapok/kapok/internal/tenant"
)

// Stats returns aggregate dashboard statistics.
func Stats(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var totalTenants, activeTenants int
		if err := deps.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM tenants`).Scan(&totalTenants); err != nil {
			deps.Logger.Error().Err(err).Msg("failed to count tenants")
		}
		if err := deps.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM tenants WHERE status = 'active'`).Scan(&activeTenants); err != nil {
			deps.Logger.Error().Err(err).Msg("failed to count active tenants")
		}

		var totalStorageBytes int64
		if err := deps.DB.QueryRowContext(ctx, `SELECT COALESCE(SUM(storage_used_bytes), 0) FROM tenants`).Scan(&totalStorageBytes); err != nil {
			deps.Logger.Error().Err(err).Msg("failed to sum storage bytes")
		}

		var totalQueriesToday int
		if err := deps.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM audit_log WHERE timestamp > CURRENT_DATE`).Scan(&totalQueriesToday); err != nil {
			deps.Logger.Warn().Err(err).Msg("failed to count today's queries (audit_log may not exist)")
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"total_tenants":       totalTenants,
			"active_tenants":      activeTenants,
			"total_storage_bytes": totalStorageBytes,
			"total_queries_today": totalQueriesToday,
		})
	}
}

type createTenantRequest struct {
	Name           string `json:"name"`
	IsolationLevel string `json:"isolation_level"`
}

// ListTenants returns all tenants.
func ListTenants(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenants, err := deps.Provisioner.ListTenants(r.Context(), "", 100, 0)
		if err != nil {
			deps.Logger.Error().Err(err).Msg("failed to list tenants")
			errorResponse(w, http.StatusInternalServerError, "failed to list tenants")
			return
		}
		if tenants == nil {
			tenants = []*tenant.Tenant{}
		}
		writeJSON(w, http.StatusOK, tenants)
	}
}

// GetTenant returns a single tenant by ID.
func GetTenant(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		t, err := deps.Provisioner.GetTenantByID(r.Context(), id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				errorResponse(w, http.StatusNotFound, "tenant not found")
				return
			}
			errorResponse(w, http.StatusInternalServerError, "failed to get tenant")
			return
		}
		writeJSON(w, http.StatusOK, t)
	}
}

// CreateTenant provisions a new tenant.
func CreateTenant(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createTenantRequest
		if err := readJSON(r, &req); err != nil {
			errorResponse(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Name == "" {
			errorResponse(w, http.StatusBadRequest, "name is required")
			return
		}

		t, err := deps.Provisioner.CreateTenant(r.Context(), req.Name)
		if err != nil {
			deps.Logger.Error().Err(err).Str("name", req.Name).Msg("failed to create tenant")
			errorResponse(w, http.StatusInternalServerError, "failed to create tenant")
			return
		}

		writeJSON(w, http.StatusCreated, t)
	}
}

// DeleteTenant soft-deletes a tenant.
func DeleteTenant(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := deps.Provisioner.DeleteTenant(r.Context(), id); err != nil {
			if strings.Contains(err.Error(), "not found") {
				errorResponse(w, http.StatusNotFound, "tenant not found")
				return
			}
			errorResponse(w, http.StatusInternalServerError, "failed to delete tenant")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}

// Metrics returns time-series metrics matching the MetricsResponse UI type.
func Metrics(deps *Dependencies) http.HandlerFunc {
	type dataPoint struct {
		Timestamp string  `json:"timestamp"`
		Value     float64 `json:"value"`
	}
	type series struct {
		Label string      `json:"label"`
		Data  []dataPoint `json:"data"`
	}

	emptySeries := func(label string) series {
		return series{Label: label, Data: []dataPoint{}}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rangeParam := r.URL.Query().Get("range")
		if rangeParam == "" {
			rangeParam = "24h"
		}

		interval := "24 hours"
		switch rangeParam {
		case "1h":
			interval = "1 hour"
		case "7d":
			interval = "7 days"
		case "30d":
			interval = "30 days"
		}

		// Generate time-series from audit_log bucketed by hour
		rows, err := deps.DB.QueryContext(ctx, `
			SELECT date_trunc('hour', timestamp) AS bucket, COUNT(*) AS cnt
			FROM audit_log
			WHERE timestamp > NOW() - $1::interval
			GROUP BY bucket
			ORDER BY bucket
		`, interval)
		if err != nil {
			deps.Logger.Error().Err(err).Msg("failed to query metrics")
			errorResponse(w, http.StatusInternalServerError, "failed to query metrics")
			return
		}
		defer rows.Close()

		var throughputData []dataPoint
		for rows.Next() {
			var ts string
			var cnt float64
			if err := rows.Scan(&ts, &cnt); err != nil {
				continue
			}
			throughputData = append(throughputData, dataPoint{Timestamp: ts, Value: cnt})
		}
		if throughputData == nil {
			throughputData = []dataPoint{}
		}

		_ = ctx // suppress unused

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"query_latency_p50": emptySeries("p50"),
			"query_latency_p95": emptySeries("p95"),
			"query_latency_p99": emptySeries("p99"),
			"error_rate":        emptySeries("errors"),
			"throughput":        series{Label: "throughput", Data: throughputData},
		})
	}
}

// ListChildren returns direct children of a tenant (Story 5)
func ListChildren(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if deps.HierarchyManager == nil {
			errorResponse(w, http.StatusNotImplemented, "hierarchy not configured")
			return
		}
		children, err := deps.HierarchyManager.GetChildren(r.Context(), id)
		if err != nil {
			deps.Logger.Error().Err(err).Str("parent_id", id).Msg("failed to list children")
			errorResponse(w, http.StatusInternalServerError, "failed to list children")
			return
		}
		if children == nil {
			children = []*tenant.Tenant{}
		}
		writeJSON(w, http.StatusOK, children)
	}
}

// CreateChildTenant creates a child tenant (Story 5)
func CreateChildTenant(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parentID := chi.URLParam(r, "id")
		var req createTenantRequest
		if err := readJSON(r, &req); err != nil {
			errorResponse(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Name == "" {
			errorResponse(w, http.StatusBadRequest, "name is required")
			return
		}

		// Check quota if quota manager is available
		if deps.QuotaManager != nil {
			if err := deps.QuotaManager.CheckChildQuota(r.Context(), parentID); err != nil {
				errorResponse(w, http.StatusForbidden, err.Error())
				return
			}
		}

		t, err := deps.Provisioner.CreateChildTenant(r.Context(), parentID, req.Name)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				errorResponse(w, http.StatusNotFound, "parent tenant not found")
				return
			}
			if strings.Contains(err.Error(), "cannot create child") {
				errorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			deps.Logger.Error().Err(err).Msg("failed to create child tenant")
			errorResponse(w, http.StatusInternalServerError, "failed to create child tenant")
			return
		}
		writeJSON(w, http.StatusCreated, t)
	}
}

// GetQuota returns the quota for a tenant (Story 6)
func GetQuota(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if deps.QuotaManager == nil {
			errorResponse(w, http.StatusNotImplemented, "quotas not configured")
			return
		}
		quota, err := deps.QuotaManager.GetQuota(r.Context(), id)
		if err != nil {
			if strings.Contains(err.Error(), "no quota found") {
				errorResponse(w, http.StatusNotFound, "no quota set for tenant")
				return
			}
			errorResponse(w, http.StatusInternalServerError, "failed to get quota")
			return
		}
		writeJSON(w, http.StatusOK, quota)
	}
}

// SetQuota sets quotas for a tenant (Story 6)
func SetQuota(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if deps.QuotaManager == nil {
			errorResponse(w, http.StatusNotImplemented, "quotas not configured")
			return
		}
		var quota tenant.Quota
		if err := readJSON(r, &quota); err != nil {
			errorResponse(w, http.StatusBadRequest, "invalid request body")
			return
		}
		quota.TenantID = id
		if err := deps.QuotaManager.SetQuota(r.Context(), &quota); err != nil {
			deps.Logger.Error().Err(err).Msg("failed to set quota")
			errorResponse(w, http.StatusInternalServerError, "failed to set quota")
			return
		}
		writeJSON(w, http.StatusOK, &quota)
	}
}

// GetTenantUsage returns current resource usage for a tenant (Story 3)
func GetTenantUsage(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if deps.UsageCollector == nil {
			errorResponse(w, http.StatusNotImplemented, "usage monitoring not configured")
			return
		}
		usage, err := deps.UsageCollector.CollectUsage(r.Context(), id)
		if err != nil {
			deps.Logger.Error().Err(err).Str("tenant_id", id).Msg("failed to collect usage")
			errorResponse(w, http.StatusInternalServerError, "failed to collect usage")
			return
		}
		writeJSON(w, http.StatusOK, usage)
	}
}

// MigrateTenant initiates a migration for a tenant (Story 2)
func MigrateTenant(deps *Dependencies) http.HandlerFunc {
	type migrateRequest struct {
		ToIsolation string `json:"to_isolation"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if deps.MigrationManager == nil {
			errorResponse(w, http.StatusNotImplemented, "migration not configured")
			return
		}
		var req migrateRequest
		if err := readJSON(r, &req); err != nil {
			errorResponse(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.ToIsolation == "" {
			errorResponse(w, http.StatusBadRequest, "to_isolation is required")
			return
		}

		t, err := deps.Provisioner.GetTenantByID(r.Context(), id)
		if err != nil {
			errorResponse(w, http.StatusNotFound, "tenant not found")
			return
		}

		record, err := deps.MigrationManager.Migrate(r.Context(), id, t.IsolationLevel, req.ToIsolation)
		if err != nil {
			deps.Logger.Error().Err(err).Str("tenant_id", id).Msg("migration failed")
			errorResponse(w, http.StatusInternalServerError, "migration failed: "+err.Error())
			return
		}
		writeJSON(w, http.StatusAccepted, record)
	}
}

// GetMigration returns a migration record (Story 2)
func GetMigration(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if deps.MigrationManager == nil {
			errorResponse(w, http.StatusNotImplemented, "migration not configured")
			return
		}
		record, err := deps.MigrationManager.GetMigration(r.Context(), id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				errorResponse(w, http.StatusNotFound, "migration not found")
				return
			}
			errorResponse(w, http.StatusInternalServerError, "failed to get migration")
			return
		}
		writeJSON(w, http.StatusOK, record)
	}
}

// RollbackMigration rolls back a completed migration (Story 2)
func RollbackMigration(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if deps.MigrationManager == nil {
			errorResponse(w, http.StatusNotImplemented, "migration not configured")
			return
		}
		if err := deps.MigrationManager.Rollback(r.Context(), id); err != nil {
			deps.Logger.Error().Err(err).Str("migration_id", id).Msg("rollback failed")
			errorResponse(w, http.StatusInternalServerError, "rollback failed: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "rolled_back"})
	}
}

// ApproveMigration approves a pending auto-migration (Story 4)
func ApproveMigration(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if deps.DecisionEngine == nil {
			errorResponse(w, http.StatusNotImplemented, "auto-migration not configured")
			return
		}
		record, err := deps.DecisionEngine.ApproveMigration(r.Context(), id)
		if err != nil {
			deps.Logger.Error().Err(err).Str("tenant_id", id).Msg("approve migration failed")
			errorResponse(w, http.StatusInternalServerError, "approve migration failed: "+err.Error())
			return
		}
		writeJSON(w, http.StatusAccepted, record)
	}
}

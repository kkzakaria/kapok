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
		_ = deps.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM tenants`).Scan(&totalTenants)
		_ = deps.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM tenants WHERE status = 'active'`).Scan(&activeTenants)

		var totalStorageBytes int64
		_ = deps.DB.QueryRowContext(ctx, `SELECT COALESCE(SUM(storage_used_bytes), 0) FROM tenants`).Scan(&totalStorageBytes)

		var totalQueriesToday int
		_ = deps.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM audit_log WHERE timestamp > CURRENT_DATE`).Scan(&totalQueriesToday)

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"total_tenants":      totalTenants,
			"active_tenants":     activeTenants,
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

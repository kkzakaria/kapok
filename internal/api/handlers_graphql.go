package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kapok/kapok/internal/tenant"
)

// GraphQLProxy extracts tenantId from the URL, loads the tenant, injects it
// into the request context, and delegates to the existing graphql.Handler.
func GraphQLProxy(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := chi.URLParam(r, "tenantId")
		if tenantID == "" {
			errorResponse(w, http.StatusBadRequest, "tenantId is required")
			return
		}

		t, err := deps.Provisioner.GetTenantByID(r.Context(), tenantID)
		if err != nil {
			errorResponse(w, http.StatusNotFound, "tenant not found")
			return
		}

		ctx := tenant.WithTenant(r.Context(), t)
		deps.GQLHandler.ServeHTTP(w, r.WithContext(ctx))
	}
}

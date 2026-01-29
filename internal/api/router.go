package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type contextKeyType string

const claimsContextKey contextKeyType = "claims"

// NewRouter creates the chi router with all routes.
func NewRouter(deps *Dependencies) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   deps.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Public routes
	r.Post("/api/v1/auth/login", Login(deps))

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(deps))

		r.Get("/api/v1/auth/me", Me(deps))

		// Admin (requires admin role)
		r.Group(func(r chi.Router) {
			r.Use(RequireRole("admin"))
			r.Get("/api/v1/admin/stats", Stats(deps))
			r.Get("/api/v1/admin/tenants", ListTenants(deps))
			r.Get("/api/v1/admin/tenants/{id}", GetTenant(deps))
			r.Post("/api/v1/admin/tenants", CreateTenant(deps))
			r.Delete("/api/v1/admin/tenants/{id}", DeleteTenant(deps))
			r.Get("/api/v1/admin/metrics", Metrics(deps))

			// Backup routes
			r.Post("/api/v1/admin/tenants/{id}/backups", TriggerBackup(deps))
			r.Get("/api/v1/admin/tenants/{id}/backups", ListBackups(deps))
			r.Get("/api/v1/admin/backups/{backupId}", GetBackup(deps))
			r.Post("/api/v1/admin/backups/{backupId}/restore", RestoreBackup(deps))
			r.Delete("/api/v1/admin/backups/{backupId}", DeleteBackup(deps))
		})

		// GraphQL proxy
		r.Post("/api/v1/tenants/{tenantId}/graphql", GraphQLProxy(deps))
	})

	return r
}

// AuthMiddleware validates the JWT Bearer token and injects claims into context.
func AuthMiddleware(deps *Dependencies) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				errorResponse(w, http.StatusUnauthorized, "missing or invalid authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")
			claims, err := deps.JWTManager.ValidateToken(tokenStr)
			if err != nil {
				errorResponse(w, http.StatusUnauthorized, "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), claimsContextKey, map[string]interface{}(claims))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole returns middleware that checks the JWT claims for the given role.
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(claimsContextKey).(map[string]interface{})
			if !ok {
				errorResponse(w, http.StatusForbidden, "forbidden")
				return
			}

			if hasRole(claims, role) {
				next.ServeHTTP(w, r)
				return
			}

			errorResponse(w, http.StatusForbidden, "forbidden: requires "+role+" role")
		})
	}
}

func hasRole(claims map[string]interface{}, role string) bool {
	// Check []interface{} format (JSON array from JWT)
	if roles, ok := claims["roles"].([]interface{}); ok {
		for _, v := range roles {
			if s, ok := v.(string); ok && s == role {
				return true
			}
		}
	}
	// Check string format (comma-separated)
	if rolesStr, ok := claims["roles"].(string); ok {
		for _, v := range strings.Split(rolesStr, ",") {
			if strings.TrimSpace(v) == role {
				return true
			}
		}
	}
	return false
}

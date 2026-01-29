package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-jwt/jwt/v5"
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
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:5173"},
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

		// Admin
		r.Get("/api/v1/admin/stats", Stats(deps))
		r.Get("/api/v1/admin/tenants", ListTenants(deps))
		r.Get("/api/v1/admin/tenants/{id}", GetTenant(deps))
		r.Post("/api/v1/admin/tenants", CreateTenant(deps))
		r.Delete("/api/v1/admin/tenants/{id}", DeleteTenant(deps))
		r.Get("/api/v1/admin/metrics", Metrics(deps))

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

// suppress unused import warning
var _ jwt.MapClaims

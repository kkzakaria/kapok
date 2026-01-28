package tenant

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/kapok/kapok/internal/auth"
	"github.com/rs/zerolog"
)

// RouterMiddleware handles tenant routing by extracting tenant_id from JWT
// and injecting it into the request context
type RouterMiddleware struct {
	logger zerolog.Logger
}

// NewRouterMiddleware creates a new tenant router middleware
func NewRouterMiddleware(logger zerolog.Logger) *RouterMiddleware {
	return &RouterMiddleware{
		logger: logger,
	}
}

// Middleware is the HTTP middleware function that extracts tenant_id
func (m *RouterMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract tenant_id from JWT claims (claims should be in context from auth middleware)
		claims, ok := ctx.Value(auth.JwtClaimsKey).(map[string]interface{})
		if !ok {
			m.logger.Error().Msg("JWT claims not found in context")
			http.Error(w, "Unauthorized: JWT claims missing", http.StatusUnauthorized)
			return
		}

		// Extract tenant_id from claims
		tenantIDInterface, exists := claims["tenant_id"]
		if !exists {
			m.logger.Error().Msg("tenant_id not found in JWT claims")
			http.Error(w, "Unauthorized: tenant_id missing from token", http.StatusUnauthorized)
			return
		}

		tenantID, ok := tenantIDInterface.(string)
		if !ok || tenantID == "" {
			m.logger.Error().Msg("tenant_id is not a valid string")
			http.Error(w, "Unauthorized: invalid tenant_id", http.StatusUnauthorized)
			return
		}

		// Validate tenant_id is a valid UUID
		if _, err := uuid.Parse(tenantID); err != nil {
			m.logger.Error().
				Err(err).
				Str("tenant_id", tenantID).
				Msg("tenant_id is not a valid UUID")
			http.Error(w, "Unauthorized: tenant_id must be a valid UUID", http.StatusUnauthorized)
			return
		}

		// Inject tenant_id into context
		ctx = WithTenantID(ctx, tenantID)

		// Log the request with tenant context
		m.logger.Info().
			Str("tenant_id", tenantID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("request routed to tenant")

		// Continue to next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SetTenantSessionVariable sets the PostgreSQL session variable for tenant isolation
// This should be called before any database operations
func SetTenantSessionVariable(db interface{}, tenantID string) error {
	// Type assertion for database interface
	type execer interface {
		Exec(query string, args ...interface{}) error
	}

	if dbExec, ok := db.(execer); ok {
		query := "SET LOCAL app.tenant_id = $1"
		if err := dbExec.Exec(query, tenantID); err != nil {
			return fmt.Errorf("failed to set tenant session variable: %w", err)
		}
		return nil
	}

	return fmt.Errorf("database does not support Exec interface")
}

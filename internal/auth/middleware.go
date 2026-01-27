package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
)

// AuthMiddleware handles JWT authentication for HTTP requests
type AuthMiddleware struct {
	jwtManager *JWTManager
	logger     zerolog.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtManager *JWTManager, logger zerolog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// Middleware is the HTTP middleware function that validates JWT tokens
func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.logger.Warn().Msg("missing Authorization header")
			http.Error(w, "Unauthorized: missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Check for Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.logger.Warn().Msg("invalid Authorization header format")
			http.Error(w, "Unauthorized: invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := m.jwtManager.ValidateToken(tokenString)
		if err != nil {
			m.logger.Warn().Err(err).Msg("invalid JWT token")
			http.Error(w, "Unauthorized: invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Extract user_id and tenant_id for logging
		userID, _ := ExtractUserID(claims)
		tenantID, _ := ExtractTenantID(claims)

		m.logger.Info().
			Str("user_id", userID).
			Str("tenant_id", tenantID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("authenticated request")

		// Add claims to context (convert jwt.MapClaims to map[string]interface{})
		claimsMap := map[string]interface{}(claims)
		ctx := context.WithValue(r.Context(), "jwt_claims", claimsMap)

		// Continue to next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthMiddleware is similar to Middleware but allows unauthenticated requests
func (m *AuthMiddleware) OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No auth header, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Try to validate token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			claims, err := m.jwtManager.ValidateToken(parts[1])
			if err == nil {
				// Valid token, add to context
				claimsMap := map[string]interface{}(claims)
				ctx := context.WithValue(r.Context(), "jwt_claims", claimsMap)
				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
	})
}

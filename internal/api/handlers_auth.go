package api

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/kapok/kapok/internal/auth"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login authenticates a user and returns a JWT token pair.
func Login(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := readJSON(r, &req); err != nil {
			errorResponse(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Email == "" || req.Password == "" {
			errorResponse(w, http.StatusBadRequest, "email and password are required")
			return
		}

		// Query user
		var user auth.User
		var rolesStr string
		var tenantID sql.NullString
		err := deps.DB.QueryRowContext(r.Context(),
			`SELECT id, email, password_hash, roles, tenant_id FROM users WHERE email = $1`,
			req.Email,
		).Scan(&user.ID, &user.Email, &user.PasswordHash, &rolesStr, &tenantID)
		if err != nil {
			deps.Logger.Warn().Str("email", req.Email).Msg("login failed: user not found")
			errorResponse(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		user.Roles = splitRoles(rolesStr)
		if tenantID.Valid {
			user.TenantID = tenantID.String
		}

		// Compare password
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			deps.Logger.Warn().Str("email", req.Email).Msg("login failed: wrong password")
			errorResponse(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		// Generate token pair
		tokenPair, err := deps.JWTManager.GenerateTokenPair(&user, user.Roles)
		if err != nil {
			deps.Logger.Error().Err(err).Msg("failed to generate tokens")
			errorResponse(w, http.StatusInternalServerError, "internal server error")
			return
		}

		deps.Logger.Info().Str("email", user.Email).Msg("user logged in")
		writeJSON(w, http.StatusOK, tokenPair)
	}
}

// Me returns the authenticated user's info from JWT claims.
func Me(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(claimsContextKey).(map[string]interface{})
		if !ok {
			errorResponse(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"id":        claims["sub"],
			"email":     claims["email"],
			"roles":     claims["roles"],
			"tenant_id": claims["tenant_id"],
		})
	}
}

func splitRoles(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

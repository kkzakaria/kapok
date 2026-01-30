package security

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

const (
	// CSRFTokenLength is the length of CSRF tokens in bytes
	CSRFTokenLength = 32

	// CSRFTokenExpiry is how long CSRF tokens are valid
	CSRFTokenExpiry = 24 * time.Hour

	// CSRFCookieName is the name of the CSRF cookie
	CSRFCookieName = "csrf_token"

	// CSRFHeaderName is the name of the CSRF header
	CSRFHeaderName = "X-CSRF-Token"
)

// CSRFProtection provides CSRF token generation and validation
type CSRFProtection struct {
	redis  *redis.Client
	logger zerolog.Logger
}

// NewCSRFProtection creates a new CSRF protection instance
func NewCSRFProtection(redisClient *redis.Client, logger zerolog.Logger) *CSRFProtection {
	return &CSRFProtection{
		redis:  redisClient,
		logger: logger,
	}
}

// GenerateToken generates a new CSRF token
func (csrf *CSRFProtection) GenerateToken(ctx context.Context, sessionID string) (string, error) {
	// Generate random token
	tokenBytes := make([]byte, CSRFTokenLength)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Store token in Redis with session ID as key
	key := fmt.Sprintf("csrf:%s", sessionID)
	err = csrf.redis.Set(ctx, key, token, CSRFTokenExpiry).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store CSRF token: %w", err)
	}

	return token, nil
}

// ValidateToken validates a CSRF token against the stored token
func (csrf *CSRFProtection) ValidateToken(ctx context.Context, sessionID, token string) error {
	if token == "" {
		return fmt.Errorf("CSRF token is empty")
	}

	// Retrieve stored token from Redis
	key := fmt.Sprintf("csrf:%s", sessionID)
	storedToken, err := csrf.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("CSRF token not found or expired")
		}
		return fmt.Errorf("failed to retrieve CSRF token: %w", err)
	}

	// Compare tokens using constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(token), []byte(storedToken)) != 1 {
		return fmt.Errorf("CSRF token mismatch")
	}

	return nil
}

// DeleteToken deletes a CSRF token (e.g., on logout)
func (csrf *CSRFProtection) DeleteToken(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("csrf:%s", sessionID)
	err := csrf.redis.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete CSRF token: %w", err)
	}
	return nil
}

// Middleware creates HTTP middleware for CSRF protection
func (csrf *CSRFProtection) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF check for GET, HEAD, OPTIONS (safe methods)
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Get session ID from context or cookie
		sessionID := csrf.getSessionID(r)
		if sessionID == "" {
			csrf.logger.Warn().Msg("CSRF check failed: no session ID")
			http.Error(w, "Forbidden: missing session", http.StatusForbidden)
			return
		}

		// Get CSRF token from header or form
		token := r.Header.Get(CSRFHeaderName)
		if token == "" {
			token = r.FormValue("csrf_token")
		}

		// Validate token
		err := csrf.ValidateToken(r.Context(), sessionID, token)
		if err != nil {
			csrf.logger.Warn().
				Err(err).
				Str("session_id", sessionID).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("CSRF validation failed")
			http.Error(w, "Forbidden: invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SetTokenCookie sets the CSRF token as a cookie readable by JavaScript.
// HttpOnly is intentionally false so the client can read the token and send it
// back in the X-CSRF-Token header (Double Submit Cookie pattern).
// SameSite=Strict prevents the cookie from being sent in cross-origin requests.
func (csrf *CSRFProtection) SetTokenCookie(w http.ResponseWriter, token string, secure bool) {
	cookie := &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(CSRFTokenExpiry.Seconds()),
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
}

// getSessionID extracts session ID from request
func (csrf *CSRFProtection) getSessionID(r *http.Request) string {
	// Try to get from JWT claims in context
	claims, ok := r.Context().Value("jwt_claims").(map[string]interface{})
	if ok {
		if sessionID, ok := claims["session_id"].(string); ok {
			return sessionID
		}
		if userID, ok := claims["sub"].(string); ok {
			return userID // Fallback to user ID
		}
	}

	// Try to get from session cookie
	cookie, err := r.Cookie("session_id")
	if err == nil {
		return cookie.Value
	}

	return ""
}

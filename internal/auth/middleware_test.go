package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware_Success(t *testing.T) {
	logger := zerolog.Nop()
	manager := NewJWTManager(testSecretKey)
	middleware := NewAuthMiddleware(manager, logger)

	// Create test user and token
	user := &User{
		ID:       "user-123",
		TenantID: "tenant-456",
		Email:    "test@example.com",
		Roles:    []string{"admin"},
	}
	token, err := manager.GenerateToken(user, []string{"read", "write"})
	require.NoError(t, err)

	// Create test handler
	var capturedClaims map[string]interface{}
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(JwtClaimsKey).(map[string]interface{})
		require.True(t, ok)
		capturedClaims = claims
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with middleware
	handler := middleware.Middleware(testHandler)

	// Create request with valid token
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotNil(t, capturedClaims)
	assert.Equal(t, user.ID, capturedClaims["sub"])
	assert.Equal(t, user.TenantID, capturedClaims["tenant_id"])
}

func TestAuthMiddleware_MissingAuthHeader(t *testing.T) {
	logger := zerolog.Nop()
	manager := NewJWTManager(testSecretKey)
	middleware := NewAuthMiddleware(manager, logger)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	handler := middleware.Middleware(testHandler)

	// Request without Authorization header
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "missing Authorization header")
}

func TestAuthMiddleware_InvalidHeaderFormat(t *testing.T) {
	logger := zerolog.Nop()
	manager := NewJWTManager(testSecretKey)
	middleware := NewAuthMiddleware(manager, logger)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	handler := middleware.Middleware(testHandler)

	tests := []struct {
		name   string
		header string
	}{
		{"no Bearer prefix", "some-token"},
		{"wrong prefix", "Basic some-token"},
		{"empty after Bearer", "Bearer "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tt.header)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusUnauthorized, rr.Code)
		})
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	logger := zerolog.Nop()
	manager := NewJWTManager(testSecretKey)
	middleware := NewAuthMiddleware(manager, logger)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	handler := middleware.Middleware(testHandler)

	// Request with invalid token
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid or expired token")
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	logger := zerolog.Nop()
	manager := NewJWTManager(testSecretKey)
	differentManager := NewJWTManager("different-secret")
	middleware := NewAuthMiddleware(manager, logger)

	user := &User{
		ID:       "user-123",
		TenantID: "tenant-456",
		Email:    "test@example.com",
		Roles:    []string{"admin"},
	}

	// Generate token with different secret (will fail validation)
	token, err := differentManager.GenerateToken(user, nil)
	require.NoError(t, err)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	handler := middleware.Middleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestOptionalAuthMiddleware_WithValidToken(t *testing.T) {
	logger := zerolog.Nop()
	manager := NewJWTManager(testSecretKey)
	middleware := NewAuthMiddleware(manager, logger)

	user := &User{
		ID:       "user-123",
		TenantID: "tenant-456",
		Email:    "test@example.com",
		Roles:    []string{"admin"},
	}
	token, err := manager.GenerateToken(user, nil)
	require.NoError(t, err)

	var hasClaims bool
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hasClaims = r.Context().Value(JwtClaimsKey).(map[string]interface{})
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.OptionalAuthMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, hasClaims, "claims should be in context")
}

func TestOptionalAuthMiddleware_WithoutToken(t *testing.T) {
	logger := zerolog.Nop()
	manager := NewJWTManager(testSecretKey)
	middleware := NewAuthMiddleware(manager, logger)

	var hasClaims bool
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hasClaims = r.Context().Value(JwtClaimsKey).(map[string]interface{})
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.OptionalAuthMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.False(t, hasClaims, "claims should not be in context")
}

func TestOptionalAuthMiddleware_WithInvalidToken(t *testing.T) {
	logger := zerolog.Nop()
	manager := NewJWTManager(testSecretKey)
	middleware := NewAuthMiddleware(manager, logger)

	var hasClaims bool
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hasClaims = r.Context().Value(JwtClaimsKey).(map[string]interface{})
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.OptionalAuthMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Should still succeed but without claims
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.False(t, hasClaims, "invalid token should not add claims to context")
}

func TestAuthMiddleware_ContextPropagation(t *testing.T) {
	logger := zerolog.Nop()
	manager := NewJWTManager(testSecretKey)
	middleware := NewAuthMiddleware(manager, logger)

	user := &User{
		ID:       "user-123",
		TenantID: "tenant-456",
		Email:    "test@example.com",
		Roles:    []string{"admin"},
	}
	token, err := manager.GenerateToken(user, []string{"read"})
	require.NoError(t, err)

	// Chain multiple handlers
	var receivedContext context.Context
	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContext = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Middleware(handler1)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotNil(t, receivedContext)
	
	claims, ok := receivedContext.Value(JwtClaimsKey).(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, user.ID, claims["sub"])
}

package tenant

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouterMiddleware_Success(t *testing.T) {
	logger := zerolog.Nop()
	middleware := NewRouterMiddleware(logger)

	// Create a test handler that checks if tenant_id is in context
	var capturedTenantID string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID, err := GetTenantID(r.Context())
		require.NoError(t, err)
		capturedTenantID = tenantID
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with middleware
	handler := middleware.Middleware(testHandler)

	// Create request with JWT claims in context
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), "jwt_claims", map[string]interface{}{
		"tenant_id": "test-tenant-123",
		"user_id":   "user-456",
	})
	req = req.WithContext(ctx)

	// Record response
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Assertions
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "test-tenant-123", capturedTenantID)
}

func TestRouterMiddleware_MissingJWTClaims(t *testing.T) {
	logger := zerolog.Nop()
	middleware := NewRouterMiddleware(logger)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	handler := middleware.Middleware(testHandler)

	// Request without JWT claims
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Should return 401
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "JWT claims missing")
}

func TestRouterMiddleware_MissingTenantID(t *testing.T) {
	logger := zerolog.Nop()
	middleware := NewRouterMiddleware(logger)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	handler := middleware.Middleware(testHandler)

	// Request with JWT claims but no tenant_id
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), "jwt_claims", map[string]interface{}{
		"user_id": "user-456",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Should return 401
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "tenant_id missing")
}

func TestRouterMiddleware_InvalidTenantIDType(t *testing.T) {
	logger := zerolog.Nop()
	middleware := NewRouterMiddleware(logger)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	handler := middleware.Middleware(testHandler)

	// Request with tenant_id as wrong type
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), "jwt_claims", map[string]interface{}{
		"tenant_id": 12345, // Wrong type - should be string
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Should return 401
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid tenant_id")
}

func TestRouterMiddleware_EmptyTenantID(t *testing.T) {
	logger := zerolog.Nop()
	middleware := NewRouterMiddleware(logger)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	handler := middleware.Middleware(testHandler)

	// Request with empty tenant_id
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), "jwt_claims", map[string]interface{}{
		"tenant_id": "",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Should return 401
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid tenant_id")
}

func TestRouterMiddleware_ContextPropagation(t *testing.T) {
	logger := zerolog.Nop()
	middleware := NewRouterMiddleware(logger)

	// Test that context is properly propagated through middleware chain
	var contextChecks []bool
	
	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First handler in chain
		contextChecks = append(contextChecks, HasTenantID(r.Context()))
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with middleware
	wrappedHandler := middleware.Middleware(handler1)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), "jwt_claims", map[string]interface{}{
		"tenant_id": "test-123",
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	// Verify context was available in handler
	assert.Equal(t, http.StatusOK, rr.Code)
	require.Len(t, contextChecks, 1)
	assert.True(t, contextChecks[0], "tenant_id should be in context")
}

func TestRouterMiddleware_MultipleRequests(t *testing.T) {
	logger := zerolog.Nop()
	middleware := NewRouterMiddleware(logger)

	var capturedTenantIDs []string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID, _ := GetTenantID(r.Context())
		capturedTenantIDs = append(capturedTenantIDs, tenantID)
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Middleware(testHandler)

	// Test multiple requests with different tenant IDs
	tenants := []string{"tenant-1", "tenant-2", "tenant-3"}
	for _, tenantID := range tenants {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx := context.WithValue(req.Context(), "jwt_claims", map[string]interface{}{
			"tenant_id": tenantID,
		})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	}

	// Verify all tenant IDs were captured correctly
	assert.Equal(t, tenants, capturedTenantIDs)
}

package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecretKey = "test-secret-key-for-jwt-signing"

func TestGenerateToken_Success(t *testing.T) {
	manager := NewJWTManager(testSecretKey)
	
	user := &User{
		ID:       "user-123",
		Email:    "test@example.com",
		TenantID: "tenant-456",
		Roles:    []string{"admin", "developer"},
	}
	permissions := []string{"read", "write", "delete"}

	token, err := manager.GenerateToken(user, permissions)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the token
	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)
	
	assert.Equal(t, user.ID, claims["sub"])
	assert.Equal(t, user.TenantID, claims["tenant_id"])
	assert.Equal(t, user.Email, claims["email"])
}

func TestGenerateRefreshToken_Success(t *testing.T) {
	manager := NewJWTManager(testSecretKey)

	token, err := manager.GenerateRefreshToken("user-123", "tenant-456")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the token
	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)
	
	assert.Equal(t, "user-123", claims["sub"])
	assert.Equal(t, "tenant-456", claims["tenant_id"])
	assert.Equal(t, "refresh", claims["type"])
}

func TestGenerateTokenPair_Success(t *testing.T) {
	manager := NewJWTManager(testSecretKey)
	
	user := &User{
		ID:       "user-123",
		Email:    "test@example.com",
		TenantID: "tenant-456",
		Roles:    []string{"admin"},
	}
	permissions := []string{"read", "write"}

	tokenPair, err := manager.GenerateTokenPair(user, permissions)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.Equal(t, int64(TokenExpiry.Seconds()), tokenPair.ExpiresIn)

	// Validate both tokens
	accessClaims, err := manager.ValidateToken(tokenPair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, user.ID, accessClaims["sub"])

	refreshClaims, err := manager.ValidateToken(tokenPair.RefreshToken)
	require.NoError(t, err)
	assert.Equal(t, user.ID, refreshClaims["sub"])
	assert.Equal(t, "refresh", refreshClaims["type"])
}

func TestValidateToken_ValidToken(t *testing.T) {
	manager := NewJWTManager(testSecretKey)
	
	user := &User{
		ID:       "user-123",
		TenantID: "tenant-456",
		Email:    "test@example.com",
		Roles:    []string{"viewer"},
	}

	token, err := manager.GenerateToken(user, nil)
	require.NoError(t, err)

	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, user.ID, claims["sub"])
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	manager := NewJWTManager(testSecretKey)
	differentManager := NewJWTManager("different-secret")
	
	user := &User{
		ID:       "user-123",
		TenantID: "tenant-456",
		Email:    "test@example.com",
		Roles:    []string{"admin"},
	}

	// Generate with one manager
	token, err := differentManager.GenerateToken(user, nil)
	require.NoError(t, err)

	// Validate with different manager (different secret)
	_, err = manager.ValidateToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse token")
}

func TestValidateToken_MalformedToken(t *testing.T) {
	manager := NewJWTManager(testSecretKey)

	_, err := manager.ValidateToken("not.a.valid.token")
	assert.Error(t, err)
}

func TestValidateToken_EmptyToken(t *testing.T) {
	manager := NewJWTManager(testSecretKey)

	_, err := manager.ValidateToken("")
	assert.Error(t, err)
}

func TestExtractTenantID_Success(t *testing.T) {
	manager := NewJWTManager(testSecretKey)
	
	user := &User{
		ID:       "user-123",
		TenantID: "tenant-456",
		Email:    "test@example.com",
		Roles:    []string{"admin"},
	}

	token, err := manager.GenerateToken(user, nil)
	require.NoError(t, err)

	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)

	tenantID, err := ExtractTenantID(claims)
	require.NoError(t, err)
	assert.Equal(t, "tenant-456", tenantID)
}

func TestExtractTenantID_Missing(t *testing.T) {
	claims := map[string]interface{}{
		"sub":   "user-123",
		"email": "test@example.com",
	}

	_, err := ExtractTenantID(claims)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant_id not found")
}

func TestExtractUserID_Success(t *testing.T) {
	manager := NewJWTManager(testSecretKey)
	
	user := &User{
		ID:       "user-123",
		TenantID: "tenant-456",
		Email:    "test@example.com",
		Roles:    []string{"admin"},
	}

	token, err := manager.GenerateToken(user, nil)
	require.NoError(t, err)

	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)

	userID, err := ExtractUserID(claims)
	require.NoError(t, err)
	assert.Equal(t, "user-123", userID)
}

func TestExtractUserID_Missing(t *testing.T) {
	claims := map[string]interface{}{
		"tenant_id": "tenant-456",
		"email":     "test@example.com",
	}

	_, err := ExtractUserID(claims)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_id (sub) not found")
}

func TestRefreshAccessToken_Success(t *testing.T) {
	manager := NewJWTManager(testSecretKey)
	
	user := &User{
		ID:       "user-123",
		TenantID: "tenant-456",
		Email:    "test@example.com",
		Roles:    []string{"admin"},
	}
	permissions := []string{"read", "write"}

	// Generate refresh token
	refreshToken, err := manager.GenerateRefreshToken(user.ID, user.TenantID)
	require.NoError(t, err)

	// Use refresh token to get new access token
	newAccessToken, err := manager.RefreshAccessToken(refreshToken, user, permissions)
	require.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)

	// Validate new access token
	claims, err := manager.ValidateToken(newAccessToken)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims["sub"])
}

func TestRefreshAccessToken_NotRefreshToken(t *testing.T) {
	manager := NewJWTManager(testSecretKey)
	
	user := &User{
		ID:       "user-123",
		TenantID: "tenant-456",
		Email:    "test@example.com",
		Roles:    []string{"admin"},
	}

	// Generate regular access token (not refresh)
	accessToken, err := manager.GenerateToken(user, nil)
	require.NoError(t, err)

	// Try to use access token as refresh token
	_, err = manager.RefreshAccessToken(accessToken, user, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a refresh token")
}

func TestRefreshAccessToken_WrongUserID(t *testing.T) {
	manager := NewJWTManager(testSecretKey)

	// Generate refresh token for user-123
	refreshToken, err := manager.GenerateRefreshToken("user-123", "tenant-456")
	require.NoError(t, err)

	// Try to refresh with different user
	differentUser := &User{
		ID:       "user-999", // Different user
		TenantID: "tenant-456",
		Email:    "other@example.com",
		Roles:    []string{"admin"},
	}

	_, err = manager.RefreshAccessToken(refreshToken, differentUser, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not match")
}

func TestTokenExpiry(t *testing.T) {
	manager := NewJWTManager(testSecretKey)
	
	user := &User{
		ID:       "user-123",
		TenantID: "tenant-456",
		Email:    "test@example.com",
		Roles:    []string{"admin"},
	}

	token, err := manager.GenerateToken(user, nil)
	require.NoError(t, err)

	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)

	// Check expiry is set correctly
	exp, ok := claims["exp"].(float64)
	require.True(t, ok)
	
	iat, ok := claims["iat"].(float64)
	require.True(t, ok)

	// Expiry should be ~24 hours after issued
	expiryDuration := time.Duration(exp-iat) * time.Second
	assert.InDelta(t, TokenExpiry.Seconds(), expiryDuration.Seconds(), 1.0)
}

func TestTokenClaims_AllFields(t *testing.T) {
	manager := NewJWTManager(testSecretKey)
	
	user := &User{
		ID:       "user-123",
		TenantID: "tenant-456",
		Email:    "test@example.com",
		Roles:    []string{"admin", "developer", "viewer"},
	}
	permissions := []string{"read", "write", "delete", "manage"}

	token, err := manager.GenerateToken(user, permissions)
	require.NoError(t, err)

	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)

	// Verify all expected fields
	assert.Equal(t, user.ID, claims["sub"])
	assert.Equal(t, user.TenantID, claims["tenant_id"])
	assert.Equal(t, user.Email, claims["email"])
	
	// Check roles array
	rolesInterface := claims["roles"].([]interface{})
	roles := make([]string, len(rolesInterface))
	for i, r := range rolesInterface {
		roles[i] = r.(string)
	}
	assert.Equal(t, user.Roles, roles)

	// Check permissions array
	permsInterface := claims["permissions"].([]interface{})
	perms := make([]string, len(permsInterface))
	for i, p := range permsInterface {
		perms[i] = p.(string)
	}
	assert.Equal(t, permissions, perms)

	// Check timestamps exist
	assert.NotNil(t, claims["iat"])
	assert.NotNil(t, claims["exp"])
}

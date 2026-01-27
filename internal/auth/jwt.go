package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// TokenExpiry is the default expiration time for access tokens
	TokenExpiry = 24 * time.Hour
	// RefreshTokenExpiry is the default expiration time for refresh tokens
	RefreshTokenExpiry = 7 * 24 * time.Hour
)

// JWTManager handles JWT token generation and validation
type JWTManager struct {
	secretKey []byte
}

// NewJWTManager creates a new JWT manager with the given secret key
func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{
		secretKey: []byte(secretKey),
	}
}

// GenerateToken creates a new JWT access token for a user
func (m *JWTManager) GenerateToken(user *User, permissions []string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(TokenExpiry)

	claims := jwt.MapClaims{
		"sub":         user.ID,
		"tenant_id":   user.TenantID,
		"email":       user.Email,
		"roles":       user.Roles,
		"permissions": permissions,
		"iat":         now.Unix(),
		"exp":         expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken creates a new refresh token
func (m *JWTManager) GenerateRefreshToken(userID, tenantID string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(RefreshTokenExpiry)

	claims := jwt.MapClaims{
		"sub":       userID,
		"tenant_id": tenantID,
		"type":      "refresh",
		"iat":       now.Unix(),
		"exp":       expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}

// GenerateTokenPair creates both access and refresh tokens
func (m *JWTManager) GenerateTokenPair(user *User, permissions []string) (*TokenPair, error) {
	accessToken, err := m.GenerateToken(user, permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := m.GenerateRefreshToken(user.ID, user.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(TokenExpiry.Seconds()),
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func (m *JWTManager) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// ExtractTenantID extracts the tenant ID from JWT claims
func ExtractTenantID(claims jwt.MapClaims) (string, error) {
	tenantID, ok := claims["tenant_id"].(string)
	if !ok || tenantID == "" {
		return "", fmt.Errorf("tenant_id not found in token claims")
	}
	return tenantID, nil
}

// ExtractUserID extracts the user ID from JWT claims
func ExtractUserID(claims jwt.MapClaims) (string, error) {
	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("user_id (sub) not found in token claims")
	}
	return userID, nil
}

// RefreshAccessToken validates a refresh token and generates a new access token
func (m *JWTManager) RefreshAccessToken(refreshTokenString string, user *User, permissions []string) (string, error) {
	// Validate refresh token
	claims, err := m.ValidateToken(refreshTokenString)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Verify it's a refresh token
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return "", fmt.Errorf("token is not a refresh token")
	}

	// Verify user ID matches
	tokenUserID, err := ExtractUserID(claims)
	if err != nil {
		return "", err
	}
	if tokenUserID != user.ID {
		return "", fmt.Errorf("token user_id does not match")
	}

	// Generate new access token
	return m.GenerateToken(user, permissions)
}

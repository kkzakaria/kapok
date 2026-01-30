package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost is the cost factor for bcrypt hashing (higher = more secure but slower)
	// Cost of 12 means 2^12 iterations (~250ms on modern CPU)
	BcryptCost = 12

	// PasswordResetTokenLength is the length of password reset tokens in bytes
	PasswordResetTokenLength = 32

	// PasswordResetExpiry is how long password reset tokens are valid
	PasswordResetExpiry = 1 * time.Hour
)

// PasswordManager handles secure password operations
type PasswordManager struct {
	validator *InputValidator
}

// NewPasswordManager creates a new password manager
func NewPasswordManager() *PasswordManager {
	return &PasswordManager{
		validator: NewInputValidator(),
	}
}

// HashPassword hashes a password using bcrypt
func (pm *PasswordManager) HashPassword(password string) (string, error) {
	// Validate password strength first
	if err := pm.validator.ValidatePassword(password); err != nil {
		return "", fmt.Errorf("password validation failed: %w", err)
	}

	// Check against common passwords
	if pm.validator.IsCommonPassword(password) {
		return "", fmt.Errorf("password is too common, please choose a stronger password")
	}

	// Hash password
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyPassword compares a plaintext password with a hashed password
func (pm *PasswordManager) VerifyPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("password verification failed: %w", err)
	}
	return nil
}

// GeneratePasswordResetToken generates a cryptographically secure random token
func (pm *PasswordManager) GeneratePasswordResetToken() (string, error) {
	// Generate random bytes
	tokenBytes := make([]byte, PasswordResetTokenLength)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// Encode to base64 for URL safety
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	return token, nil
}

// PasswordResetToken represents a password reset request
type PasswordResetToken struct {
	Token     string
	UserID    string
	TenantID  string
	ExpiresAt time.Time
	Used      bool
}

// IsExpired checks if the reset token has expired
func (prt *PasswordResetToken) IsExpired() bool {
	return time.Now().After(prt.ExpiresAt)
}

// IsValid checks if the token is still valid (not expired and not used)
func (prt *PasswordResetToken) IsValid() bool {
	return !prt.Used && !prt.IsExpired()
}

// GenerateSecureRandomPassword generates a cryptographically secure random password
func (pm *PasswordManager) GenerateSecureRandomPassword(length int) (string, error) {
	if length < 12 {
		length = 12
	}

	const (
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		digits    = "0123456789"
		special   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	)

	allChars := uppercase + lowercase + digits + special
	charsetLen := big.NewInt(int64(len(allChars)))

	// Generate password characters using unbiased random selection
	password := make([]byte, length)
	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random password: %w", err)
		}
		password[i] = allChars[idx.Int64()]
	}

	// Ensure password meets complexity requirements
	// by forcing at least one character from each category
	if length >= 4 {
		categories := []string{uppercase, lowercase, digits, special}
		for i, cat := range categories {
			idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(cat))))
			if err != nil {
				return "", fmt.Errorf("failed to generate random password: %w", err)
			}
			password[i] = cat[idx.Int64()]
		}
	}

	return string(password), nil
}

package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// EncryptionManager handles encryption and decryption operations
type EncryptionManager struct {
	key []byte
}

// NewEncryptionManager creates a new encryption manager with AES-256
// Key must be 32 bytes for AES-256
func NewEncryptionManager(key []byte) (*EncryptionManager, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes for AES-256, got %d", len(key))
	}

	return &EncryptionManager{
		key: key,
	}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM
func (em *EncryptionManager) Encrypt(plaintext []byte) (string, error) {
	// Create AES cipher
	block, err := aes.NewCipher(em.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode (provides both encryption and authentication)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Encode to base64 for storage/transmission
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	return encoded, nil
}

// Decrypt decrypts ciphertext encrypted with Encrypt
func (em *EncryptionManager) Decrypt(encoded string) ([]byte, error) {
	// Decode from base64
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(em.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Check minimum length
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt and verify
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// EncryptString is a convenience method for encrypting strings
func (em *EncryptionManager) EncryptString(plaintext string) (string, error) {
	return em.Encrypt([]byte(plaintext))
}

// DecryptString is a convenience method for decrypting to strings
func (em *EncryptionManager) DecryptString(encoded string) (string, error) {
	plaintext, err := em.Decrypt(encoded)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// GenerateEncryptionKey generates a cryptographically secure 32-byte key for AES-256
func GenerateEncryptionKey() ([]byte, error) {
	key := make([]byte, 32) // 256 bits
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}
	return key, nil
}

// GenerateEncryptionKeyBase64 generates a base64-encoded encryption key
func GenerateEncryptionKeyBase64() (string, error) {
	key, err := GenerateEncryptionKey()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

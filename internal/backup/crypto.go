package backup

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

// Encrypt encrypts data from src and writes it to dst using AES-256-GCM.
// The key must be 32 bytes. A random nonce is prepended to the output.
func Encrypt(dst io.Writer, src io.Reader, key []byte) error {
	if len(key) != 32 {
		return fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
	}

	plaintext, err := io.ReadAll(src)
	if err != nil {
		return fmt.Errorf("failed to read plaintext: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	if _, err := dst.Write(ciphertext); err != nil {
		return fmt.Errorf("failed to write ciphertext: %w", err)
	}
	return nil
}

// Decrypt decrypts data from src and writes it to dst using AES-256-GCM.
// Expects the nonce prepended to the ciphertext.
func Decrypt(dst io.Writer, src io.Reader, key []byte) error {
	if len(key) != 32 {
		return fmt.Errorf("decryption key must be 32 bytes, got %d", len(key))
	}

	ciphertext, err := io.ReadAll(src)
	if err != nil {
		return fmt.Errorf("failed to read ciphertext: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("failed to decrypt: %w", err)
	}

	if _, err := dst.Write(plaintext); err != nil {
		return fmt.Errorf("failed to write plaintext: %w", err)
	}
	return nil
}

// Checksum computes the SHA-256 hex checksum of data from r.
func Checksum(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("failed to compute checksum: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

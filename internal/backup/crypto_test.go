package backup

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
	}{
		{"empty", ""},
		{"short", "hello world"},
		{"medium", "The quick brown fox jumps over the lazy dog. " + "Repeated. "},
		{"binary-like", string([]byte{0, 1, 2, 255, 254, 253})},
	}

	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var encrypted bytes.Buffer
			err := Encrypt(&encrypted, bytes.NewReader([]byte(tt.plaintext)), key)
			require.NoError(t, err)

			var decrypted bytes.Buffer
			err = Decrypt(&decrypted, bytes.NewReader(encrypted.Bytes()), key)
			require.NoError(t, err)

			assert.Equal(t, tt.plaintext, decrypted.String())
		})
	}
}

func TestEncryptInvalidKeyLength(t *testing.T) {
	var buf bytes.Buffer
	err := Encrypt(&buf, bytes.NewReader([]byte("test")), []byte("short"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "32 bytes")
}

func TestDecryptInvalidKeyLength(t *testing.T) {
	var buf bytes.Buffer
	err := Decrypt(&buf, bytes.NewReader([]byte("test")), []byte("short"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "32 bytes")
}

func TestDecryptTamperedData(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	var encrypted bytes.Buffer
	require.NoError(t, Encrypt(&encrypted, bytes.NewReader([]byte("secret")), key))

	// Tamper with the ciphertext
	data := encrypted.Bytes()
	data[len(data)-1] ^= 0xff

	var decrypted bytes.Buffer
	err := Decrypt(&decrypted, bytes.NewReader(data), key)
	assert.Error(t, err)
}

func TestChecksumDeterministic(t *testing.T) {
	data := []byte("deterministic checksum test")
	c1, err := Checksum(bytes.NewReader(data))
	require.NoError(t, err)
	c2, err := Checksum(bytes.NewReader(data))
	require.NoError(t, err)
	assert.Equal(t, c1, c2)
	assert.Len(t, c1, 64) // SHA-256 hex = 64 chars
}

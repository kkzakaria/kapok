package security

import (
	"crypto/tls"
	"fmt"
)

// TLSConfig provides secure TLS configuration
type TLSConfig struct {
	MinVersion               uint16
	MaxVersion               uint16
	CipherSuites             []uint16
	PreferServerCipherSuites bool
	CurvePreferences         []tls.CurveID
}

// NewSecureTLSConfig creates a secure TLS 1.3+ configuration
func NewSecureTLSConfig() *TLSConfig {
	return &TLSConfig{
		MinVersion: tls.VersionTLS13, // TLS 1.3 minimum
		MaxVersion: tls.VersionTLS13, // TLS 1.3 maximum (most secure)
		// TLS 1.3 cipher suites (order matters - most secure first)
		CipherSuites: []uint16{
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
		},
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.X25519,    // Modern, fast, secure
			tls.CurveP256, // NIST P-256
		},
	}
}

// NewCompatibleTLSConfig creates a TLS configuration compatible with TLS 1.2+
// Use only if TLS 1.3 is not supported by all clients
func NewCompatibleTLSConfig() *TLSConfig {
	return &TLSConfig{
		MinVersion: tls.VersionTLS12, // TLS 1.2 minimum
		MaxVersion: tls.VersionTLS13, // TLS 1.3 preferred
		// Secure cipher suites for both TLS 1.2 and 1.3
		CipherSuites: []uint16{
			// TLS 1.3 cipher suites
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
			// TLS 1.2 cipher suites
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},
	}
}

// ToNativeTLSConfig converts to crypto/tls.Config
func (tc *TLSConfig) ToNativeTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion:               tc.MinVersion,
		MaxVersion:               tc.MaxVersion,
		CipherSuites:             tc.CipherSuites,
		PreferServerCipherSuites: tc.PreferServerCipherSuites,
		CurvePreferences:         tc.CurvePreferences,
	}
}

// ValidateTLSConfig validates TLS configuration security
func ValidateTLSConfig(config *tls.Config) error {
	if config.MinVersion < tls.VersionTLS12 {
		return fmt.Errorf("TLS version too old: minimum version must be TLS 1.2 or higher")
	}

	// Warn if not using TLS 1.3
	if config.MinVersion < tls.VersionTLS13 {
		// This is acceptable but not ideal
	}

	// Check that no weak cipher suites are enabled
	weakCiphers := []uint16{
		tls.TLS_RSA_WITH_RC4_128_SHA,
		tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
		tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	}

	for _, weak := range weakCiphers {
		for _, configured := range config.CipherSuites {
			if weak == configured {
				return fmt.Errorf("weak cipher suite detected: %x", weak)
			}
		}
	}

	return nil
}

// GetTLSVersionName returns the human-readable name of a TLS version
func GetTLSVersionName(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return "Unknown"
	}
}

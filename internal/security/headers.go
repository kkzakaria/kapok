package security

import (
	"net/http"
)

// SecurityHeadersMiddleware adds comprehensive security headers to all responses
type SecurityHeadersMiddleware struct {
	config SecurityHeadersConfig
}

// SecurityHeadersConfig configures security headers
type SecurityHeadersConfig struct {
	// ContentSecurityPolicy defines CSP directives
	ContentSecurityPolicy string
	// EnableHSTS enables HTTP Strict Transport Security
	EnableHSTS bool
	// HSTSMaxAge is the max-age for HSTS in seconds
	HSTSMaxAge int
	// HSTSIncludeSubDomains includes subdomains in HSTS
	HSTSIncludeSubDomains bool
	// HSTSPreload enables HSTS preloading
	HSTSPreload bool
}

// DefaultSecurityHeadersConfig returns the default security headers configuration
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		ContentSecurityPolicy: "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none';",
		EnableHSTS:           true,
		HSTSMaxAge:           31536000, // 1 year
		HSTSIncludeSubDomains: true,
		HSTSPreload:          false,
	}
}

// NewSecurityHeadersMiddleware creates a new security headers middleware
func NewSecurityHeadersMiddleware(config SecurityHeadersConfig) *SecurityHeadersMiddleware {
	return &SecurityHeadersMiddleware{
		config: config,
	}
}

// Middleware is the HTTP middleware function that adds security headers
func (m *SecurityHeadersMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Content Security Policy
		if m.config.ContentSecurityPolicy != "" {
			w.Header().Set("Content-Security-Policy", m.config.ContentSecurityPolicy)
		}

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Enable XSS protection (for older browsers)
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy (formerly Feature-Policy)
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=()")

		// HTTP Strict Transport Security
		if m.config.EnableHSTS {
			hstsValue := ""
			if m.config.HSTSMaxAge > 0 {
				hstsValue = "max-age=" + string(rune(m.config.HSTSMaxAge))
			} else {
				hstsValue = "max-age=31536000" // Default to 1 year
			}

			if m.config.HSTSIncludeSubDomains {
				hstsValue += "; includeSubDomains"
			}

			if m.config.HSTSPreload {
				hstsValue += "; preload"
			}

			w.Header().Set("Strict-Transport-Security", hstsValue)
		}

		// Prevent browsers from caching sensitive data
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// Indicate this is not a search engine indexable page (for admin panels)
		w.Header().Set("X-Robots-Tag", "noindex, nofollow")

		next.ServeHTTP(w, r)
	})
}

// CORSConfig defines CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns a secure default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{"https://*"}, // Only HTTPS origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
type CORSMiddleware struct {
	config CORSConfig
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(config CORSConfig) *CORSMiddleware {
	return &CORSMiddleware{
		config: config,
	}
}

// Middleware is the HTTP middleware function for CORS
func (m *CORSMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		if origin != "" && m.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)

			if m.config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			// Set allowed methods
			if len(m.config.AllowedMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", joinStrings(m.config.AllowedMethods, ", "))
			}

			// Set allowed headers
			if len(m.config.AllowedHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", joinStrings(m.config.AllowedHeaders, ", "))
			}

			// Set exposed headers
			if len(m.config.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", joinStrings(m.config.ExposedHeaders, ", "))
			}

			// Set max age
			if m.config.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", string(rune(m.config.MaxAge)))
			}

			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isOriginAllowed checks if an origin is in the allowed list
func (m *CORSMiddleware) isOriginAllowed(origin string) bool {
	for _, allowed := range m.config.AllowedOrigins {
		if allowed == "*" {
			return true
		}
		if origin == allowed {
			return true
		}
		// Simple wildcard matching for subdomains
		if len(allowed) > 0 && allowed[0] == '*' {
			suffix := allowed[1:]
			if len(origin) >= len(suffix) && origin[len(origin)-len(suffix):] == suffix {
				return true
			}
		}
	}
	return false
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

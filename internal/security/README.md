# Security Package

This package provides comprehensive security features for the Kapok platform, implementing OWASP Top 10 mitigations and compliance with ISO 27001, SOC 2, GDPR, and HIPAA standards.

## Features

### Input Validation (`validation.go`)
- Email validation
- URL validation
- HTML sanitization and stripping
- XSS detection
- SQL injection detection
- Password strength validation
- Filename sanitization
- Alphanumeric validation

### Password Management (`password.go`)
- Bcrypt password hashing (cost factor 12)
- Password strength validation
- Common password checking
- Password reset token generation
- Secure random password generation

### Rate Limiting (`rate_limiter.go`)
- Distributed rate limiting using Redis
- Per-tenant and per-IP rate limits
- Configurable limits per endpoint type
- Automatic IP blocking on excessive requests
- Rate limit headers (X-RateLimit-*)

### Audit Logging (`audit.go`)
- Immutable audit trail
- HMAC signature for tamper detection
- Comprehensive event logging
- Query and reporting capabilities
- Integrity verification

### Encryption (`encryption.go`)
- AES-256-GCM encryption/decryption
- Cryptographically secure key generation
- Base64 encoding for storage

### TLS Configuration (`tls_config.go`)
- Secure TLS 1.3+ configuration
- Strong cipher suite selection
- TLS configuration validation

### Security Headers (`headers.go`)
- Content Security Policy
- X-Frame-Options
- X-Content-Type-Options
- HSTS (HTTP Strict Transport Security)
- Referrer Policy
- Permissions Policy
- CORS support

### CSRF Protection (`csrf.go`)
- CSRF token generation and validation
- Redis-based token storage
- HTTP middleware integration

### Multi-Factor Authentication (`mfa_totp.go`, `mfa_webauthn.go`)
- TOTP (Time-based One-Time Password)
- WebAuthn / FIDO2 support
- Backup codes generation
- QR code generation for TOTP setup

## Usage Examples

### Input Validation

```go
import "github.com/kapok/kapok/internal/security"

validator := security.NewInputValidator()

// Validate email
if err := validator.ValidateEmail(email); err != nil {
    return fmt.Errorf("invalid email: %w", err)
}

// Sanitize HTML
safe := validator.SanitizeHTML(userInput)

// Check for XSS
if validator.ContainsXSS(input) {
    return errors.New("potentially malicious input")
}
```

### Password Management

```go
pm := security.NewPasswordManager()

// Hash password
hashedPassword, err := pm.HashPassword(password)
if err != nil {
    return err
}

// Verify password
err = pm.VerifyPassword(hashedPassword, providedPassword)
if err != nil {
    return errors.New("invalid password")
}
```

### Rate Limiting

```go
rateLimiter := security.NewRateLimiter(redisClient, logger)

// Apply middleware
router.Use(rateLimiter.RateLimitMiddleware(security.DefaultRateLimitConfig()))

// For authentication endpoints (stricter limits)
authRouter.Use(rateLimiter.RateLimitMiddleware(security.AuthRateLimitConfig()))
```

### Audit Logging

```go
auditLogger := security.NewAuditLogger(db, secretKey, logger)

// Initialize audit table
err := auditLogger.InitializeAuditTable(ctx)

// Log security event
err = auditLogger.LogLoginSuccess(ctx, userID, tenantID, ipAddress, userAgent)

// Query audit logs
filters := security.AuditQueryFilters{
    TenantID: tenantID,
    StartTime: startTime,
    EndTime: endTime,
}
events, err := auditLogger.QueryAuditLogs(ctx, filters)
```

### Encryption

```go
// Create encryption manager (32-byte key for AES-256)
em, err := security.NewEncryptionManager(encryptionKey)

// Encrypt
encrypted, err := em.EncryptString("sensitive data")

// Decrypt
decrypted, err := em.DecryptString(encrypted)
```

### MFA - TOTP

```go
mfaManager := security.NewMFAManager("Kapok")

// Setup TOTP for user
setup, err := mfaManager.GenerateTOTPSecret(userEmail)
// Display setup.QRCodeURL to user

// Verify TOTP code
valid, err := mfaManager.VerifyTOTP(setup.Secret, userProvidedCode)
```

### MFA - WebAuthn

```go
webAuthnMgr, err := security.NewWebAuthnManager(
    "Kapok",
    "kapok.io",
    "https://kapok.io",
)

// Begin registration
user := &security.WebAuthnUser{
    ID: []byte(userID),
    Name: email,
    DisplayName: displayName,
}
options, session, err := webAuthnMgr.BeginRegistration(user)

// Finish registration (after user response)
credential, err := webAuthnMgr.FinishRegistration(user, session, response)
```

### Security Headers

```go
securityHeaders := security.NewSecurityHeadersMiddleware(
    security.DefaultSecurityHeadersConfig(),
)
router.Use(securityHeaders.Middleware)
```

### CSRF Protection

```go
csrfProtection := security.NewCSRFProtection(redisClient, logger)

// Apply middleware (protects POST/PUT/DELETE)
router.Use(csrfProtection.Middleware)

// Generate token for user session
token, err := csrfProtection.GenerateToken(ctx, sessionID)

// Set as cookie
csrfProtection.SetTokenCookie(w, token, true)
```

### TLS Configuration

```go
// Secure TLS 1.3 configuration
tlsConfig := security.NewSecureTLSConfig()

server := &http.Server{
    TLSConfig: tlsConfig.ToNativeTLSConfig(),
}

// Compatible TLS 1.2+ configuration
compatConfig := security.NewCompatibleTLSConfig()
```

## Security Best Practices

1. **Always validate input** before processing
2. **Use parameterized queries** to prevent SQL injection
3. **Hash passwords** with bcrypt (never store plaintext)
4. **Encrypt sensitive data** at rest (AES-256)
5. **Use TLS 1.3** for all network communication
6. **Log security events** to immutable audit trail
7. **Implement rate limiting** on all endpoints
8. **Enable CSRF protection** for state-changing operations
9. **Set security headers** on all HTTP responses
10. **Use MFA** for sensitive operations

## Testing

Run tests with:

```bash
go test ./internal/security/...
```

Run with race detection:

```bash
go test -race ./internal/security/...
```

## Compliance

This package implements controls required for:

- **ISO 27001**: Information security management
- **SOC 2**: Trust service criteria (Security, Availability, Confidentiality)
- **GDPR**: EU data protection regulation
- **HIPAA**: Healthcare data protection
- **OWASP Top 10**: Web application security risks

See compliance documentation in `/docs/compliance/`.

## Dependencies

- `golang.org/x/crypto/bcrypt` - Password hashing
- `github.com/golang-jwt/jwt/v5` - JWT tokens
- `github.com/redis/go-redis/v9` - Redis client for rate limiting and CSRF
- `github.com/pquerna/otp` - TOTP implementation
- `github.com/go-webauthn/webauthn` - WebAuthn/FIDO2 support
- `github.com/rs/zerolog` - Structured logging

## Contact

**Security Team**: security@kapok.io

## License

See LICENSE file in project root.

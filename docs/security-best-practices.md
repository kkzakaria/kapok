# Security Best Practices for Kapok Development

## Overview

This document outlines security best practices for developers working on the Kapok platform.

## Secure Coding Guidelines

### 1. Input Validation

**Always validate and sanitize user input:**

```go
// GOOD: Validate email before using
validator := security.NewInputValidator()
if err := validator.ValidateEmail(email); err != nil {
    return fmt.Errorf("invalid email: %w", err)
}

// BAD: Trust user input
email := r.FormValue("email") // No validation!
```

**Check for XSS and SQL injection patterns:**

```go
// GOOD: Check for malicious content
if validator.ContainsXSS(input) {
    return errors.New("input contains potentially malicious content")
}

// GOOD: Use parameterized queries
db.Query("SELECT * FROM users WHERE email = $1", email)

// BAD: String concatenation in SQL
query := "SELECT * FROM users WHERE email = '" + email + "'" // SQL injection!
```

### 2. Authentication and Authorization

**Always use JWT middleware:**

```go
// GOOD: Require authentication
router.Use(authMiddleware.Middleware)

// BAD: No authentication check
router.HandleFunc("/api/users", handleUsers) // Anyone can access!
```

**Check permissions before data access:**

```go
// GOOD: Verify tenant access
tenantID, err := auth.ExtractTenantID(claims)
if err != nil {
    return http.StatusUnauthorized
}

// Verify user has permission for this tenant
if !hasPermission(userID, tenantID, "read") {
    return http.StatusForbidden
}

// BAD: Trust user-provided tenant ID
tenantID := r.URL.Query().Get("tenant_id") // User can access any tenant!
```

### 3. Password Handling

**Never store plaintext passwords:**

```go
// GOOD: Hash passwords with bcrypt
pm := security.NewPasswordManager()
hashedPassword, err := pm.HashPassword(password)

// Store hashedPassword in database

// GOOD: Verify password
err = pm.VerifyPassword(hashedPassword, providedPassword)

// BAD: Store plaintext password
user.Password = password // NEVER DO THIS!
```

**Enforce strong password policy:**

```go
// GOOD: Validate password strength
validator := security.NewInputValidator()
if err := validator.ValidatePassword(password); err != nil {
    return errors.New("password does not meet security requirements")
}

// Check against common passwords
if validator.IsCommonPassword(password) {
    return errors.New("password is too common")
}
```

### 4. Encryption

**Encrypt sensitive data at rest:**

```go
// GOOD: Encrypt sensitive data
em, err := security.NewEncryptionManager(encryptionKey)
encryptedData, err := em.EncryptString(sensitiveData)

// Store encryptedData

// Later, decrypt when needed
decryptedData, err := em.DecryptString(encryptedData)

// BAD: Store sensitive data in plaintext
db.Exec("INSERT INTO secrets (api_key) VALUES ($1)", apiKey) // Plaintext!
```

**Use TLS for all network communication:**

```go
// GOOD: Configure secure TLS
tlsConfig := security.NewSecureTLSConfig()
server := &http.Server{
    TLSConfig: tlsConfig.ToNativeTLSConfig(),
}
server.ListenAndServeTLS(certFile, keyFile)

// BAD: No TLS
server.ListenAndServe(":8080", nil) // Unencrypted HTTP!
```

### 5. Logging and Audit Trails

**Log security events:**

```go
// GOOD: Log security events to audit trail
auditLogger.LogLoginSuccess(ctx, userID, tenantID, ipAddress, userAgent)

// Log failed attempts
auditLogger.LogLoginFailure(ctx, email, ipAddress, userAgent, "invalid password")

// GOOD: Structured logging with context
logger.Info().
    Str("user_id", userID).
    Str("tenant_id", tenantID).
    Str("action", "data_access").
    Msg("user accessed customer data")

// BAD: Print statements
fmt.Println("User logged in:", userID) // Lost in stdout, no audit trail
```

**Never log sensitive data:**

```go
// GOOD: Redact sensitive data
logger.Info().
    Str("user_id", userID).
    Str("password", "[REDACTED]"). // Don't log passwords!
    Msg("login attempt")

// BAD: Log passwords
logger.Info().
    Str("password", password). // NEVER LOG PASSWORDS!
    Msg("login attempt")
```

### 6. Rate Limiting

**Implement rate limiting on all endpoints:**

```go
// GOOD: Apply rate limiting middleware
rateLimiter := security.NewRateLimiter(redisClient, logger)

// Normal endpoints
router.Use(rateLimiter.RateLimitMiddleware(security.DefaultRateLimitConfig()))

// Auth endpoints (stricter limits)
authRouter.Use(rateLimiter.RateLimitMiddleware(security.AuthRateLimitConfig()))

// BAD: No rate limiting
router.HandleFunc("/api/login", handleLogin) // Brute force vulnerable!
```

### 7. CSRF Protection

**Enable CSRF protection for state-changing operations:**

```go
// GOOD: CSRF middleware
csrfProtection := security.NewCSRFProtection(redisClient, logger)
router.Use(csrfProtection.Middleware)

// BAD: No CSRF protection on POST/PUT/DELETE
router.HandleFunc("/api/users", handleCreateUser) // CSRF vulnerable!
```

### 8. Security Headers

**Always set security headers:**

```go
// GOOD: Security headers middleware
securityHeaders := security.NewSecurityHeadersMiddleware(
    security.DefaultSecurityHeadersConfig(),
)
router.Use(securityHeaders.Middleware)

// BAD: No security headers
// Missing CSP, X-Frame-Options, etc.
```

## Common Security Anti-Patterns

### ❌ DON'T: Trust User Input

```go
// BAD
filename := r.FormValue("filename")
file, _ := os.Open(filename) // Path traversal vulnerability!
```

```go
// GOOD
filename := r.FormValue("filename")
validator := security.NewInputValidator()
sanitized := validator.SanitizeFilename(filename)
if strings.Contains(sanitized, "..") {
    return errors.New("invalid filename")
}
```

### ❌ DON'T: Use String Concatenation for SQL

```go
// BAD
query := "SELECT * FROM users WHERE id = " + userID // SQL injection!
```

```go
// GOOD
query := "SELECT * FROM users WHERE id = $1"
db.Query(query, userID)
```

### ❌ DON'T: Hardcode Secrets

```go
// BAD
const jwtSecret = "my-secret-key-123" // Hardcoded!
```

```go
// GOOD
jwtSecret := os.Getenv("KAPOK_JWT_SECRET")
if jwtSecret == "" {
    return errors.New("JWT_SECRET not configured")
}
```

### ❌ DON'T: Disable Security Features

```go
// BAD
tlsConfig := &tls.Config{
    InsecureSkipVerify: true, // NEVER DO THIS!
}
```

### ❌ DON'T: Return Detailed Error Messages to Users

```go
// BAD
http.Error(w, "Database error: "+err.Error(), 500) // Leaks internal details!
```

```go
// GOOD
logger.Error().Err(err).Msg("database query failed")
http.Error(w, "Internal server error", 500)
```

## Security Checklist for Pull Requests

Before submitting a PR, verify:

- [ ] All user input is validated and sanitized
- [ ] No SQL injection vulnerabilities (use parameterized queries)
- [ ] No XSS vulnerabilities (sanitize HTML output)
- [ ] Authentication required for protected endpoints
- [ ] Authorization checks for tenant data access
- [ ] Passwords are hashed with bcrypt
- [ ] Sensitive data encrypted at rest
- [ ] TLS used for network communication
- [ ] Security events logged to audit trail
- [ ] No sensitive data in logs (passwords, keys, etc.)
- [ ] Rate limiting applied to endpoints
- [ ] CSRF protection for state-changing operations
- [ ] No hardcoded secrets or credentials
- [ ] Error messages don't leak sensitive information
- [ ] Dependencies scanned for vulnerabilities

## Security Testing

### Unit Tests

**Test input validation:**

```go
func TestValidateEmail(t *testing.T) {
    validator := security.NewInputValidator()

    // Test valid email
    err := validator.ValidateEmail("user@example.com")
    assert.NoError(t, err)

    // Test SQL injection attempt
    err = validator.ValidateEmail("user'; DROP TABLE users--")
    assert.Error(t, err)
}
```

**Test authentication:**

```go
func TestAuthenticationRequired(t *testing.T) {
    req := httptest.NewRequest("GET", "/api/users", nil)
    // No Authorization header

    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
```

### Integration Tests

**Test audit logging:**

```go
func TestAuditLogging(t *testing.T) {
    // Perform action
    user.Login(email, password)

    // Verify audit log created
    logs, _ := auditLogger.QueryAuditLogs(ctx, filters)
    assert.Len(t, logs, 1)
    assert.Equal(t, security.EventLoginSuccess, logs[0].EventType)
}
```

**Test rate limiting:**

```go
func TestRateLimiting(t *testing.T) {
    // Make 101 requests (limit is 100)
    for i := 0; i < 101; i++ {
        resp := makeRequest()
    }

    // 101st request should be rate limited
    assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}
```

## Security Tools

### Static Analysis

```bash
# Run gosec
gosec ./...

# Run semgrep
semgrep --config=p/security-audit ./...
```

### Dependency Scanning

```bash
# Check for vulnerable dependencies
govulncheck ./...

# Snyk scanning
snyk test
```

### Container Scanning

```bash
# Scan Docker images
trivy image kapok-gateway:latest
```

## Incident Response

### If you discover a security vulnerability:

1. **DO NOT** create a public GitHub issue
2. **Immediately** notify the security team: security@kapok.io
3. **Include** details: affected component, reproduction steps, impact assessment
4. **Wait** for security team response before disclosing

### If you suspect a security incident:

1. **Document** what you observed
2. **Notify** security team immediately
3. **Preserve** logs and evidence
4. **Do not** attempt to fix without approval

## Security Resources

- **OWASP Top 10**: https://owasp.org/www-project-top-ten/
- **OWASP Cheat Sheets**: https://cheatsheetseries.owasp.org/
- **Go Security**: https://go.dev/security/
- **CWE List**: https://cwe.mitre.org/
- **NIST Guidelines**: https://csrc.nist.gov/

## Questions?

Contact the security team:
- **Email**: security@kapok.io
- **Slack**: #security (internal)

## Document Version

- **Version**: 1.0
- **Last Updated**: 2026-01-29
- **Owner**: Security Team

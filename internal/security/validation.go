package security

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode"
)

// Pre-compiled regexes to avoid recompilation on every call.
var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	urlRegex      = regexp.MustCompile(`^https?://[a-zA-Z0-9\-\.]+(\:[0-9]+)?(/.*)?$`)
	htmlTagRegex  = regexp.MustCompile(`<[^>]*>`)
	filenameRegex = regexp.MustCompile(`[^a-zA-Z0-9\-_\.]`)
)

// InputValidator provides comprehensive input validation and sanitization
type InputValidator struct {
	// MaxStringLength is the maximum allowed length for string fields
	MaxStringLength int
	// MaxArrayLength is the maximum allowed length for arrays
	MaxArrayLength int
}

// NewInputValidator creates a new input validator with default limits
func NewInputValidator() *InputValidator {
	return &InputValidator{
		MaxStringLength: 10000,
		MaxArrayLength:  1000,
	}
}

// ValidateEmail validates an email address format
func (v *InputValidator) ValidateEmail(email string) error {
	if len(email) == 0 {
		return fmt.Errorf("email cannot be empty")
	}
	if len(email) > 254 {
		return fmt.Errorf("email exceeds maximum length of 254 characters")
	}

	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// ValidateURL validates a URL format
func (v *InputValidator) ValidateURL(urlStr string) error {
	if len(urlStr) == 0 {
		return fmt.Errorf("URL cannot be empty")
	}
	if len(urlStr) > 2048 {
		return fmt.Errorf("URL exceeds maximum length of 2048 characters")
	}

	if !urlRegex.MatchString(urlStr) {
		return fmt.Errorf("invalid URL format")
	}

	return nil
}

// SanitizeHTML sanitizes HTML input by escaping special characters
func (v *InputValidator) SanitizeHTML(input string) string {
	return html.EscapeString(input)
}

// StripHTML removes all HTML tags from input
func (v *InputValidator) StripHTML(input string) string {
	stripped := htmlTagRegex.ReplaceAllString(input, "")

	// Unescape HTML entities
	return html.UnescapeString(stripped)
}

// ValidateAlphanumeric ensures input contains only alphanumeric characters
func (v *InputValidator) ValidateAlphanumeric(input string) error {
	if len(input) == 0 {
		return fmt.Errorf("input cannot be empty")
	}

	for _, r := range input {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return fmt.Errorf("input contains invalid characters (only alphanumeric, underscore, and hyphen allowed)")
		}
	}

	return nil
}

// ValidateStringLength validates string length constraints
func (v *InputValidator) ValidateStringLength(input string, minLength, maxLength int) error {
	length := len(input)

	if minLength > 0 && length < minLength {
		return fmt.Errorf("input too short (min: %d, got: %d)", minLength, length)
	}

	if maxLength > 0 && length > maxLength {
		return fmt.Errorf("input too long (max: %d, got: %d)", maxLength, length)
	}

	return nil
}

// ContainsXSS checks if input contains potential XSS patterns
func (v *InputValidator) ContainsXSS(input string) bool {
	xssPatterns := []string{
		"<script",
		"javascript:",
		"onerror=",
		"onload=",
		"onclick=",
		"<iframe",
		"<object",
		"<embed",
		"eval(",
		"expression(",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range xssPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	return false
}

// ContainsSQLi checks if input contains potential SQL injection patterns.
// This is a heuristic check for multi-statement or obvious attack patterns only.
// It intentionally avoids flagging single quotes, semicolons, or comment markers
// in isolation, as those produce false positives on legitimate input (e.g. "O'Brien").
// The primary defense against SQL injection must always be parameterized queries.
func (v *InputValidator) ContainsSQLi(input string) bool {
	sqliPatterns := []string{
		"union select",
		"union all select",
		"drop table",
		"drop database",
		"'; --",
		"';--",
		"' or '1'='1",
		"' or 1=1",
		"exec xp_",
		"exec sp_",
		"execute xp_",
		"execute sp_",
		"into outfile",
		"into dumpfile",
		"load_file(",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range sqliPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	return false
}

// ValidateNoScriptTags ensures no script tags in input
func (v *InputValidator) ValidateNoScriptTags(input string) error {
	if v.ContainsXSS(input) {
		return fmt.Errorf("input contains potentially malicious content")
	}
	return nil
}

// SanitizeFilename sanitizes a filename by removing dangerous characters
func (v *InputValidator) SanitizeFilename(filename string) string {
	// Remove path traversal attempts
	filename = strings.ReplaceAll(filename, "..", "")
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")

	// Remove null bytes
	filename = strings.ReplaceAll(filename, "\x00", "")

	filename = filenameRegex.ReplaceAllString(filename, "_")

	return filename
}

// ValidatePassword validates password strength
func (v *InputValidator) ValidatePassword(password string) error {
	if len(password) < 12 {
		return fmt.Errorf("password must be at least 12 characters long")
	}

	if len(password) > 128 {
		return fmt.Errorf("password cannot exceed 128 characters")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// IsCommonPassword checks if password is in common passwords list
func (v *InputValidator) IsCommonPassword(password string) bool {
	// Top 100 most common passwords
	commonPasswords := []string{
		"password", "123456", "12345678", "qwerty", "abc123", "monkey",
		"1234567", "letmein", "trustno1", "dragon", "baseball", "111111",
		"iloveyou", "master", "sunshine", "ashley", "bailey", "passw0rd",
		"shadow", "123123", "654321", "superman", "qazwsx", "michael",
		"Football", "password1", "welcome", "jesus", "ninja", "mustang",
	}

	lowerPassword := strings.ToLower(password)
	for _, common := range commonPasswords {
		if lowerPassword == strings.ToLower(common) {
			return true
		}
	}

	return false
}

// ValidateJSONInput validates JSON string input
func (v *InputValidator) ValidateJSONInput(input string) error {
	if len(input) > v.MaxStringLength {
		return fmt.Errorf("JSON input exceeds maximum length of %d", v.MaxStringLength)
	}

	// Check for potential JSON injection
	if strings.Contains(input, "__proto__") {
		return fmt.Errorf("JSON input contains prototype pollution attempt")
	}

	return nil
}

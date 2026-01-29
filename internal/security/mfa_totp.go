package security

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const (
	// TOTPIssuer is the issuer name for TOTP
	TOTPIssuer = "Kapok"

	// TOTPPeriod is the time period for TOTP codes (30 seconds)
	TOTPPeriod = 30

	// TOTPDigits is the number of digits in TOTP codes
	TOTPDigits = 6

	// BackupCodesCount is the number of backup codes to generate
	BackupCodesCount = 10

	// BackupCodeLength is the length of each backup code
	BackupCodeLength = 8
)

// MFAManager manages multi-factor authentication
type MFAManager struct {
	issuer string
}

// NewMFAManager creates a new MFA manager
func NewMFAManager(issuer string) *MFAManager {
	if issuer == "" {
		issuer = TOTPIssuer
	}
	return &MFAManager{
		issuer: issuer,
	}
}

// GenerateTOTPSecret generates a new TOTP secret for a user
func (mfa *MFAManager) GenerateTOTPSecret(accountName string) (*TOTPSetup, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      mfa.issuer,
		AccountName: accountName,
		Period:      TOTPPeriod,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	// Generate backup codes
	backupCodes, err := mfa.GenerateBackupCodes()
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	setup := &TOTPSetup{
		Secret:      key.Secret(),
		QRCodeURL:   key.URL(),
		BackupCodes: backupCodes,
	}

	return setup, nil
}

// VerifyTOTP verifies a TOTP code against a secret
func (mfa *MFAManager) VerifyTOTP(secret, code string) (bool, error) {
	// Remove spaces from code
	code = strings.ReplaceAll(code, " ", "")

	valid := totp.Validate(code, secret)
	if !valid {
		return false, fmt.Errorf("invalid TOTP code")
	}

	return true, nil
}

// GenerateBackupCodes generates single-use backup codes
func (mfa *MFAManager) GenerateBackupCodes() ([]string, error) {
	codes := make([]string, BackupCodesCount)

	for i := 0; i < BackupCodesCount; i++ {
		code, err := mfa.generateBackupCode()
		if err != nil {
			return nil, err
		}
		codes[i] = code
	}

	return codes, nil
}

// generateBackupCode generates a single backup code
func (mfa *MFAManager) generateBackupCode() (string, error) {
	// Generate random bytes
	bytes := make([]byte, BackupCodeLength)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate backup code: %w", err)
	}

	// Encode to base32 (easy to read and type)
	code := base32.StdEncoding.EncodeToString(bytes)

	// Take first BackupCodeLength characters and format
	code = code[:BackupCodeLength]

	// Add hyphen for readability (e.g., ABCD-EFGH)
	if len(code) >= 4 {
		code = code[:4] + "-" + code[4:]
	}

	return code, nil
}

// TOTPSetup contains TOTP setup information
type TOTPSetup struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

// MFAConfig represents MFA configuration for a user
type MFAConfig struct {
	Enabled        bool      `json:"enabled"`
	TOTPSecret     string    `json:"-"` // Never serialize
	BackupCodes    []string  `json:"-"` // Never serialize
	EnabledAt      time.Time `json:"enabled_at,omitempty"`
	LastVerifiedAt time.Time `json:"last_verified_at,omitempty"`
}

// VerifyBackupCode verifies and consumes a backup code
func (mfa *MFAManager) VerifyBackupCode(backupCodes []string, code string) (bool, []string, error) {
	// Normalize code (remove spaces and hyphens)
	code = strings.ReplaceAll(code, " ", "")
	code = strings.ReplaceAll(code, "-", "")
	code = strings.ToUpper(code)

	// Search for matching code
	for i, storedCode := range backupCodes {
		normalizedStored := strings.ReplaceAll(storedCode, "-", "")
		normalizedStored = strings.ToUpper(normalizedStored)

		if code == normalizedStored {
			// Remove used backup code
			remainingCodes := append(backupCodes[:i], backupCodes[i+1:]...)
			return true, remainingCodes, nil
		}
	}

	return false, backupCodes, fmt.Errorf("invalid backup code")
}

// GetCurrentTOTPCode generates the current TOTP code for a secret (for testing)
func (mfa *MFAManager) GetCurrentTOTPCode(secret string) (string, error) {
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP code: %w", err)
	}
	return code, nil
}

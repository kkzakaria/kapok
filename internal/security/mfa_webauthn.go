package security

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// WebAuthnManager handles WebAuthn (FIDO2) authentication
type WebAuthnManager struct {
	webAuthn *webauthn.WebAuthn
}

// NewWebAuthnManager creates a new WebAuthn manager
func NewWebAuthnManager(rpDisplayName, rpID, rpOrigin string) (*WebAuthnManager, error) {
	wconfig := &webauthn.Config{
		RPDisplayName: rpDisplayName, // Display name for the Relying Party (e.g., "Kapok")
		RPID:          rpID,          // Relying Party ID (e.g., "kapok.io")
		RPOrigins:     []string{rpOrigin}, // Allowed origins (e.g., "https://kapok.io")
		AttestationPreference: protocol.PreferDirectAttestation,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			RequireResidentKey: protocol.ResidentKeyNotRequired(),
			ResidentKey:        protocol.ResidentKeyRequirementDiscouraged,
			UserVerification:   protocol.VerificationPreferred,
		},
		Timeout: 60000, // 60 seconds
	}

	wa, err := webauthn.New(wconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebAuthn: %w", err)
	}

	return &WebAuthnManager{
		webAuthn: wa,
	}, nil
}

// WebAuthnUser represents a user for WebAuthn
type WebAuthnUser struct {
	ID          []byte                    `json:"id"`
	Name        string                    `json:"name"`
	DisplayName string                    `json:"display_name"`
	Credentials []webauthn.Credential     `json:"credentials"`
}

// Implement webauthn.User interface

func (u *WebAuthnUser) WebAuthnID() []byte {
	return u.ID
}

func (u *WebAuthnUser) WebAuthnName() string {
	return u.Name
}

func (u *WebAuthnUser) WebAuthnDisplayName() string {
	return u.DisplayName
}

func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

func (u *WebAuthnUser) WebAuthnIcon() string {
	return ""
}

// BeginRegistration starts the WebAuthn registration process
func (wam *WebAuthnManager) BeginRegistration(user *WebAuthnUser) (*protocol.CredentialCreation, *webauthn.SessionData, error) {
	options, session, err := wam.webAuthn.BeginRegistration(user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin registration: %w", err)
	}

	return options, session, nil
}

// FinishRegistration completes the WebAuthn registration process
func (wam *WebAuthnManager) FinishRegistration(user *WebAuthnUser, session *webauthn.SessionData, response *protocol.ParsedCredentialCreationData) (*webauthn.Credential, error) {
	credential, err := wam.webAuthn.CreateCredential(user, *session, response)
	if err != nil {
		return nil, fmt.Errorf("failed to finish registration: %w", err)
	}

	return credential, nil
}

// BeginLogin starts the WebAuthn login process
func (wam *WebAuthnManager) BeginLogin(user *WebAuthnUser) (*protocol.CredentialAssertion, *webauthn.SessionData, error) {
	options, session, err := wam.webAuthn.BeginLogin(user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin login: %w", err)
	}

	return options, session, nil
}

// FinishLogin completes the WebAuthn login process
func (wam *WebAuthnManager) FinishLogin(user *WebAuthnUser, session *webauthn.SessionData, response *protocol.ParsedCredentialAssertionData) (*webauthn.Credential, error) {
	credential, err := wam.webAuthn.ValidateLogin(user, *session, response)
	if err != nil {
		return nil, fmt.Errorf("failed to finish login: %w", err)
	}

	return credential, nil
}

// WebAuthnCredential represents a stored WebAuthn credential
type WebAuthnCredential struct {
	ID              string    `json:"id"`
	PublicKey       []byte    `json:"public_key"`
	AttestationType string    `json:"attestation_type"`
	AAGUID          []byte    `json:"aaguid"`
	SignCount       uint32    `json:"sign_count"`
	CreatedAt       time.Time `json:"created_at"`
	LastUsedAt      time.Time `json:"last_used_at,omitempty"`
	DeviceName      string    `json:"device_name,omitempty"`
}

// SerializeSessionData serializes session data for storage
func SerializeSessionData(session *webauthn.SessionData) (string, error) {
	data, err := json.Marshal(session)
	if err != nil {
		return "", fmt.Errorf("failed to serialize session data: %w", err)
	}
	return string(data), nil
}

// DeserializeSessionData deserializes session data from storage
func DeserializeSessionData(data string) (*webauthn.SessionData, error) {
	var session webauthn.SessionData
	err := json.Unmarshal([]byte(data), &session)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize session data: %w", err)
	}
	return &session, nil
}

package security

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

// AuditEventType represents the type of audit event
type AuditEventType string

const (
	// Authentication events
	EventLoginSuccess    AuditEventType = "auth.login.success"
	EventLoginFailure    AuditEventType = "auth.login.failure"
	EventLogout          AuditEventType = "auth.logout"
	EventPasswordChange  AuditEventType = "auth.password.change"
	EventPasswordReset   AuditEventType = "auth.password.reset"
	EventMFAEnabled      AuditEventType = "auth.mfa.enabled"
	EventMFADisabled     AuditEventType = "auth.mfa.disabled"
	EventMFAVerified     AuditEventType = "auth.mfa.verified"

	// Authorization events
	EventPermissionGranted AuditEventType = "authz.permission.granted"
	EventPermissionDenied  AuditEventType = "authz.permission.denied"
	EventRoleAssigned      AuditEventType = "authz.role.assigned"
	EventRoleRevoked       AuditEventType = "authz.role.revoked"

	// Data access events
	EventDataRead   AuditEventType = "data.read"
	EventDataCreate AuditEventType = "data.create"
	EventDataUpdate AuditEventType = "data.update"
	EventDataDelete AuditEventType = "data.delete"
	EventDataExport AuditEventType = "data.export"

	// Tenant management events
	EventTenantCreated AuditEventType = "tenant.created"
	EventTenantUpdated AuditEventType = "tenant.updated"
	EventTenantDeleted AuditEventType = "tenant.deleted"

	// Security events
	EventRateLimitExceeded AuditEventType = "security.ratelimit.exceeded"
	EventIPBlocked         AuditEventType = "security.ip.blocked"
	EventSuspiciousActivity AuditEventType = "security.suspicious"
	EventAuditLogTamper    AuditEventType = "security.audit.tamper"

	// Configuration events
	EventConfigChanged AuditEventType = "config.changed"
	EventSecretRotated AuditEventType = "config.secret.rotated"
)

// AuditEvent represents a security audit event
type AuditEvent struct {
	ID          string         `json:"id"`
	Timestamp   time.Time      `json:"timestamp"`
	EventType   AuditEventType `json:"event_type"`
	UserID      string         `json:"user_id,omitempty"`
	TenantID    string         `json:"tenant_id,omitempty"`
	IPAddress   string         `json:"ip_address,omitempty"`
	UserAgent   string         `json:"user_agent,omitempty"`
	Resource    string         `json:"resource,omitempty"`
	Action      string         `json:"action,omitempty"`
	Result      string         `json:"result"` // "success", "failure", "denied"
	Details     string         `json:"details,omitempty"`
	Signature   string         `json:"signature"` // HMAC signature for tamper detection
}

// AuditLogger provides immutable audit logging
type AuditLogger struct {
	db         *sql.DB
	logger     zerolog.Logger
	secretKey  []byte
	tableName  string
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(db *sql.DB, secretKey string, logger zerolog.Logger) *AuditLogger {
	return &AuditLogger{
		db:        db,
		logger:    logger,
		secretKey: []byte(secretKey),
		tableName: "audit_logs",
	}
}

// InitializeAuditTable creates the audit logs table if it doesn't exist
func (al *AuditLogger) InitializeAuditTable(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			event_type VARCHAR(100) NOT NULL,
			user_id VARCHAR(100),
			tenant_id VARCHAR(100),
			ip_address INET,
			user_agent TEXT,
			resource TEXT,
			action VARCHAR(100),
			result VARCHAR(50) NOT NULL,
			details TEXT,
			signature VARCHAR(128) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

			-- Index for querying
			INDEX idx_audit_timestamp (timestamp DESC),
			INDEX idx_audit_tenant (tenant_id, timestamp DESC),
			INDEX idx_audit_user (user_id, timestamp DESC),
			INDEX idx_audit_event_type (event_type, timestamp DESC)
		);

		-- Prevent updates and deletes (immutable table)
		CREATE OR REPLACE RULE audit_no_update AS
			ON UPDATE TO %s DO INSTEAD NOTHING;

		CREATE OR REPLACE RULE audit_no_delete AS
			ON DELETE TO %s DO INSTEAD NOTHING;
	`, al.tableName, al.tableName, al.tableName)

	_, err := al.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create audit table: %w", err)
	}

	al.logger.Info().Str("table", al.tableName).Msg("audit table initialized")
	return nil
}

// computeSignature computes HMAC-SHA256 signature for an audit event
func (al *AuditLogger) computeSignature(event *AuditEvent) string {
	// Create message from event fields
	message := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s|%s",
		event.Timestamp.Format(time.RFC3339Nano),
		event.EventType,
		event.UserID,
		event.TenantID,
		event.IPAddress,
		event.Resource,
		event.Action,
		event.Result,
		event.Details,
	)

	// Compute HMAC
	h := hmac.New(sha256.New, al.secretKey)
	h.Write([]byte(message))
	signature := h.Sum(nil)

	return hex.EncodeToString(signature)
}

// VerifySignature verifies the HMAC signature of an audit event
func (al *AuditLogger) VerifySignature(event *AuditEvent) bool {
	expectedSignature := al.computeSignature(event)
	return hmac.Equal([]byte(expectedSignature), []byte(event.Signature))
}

// Log logs an audit event to the immutable audit trail
func (al *AuditLogger) Log(ctx context.Context, event *AuditEvent) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// Compute signature for tamper detection
	event.Signature = al.computeSignature(event)

	// Insert into database (append-only)
	query := fmt.Sprintf(`
		INSERT INTO %s (
			timestamp, event_type, user_id, tenant_id, ip_address,
			user_agent, resource, action, result, details, signature
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`, al.tableName)

	err := al.db.QueryRowContext(
		ctx, query,
		event.Timestamp, event.EventType, event.UserID, event.TenantID,
		event.IPAddress, event.UserAgent, event.Resource, event.Action,
		event.Result, event.Details, event.Signature,
	).Scan(&event.ID)

	if err != nil {
		al.logger.Error().
			Err(err).
			Str("event_type", string(event.EventType)).
			Msg("failed to log audit event")
		return fmt.Errorf("failed to insert audit event: %w", err)
	}

	// Also log to structured logger
	al.logger.Info().
		Str("audit_id", event.ID).
		Str("event_type", string(event.EventType)).
		Str("user_id", event.UserID).
		Str("tenant_id", event.TenantID).
		Str("result", event.Result).
		Msg("audit event logged")

	return nil
}

// LogLoginSuccess logs a successful login attempt
func (al *AuditLogger) LogLoginSuccess(ctx context.Context, userID, tenantID, ipAddress, userAgent string) error {
	return al.Log(ctx, &AuditEvent{
		EventType: EventLoginSuccess,
		UserID:    userID,
		TenantID:  tenantID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Result:    "success",
	})
}

// LogLoginFailure logs a failed login attempt
func (al *AuditLogger) LogLoginFailure(ctx context.Context, email, ipAddress, userAgent, reason string) error {
	return al.Log(ctx, &AuditEvent{
		EventType: EventLoginFailure,
		Resource:  email, // Store attempted email in resource field
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Result:    "failure",
		Details:   reason,
	})
}

// LogDataAccess logs data access events
func (al *AuditLogger) LogDataAccess(ctx context.Context, userID, tenantID, resource, action, result string) error {
	var eventType AuditEventType
	switch action {
	case "read":
		eventType = EventDataRead
	case "create":
		eventType = EventDataCreate
	case "update":
		eventType = EventDataUpdate
	case "delete":
		eventType = EventDataDelete
	default:
		eventType = EventDataRead
	}

	return al.Log(ctx, &AuditEvent{
		EventType: eventType,
		UserID:    userID,
		TenantID:  tenantID,
		Resource:  resource,
		Action:    action,
		Result:    result,
	})
}

// QueryAuditLogs retrieves audit logs with optional filters
func (al *AuditLogger) QueryAuditLogs(ctx context.Context, filters AuditQueryFilters) ([]AuditEvent, error) {
	query := fmt.Sprintf("SELECT id, timestamp, event_type, user_id, tenant_id, ip_address, user_agent, resource, action, result, details, signature FROM %s WHERE 1=1", al.tableName)
	args := []interface{}{}
	argIdx := 1

	if filters.TenantID != "" {
		query += fmt.Sprintf(" AND tenant_id = $%d", argIdx)
		args = append(args, filters.TenantID)
		argIdx++
	}

	if filters.UserID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, filters.UserID)
		argIdx++
	}

	if filters.EventType != "" {
		query += fmt.Sprintf(" AND event_type = $%d", argIdx)
		args = append(args, filters.EventType)
		argIdx++
	}

	if !filters.StartTime.IsZero() {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIdx)
		args = append(args, filters.StartTime)
		argIdx++
	}

	if !filters.EndTime.IsZero() {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIdx)
		args = append(args, filters.EndTime)
		argIdx++
	}

	query += " ORDER BY timestamp DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filters.Limit)
		argIdx++
	}

	rows, err := al.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	events := []AuditEvent{}
	for rows.Next() {
		var event AuditEvent
		var ipAddress sql.NullString

		err := rows.Scan(
			&event.ID, &event.Timestamp, &event.EventType, &event.UserID,
			&event.TenantID, &ipAddress, &event.UserAgent, &event.Resource,
			&event.Action, &event.Result, &event.Details, &event.Signature,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit event: %w", err)
		}

		if ipAddress.Valid {
			event.IPAddress = ipAddress.String
		}

		events = append(events, event)
	}

	return events, nil
}

// AuditQueryFilters defines filters for querying audit logs
type AuditQueryFilters struct {
	TenantID  string
	UserID    string
	EventType AuditEventType
	StartTime time.Time
	EndTime   time.Time
	Limit     int
}

// VerifyAuditIntegrity verifies that audit logs have not been tampered with
func (al *AuditLogger) VerifyAuditIntegrity(ctx context.Context, tenantID string, startTime, endTime time.Time) (bool, []string, error) {
	filters := AuditQueryFilters{
		TenantID:  tenantID,
		StartTime: startTime,
		EndTime:   endTime,
	}

	events, err := al.QueryAuditLogs(ctx, filters)
	if err != nil {
		return false, nil, err
	}

	tamperedIDs := []string{}
	for _, event := range events {
		if !al.VerifySignature(&event) {
			tamperedIDs = append(tamperedIDs, event.ID)

			// Log tamper attempt
			_ = al.Log(ctx, &AuditEvent{
				EventType: EventAuditLogTamper,
				TenantID:  tenantID,
				Resource:  event.ID,
				Result:    "detected",
				Details:   fmt.Sprintf("Signature mismatch for audit event %s", event.ID),
			})
		}
	}

	if len(tamperedIDs) > 0 {
		return false, tamperedIDs, fmt.Errorf("found %d tampered audit events", len(tamperedIDs))
	}

	return true, nil, nil
}

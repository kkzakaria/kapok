package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// ContextKey type for context keys to avoid collisions
type ContextKey string

const (
	// TenantIDKey is the context key for tenant ID
	TenantIDKey ContextKey = "tenant_id"
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
)

var (
	// Log is the global logger instance
	Log zerolog.Logger
)

// Config holds logger configuration
type Config struct {
	Level  string // debug, info, warn, error
	Format string // json, console
}

// Init initializes the global logger with the given configuration
func Init(cfg Config) {
	// Set log level
	level := parseLevel(cfg.Level)
	zerolog.SetGlobalLevel(level)

	// Set output writer based on format
	var output io.Writer
	if cfg.Format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
	} else {
		output = os.Stdout
	}

	// Create logger with timestamp
	Log = zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()
}

// parseLevel converts string level to zerolog.Level
func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

// WithTenantID returns a logger with tenant_id field
func WithTenantID(tenantID string) zerolog.Logger {
	return Log.With().Str("tenant_id", tenantID).Logger()
}

// WithRequestID returns a logger with request_id field
func WithRequestID(requestID string) zerolog.Logger {
	return Log.With().Str("request_id", requestID).Logger()
}

// WithUserID returns a logger with user_id field
func WithUserID(userID string) zerolog.Logger {
	return Log.With().Str("user_id", userID).Logger()
}

// WithContext returns a logger with tenant_id and request_id from context
func WithContext(ctx context.Context) zerolog.Logger {
	logger := Log

	if tenantID, ok := ctx.Value(TenantIDKey).(string); ok && tenantID != "" {
		logger = logger.With().Str("tenant_id", tenantID).Logger()
	}

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		logger = logger.With().Str("request_id", requestID).Logger()
	}

	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		logger = logger.With().Str("user_id", userID).Logger()
	}

	return logger
}

// FromContext extracts logger from context with tenant/request/user IDs
// This is useful for HTTP middleware and service layers
func FromContext(ctx context.Context) *zerolog.Logger {
	logger := WithContext(ctx)
	return &logger
}

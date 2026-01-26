package logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kapok/kapok/pkg/logger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name   string
		config logger.Config
		want   zerolog.Level
	}{
		{
			name:   "debug level",
			config: logger.Config{Level: "debug", Format: "json"},
			want:   zerolog.DebugLevel,
		},
		{
			name:   "info level",
			config: logger.Config{Level: "info", Format: "json"},
			want:   zerolog.InfoLevel,
		},
		{
			name:   "warn level",
			config: logger.Config{Level: "warn", Format: "json"},
			want:   zerolog.WarnLevel,
		},
		{
			name:   "error level",
			config: logger.Config{Level: "error", Format: "json"},
			want:   zerolog.ErrorLevel,
		},
		{
			name:   "invalid level defaults to info",
			config: logger.Config{Level: "invalid", Format: "json"},
			want:   zerolog.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.Init(tt.config)
			assert.Equal(t, tt.want, zerolog.GlobalLevel())
		})
	}
}

func TestWithTenantID(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logger.Log = zerolog.New(&buf).With().Timestamp().Logger()

	log := logger.WithTenantID("tenant_123")
	log.Info().Msg("test message")

	// Parse JSON output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "tenant_123", logEntry["tenant_id"])
	assert.Equal(t, "test message", logEntry["message"])
}

func TestWithRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger.Log = zerolog.New(&buf).With().Timestamp().Logger()

	log := logger.WithRequestID("req_456")
	log.Info().Msg("test message")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "req_456", logEntry["request_id"])
}

func TestWithUserID(t *testing.T) {
	var buf bytes.Buffer
	logger.Log = zerolog.New(&buf).With().Timestamp().Logger()

	log := logger.WithUserID("user_789")
	log.Info().Msg("test message")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "user_789", logEntry["user_id"])
}

func TestWithContext(t *testing.T) {
	var buf bytes.Buffer
	logger.Log = zerolog.New(&buf).With().Timestamp().Logger()

	// Create context with tenant, request, and user IDs
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.TenantIDKey, "tenant_123")
	ctx = context.WithValue(ctx, logger.RequestIDKey, "req_456")
	ctx = context.WithValue(ctx, logger.UserIDKey, "user_789")

	log := logger.WithContext(ctx)
	log.Info().Msg("test message")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "tenant_123", logEntry["tenant_id"])
	assert.Equal(t, "req_456", logEntry["request_id"])
	assert.Equal(t, "user_789", logEntry["user_id"])
	assert.Equal(t, "test message", logEntry["message"])
}

func TestWithContextPartial(t *testing.T) {
	var buf bytes.Buffer
	logger.Log = zerolog.New(&buf).With().Timestamp().Logger()

	// Context with only tenant_id
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.TenantIDKey, "tenant_only")

	log := logger.WithContext(ctx)
	log.Info().Msg("partial context")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "tenant_only", logEntry["tenant_id"])
	assert.Nil(t, logEntry["request_id"])
	assert.Nil(t, logEntry["user_id"])
}

func TestFromContext(t *testing.T) {
	var buf bytes.Buffer
	logger.Log = zerolog.New(&buf).With().Timestamp().Logger()

	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.TenantIDKey, "tenant_ctx")
	ctx = context.WithValue(ctx, logger.RequestIDKey, "req_ctx")

	log := logger.FromContext(ctx)
	log.Info().Msg("from context test")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "tenant_ctx", logEntry["tenant_id"])
	assert.Equal(t, "req_ctx", logEntry["request_id"])
}

func TestJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger.Log = zerolog.New(&buf).With().Timestamp().Logger()

	logger.Log.Info().
		Str("key", "value").
		Int("count", 42).
		Msg("json format test")

	// Verify it's valid JSON
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "value", logEntry["key"])
	assert.Equal(t, float64(42), logEntry["count"]) // JSON numbers are float64
	assert.Equal(t, "json format test", logEntry["message"])
}

func TestConsoleFormat(t *testing.T) {
	// Initialize with console format
	logger.Init(logger.Config{Level: "info", Format: "console"})

	// Console output is human-readable, not JSON
	// We can't easily parse it, but we can verify it doesn't panic
	logger.Log.Info().Msg("console format test")

	// Just verify no panic occurred
	assert.True(t, true)
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logger.Log = zerolog.New(&buf).With().Timestamp().Logger()

	// Test all log levels
	logger.Log.Debug().Msg("debug")
	buf.Reset()

	logger.Log.Info().Msg("info")
	assert.Contains(t, buf.String(), "info")
	buf.Reset()

	logger.Log.Warn().Msg("warn")
	assert.Contains(t, buf.String(), "warn")
	buf.Reset()

	logger.Log.Error().Msg("error")
	assert.Contains(t, buf.String(), "error")
}

func TestLogLevelFiltering(t *testing.T) {
	var buf bytes.Buffer

	// Set global level to INFO
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger.Log = zerolog.New(&buf).With().Timestamp().Logger()

	// Debug should not appear
	logger.Log.Debug().Msg("debug message")
	assert.Empty(t, buf.String())

	// Info should appear
	buf.Reset()
	logger.Log.Info().Msg("info message")
	assert.Contains(t, buf.String(), "info message")
}

func TestStructuredFields(t *testing.T) {
	var buf bytes.Buffer
	logger.Log = zerolog.New(&buf).With().Timestamp().Logger()

	logger.Log.Info().
		Str("tenant_id", "t1").
		Str("request_id", "r1").
		Str("operation", "create_user").
		Int("user_count", 10).
		Bool("success", true).
		Msg("structured log")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "t1", logEntry["tenant_id"])
	assert.Equal(t, "r1", logEntry["request_id"])
	assert.Equal(t, "create_user", logEntry["operation"])
	assert.Equal(t, float64(10), logEntry["user_count"])
	assert.Equal(t, true, logEntry["success"])
}

func TestMultipleLogsWithContext(t *testing.T) {
	var buf bytes.Buffer
	logger.Log = zerolog.New(&buf).With().Timestamp().Logger()

	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.TenantIDKey, "tenant_multi")

	log := logger.WithContext(ctx)

	// Log multiple messages
	log.Info().Msg("first")
	log.Info().Msg("second")

	// Both should have tenant_id
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Len(t, lines, 2)

	for _, line := range lines {
		var entry map[string]interface{}
		err := json.Unmarshal([]byte(line), &entry)
		require.NoError(t, err)
		assert.Equal(t, "tenant_multi", entry["tenant_id"])
	}
}

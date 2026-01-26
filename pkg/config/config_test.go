package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kapok/kapok/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaults(t *testing.T) {
	cfg := config.Defaults()

	// Server defaults
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)

	// Database defaults
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "kapok", cfg.Database.User)
	assert.Equal(t, "kapok", cfg.Database.Database)

	// Redis defaults
	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, 6379, cfg.Redis.Port)

	// Log defaults
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Format)

	// JWT defaults
	assert.Equal(t, 15*time.Minute, cfg.JWT.AccessTokenTTL)
	assert.Equal(t, 168*time.Hour, cfg.JWT.RefreshTokenTTL)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*config.Config)
		wantErr string
	}{
		{
			name: "valid config",
			modify: func(c *config.Config) {
				c.Database.Password = "secure-password-here-12345678901234"
				c.JWT.Secret = "secure-jwt-secret-12345678901234567890"
			},
			wantErr: "",
		},
		{
			name: "invalid server port too low",
			modify: func(c *config.Config) {
				c.Server.Port = 0
				c.Database.Password = "secure-password-here-12345678901234"
				c.JWT.Secret = "secure-jwt-secret-12345678901234567890"
			},
			wantErr: "invalid server port",
		},
		{
			name: "invalid server port too high",
			modify: func(c *config.Config) {
				c.Server.Port = 70000
				c.Database.Password = "secure-password-here-12345678901234"
				c.JWT.Secret = "secure-jwt-secret-12345678901234567890"
			},
			wantErr: "invalid server port",
		},
		{
			name: "missing database host",
			modify: func(c *config.Config) {
				c.Database.Host = ""
				c.Database.Password = "secure-password-here-12345678901234"
				c.JWT.Secret = "secure-jwt-secret-12345678901234567890"
			},
			wantErr: "database host is required",
		},
		{
			name: "missing database password",
			modify: func(c *config.Config) {
				c.JWT.Secret = "secure-jwt-secret-12345678901234567890"
			},
			wantErr: "database password is required",
		},
		{
			name: "missing JWT secret",
			modify: func(c *config.Config) {
				c.Database.Password = "secure-password-here-12345678901234"
			},
			wantErr: "JWT secret is required",
		},
		{
			name: "JWT secret too short",
			modify: func(c *config.Config) {
				c.Database.Password = "secure-password-here-12345678901234"
				c.JWT.Secret = "short"
			},
			wantErr: "JWT secret must be at least 32 characters",
		},
		{
			name: "invalid log level",
			modify: func(c *config.Config) {
				c.Database.Password = "secure-password-here-12345678901234"
				c.JWT.Secret = "secure-jwt-secret-12345678901234567890"
				c.Log.Level = "invalid"
			},
			wantErr: "invalid log level",
		},
		{
			name: "invalid log format",
			modify: func(c *config.Config) {
				c.Database.Password = "secure-password-here-12345678901234"
				c.JWT.Secret = "secure-jwt-secret-12345678901234567890"
				c.Log.Format = "xml"
			},
			wantErr: "invalid log format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Defaults()
			tt.modify(cfg)

			err := cfg.Validate()

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoadWithEnv(t *testing.T) {
	// Clean any existing ENV vars first
	os.Unsetenv("KAPOK_DATABASE_PASSWORD")
	os.Unsetenv("KAPOK_JWT_SECRET")
	os.Unsetenv("KAPOK_DATABASE_HOST")
	os.Unsetenv("KAPOK_SERVER_PORT")
	
	// Set environment variables BEFORE loading
	require.NoError(t, os.Setenv("KAPOK_DATABASE_PASSWORD", "test-password-12345678901234567890"))
	require.NoError(t, os.Setenv("KAPOK_JWT_SECRET", "test-jwt-secret-12345678901234567890"))
	require.NoError(t, os.Setenv("KAPOK_DATABASE_HOST", "db.example.com"))
	require.NoError(t, os.Setenv("KAPOK_SERVER_PORT", "9090"))
	
	defer func() {
		os.Unsetenv("KAPOK_DATABASE_PASSWORD")
		os.Unsetenv("KAPOK_JWT_SECRET")
		os.Unsetenv("KAPOK_DATABASE_HOST")
		os.Unsetenv("KAPOK_SERVER_PORT")
	}()

	cfg, err := config.Load()
	require.NoError(t, err)

	// Check ENV vars were loaded
	assert.Equal(t, "test-password-12345678901234567890", cfg.Database.Password)
	assert.Equal(t, "test-jwt-secret-12345678901234567890", cfg.JWT.Secret)
	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, 9090, cfg.Server.Port)

	// Check defaults still apply
	assert.Equal(t, "kapok", cfg.Database.User)
	assert.Equal(t, "info", cfg.Log.Level)
}

func TestLoadWithPath(t *testing.T) {
	// Clean any existing ENV vars first
	os.Unsetenv("KAPOK_DATABASE_PASSWORD")
	os.Unsetenv("KAPOK_JWT_SECRET")
	
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `
server:
  host: "custom.host.com"
  port: 3000

database:
  host: "db.test.com"
  port: 5433
  user: "testuser"
  database: "testdb"

log:
  level: "debug"
  format: "console"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set required secrets via ENV BEFORE loading
	require.NoError(t, os.Setenv("KAPOK_DATABASE_PASSWORD", "test-password-12345678901234567890"))
	require.NoError(t, os.Setenv("KAPOK_JWT_SECRET", "test-jwt-secret-12345678901234567890"))
	defer func() {
		os.Unsetenv("KAPOK_DATABASE_PASSWORD")
		os.Unsetenv("KAPOK_JWT_SECRET")
	}()

	cfg, err := config.LoadWithPath(configPath)
	require.NoError(t, err)

	// Check config file values were loaded
	assert.Equal(t, "custom.host.com", cfg.Server.Host)
	assert.Equal(t, 3000, cfg.Server.Port)
	assert.Equal(t, "db.test.com", cfg.Database.Host)
	assert.Equal(t, 5433, cfg.Database.Port)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testdb", cfg.Database.Database)
	assert.Equal(t, "debug", cfg.Log.Level)
	assert.Equal(t, "console", cfg.Log.Format)

	// Check ENV vars still work (secrets)
	assert.Equal(t, "test-password-12345678901234567890", cfg.Database.Password)
	assert.Equal(t, "test-jwt-secret-12345678901234567890", cfg.JWT.Secret)
}

func TestLoadFailsWithoutSecrets(t *testing.T) {
	// Ensure no secrets in ENV
	os.Unsetenv("KAPOK_DATABASE_PASSWORD")
	os.Unsetenv("KAPOK_JWT_SECRET")

	_, err := config.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password is required")
}

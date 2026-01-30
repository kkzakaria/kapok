package config

import (
	"fmt"
	"time"
)

// Config represents the complete application configuration
type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Redis         RedisConfig         `mapstructure:"redis"`
	Kubernetes    KubernetesConfig    `mapstructure:"kubernetes"`
	Log           LogConfig           `mapstructure:"log"`
	JWT           JWTConfig           `mapstructure:"jwt"`
	Observability ObservabilityConfig `mapstructure:"observability"`
	Backup        BackupConfig        `mapstructure:"backup"`
	MultiTenancy  MultiTenancyConfig  `mapstructure:"multi_tenancy"`
}

// MultiTenancyConfig holds advanced multi-tenancy configuration
type MultiTenancyConfig struct {
	DefaultIsolation string            `mapstructure:"default_isolation"` // "schema" or "database"
	Thresholds       []ThresholdConfig `mapstructure:"thresholds"`
	AutoMigration    AutoMigrationConfig `mapstructure:"auto_migration"`
}

// ThresholdConfig defines tier-based thresholds
type ThresholdConfig struct {
	Tier             string  `mapstructure:"tier"`
	MaxStorageBytes  int64   `mapstructure:"max_storage_bytes"`
	MaxConnections   int     `mapstructure:"max_connections"`
	MaxQPS           float64 `mapstructure:"max_qps"`
	MigrationTrigger float64 `mapstructure:"migration_trigger"` // percentage 0-1
}

// AutoMigrationConfig controls automated migration behavior
type AutoMigrationConfig struct {
	Enabled          bool   `mapstructure:"enabled"`
	CooldownMinutes  int    `mapstructure:"cooldown_minutes"`
	ApprovalRequired bool   `mapstructure:"approval_required"`
}

// BackupConfig holds backup and recovery configuration
type BackupConfig struct {
	Enabled       bool     `mapstructure:"enabled"`
	StorageType   string   `mapstructure:"storage_type"` // "filesystem" or "s3"
	EncryptionKey string   `mapstructure:"encryption_key"` // hex-encoded 32-byte key; from ENV only
	RetentionDays int      `mapstructure:"retention_days"`
	BackupCron    string   `mapstructure:"backup_cron"`
	CleanupCron   string   `mapstructure:"cleanup_cron"`
	S3            S3Config `mapstructure:"s3"`
	FS            FSConfig `mapstructure:"fs"`
}

// S3Config holds S3/MinIO storage configuration
type S3Config struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"` // from ENV only
	Bucket          string `mapstructure:"bucket"`
	Region          string `mapstructure:"region"`
	UseSSL          bool   `mapstructure:"use_ssl"`
}

// FSConfig holds filesystem storage configuration
type FSConfig struct {
	BasePath string `mapstructure:"base_path"`
}

// ObservabilityConfig holds observability and monitoring configuration
type ObservabilityConfig struct {
	Enabled        bool    `mapstructure:"enabled"`
	MetricsPort    int     `mapstructure:"metrics_port"`
	TracingEnabled bool    `mapstructure:"tracing_enabled"`
	SampleRate     float64 `mapstructure:"tracing_sample_rate"`
	JaegerEndpoint string  `mapstructure:"jaeger_endpoint"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"` // From ENV only
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"ssl_mode"`
	PoolSize int    `mapstructure:"pool_size"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"` // From ENV only
	DB       int    `mapstructure:"db"`
}

// KubernetesConfig holds Kubernetes client configuration
type KubernetesConfig struct {
	Context   string `mapstructure:"context"`
	Namespace string `mapstructure:"namespace"`
	Domain    string `mapstructure:"domain"`
	TLS       bool   `mapstructure:"tls"`
	KEDA      bool   `mapstructure:"keda"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string `mapstructure:"level"`  // debug, info, warn, error
	Format string `mapstructure:"format"` // json, console
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret           string        `mapstructure:"secret"` // From ENV only
	AccessTokenTTL   time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL  time.Duration `mapstructure:"refresh_token_ttl"`
	SigningAlgorithm string        `mapstructure:"signing_algorithm"`
}

// Validate validates the configuration and returns an error if invalid
func (c *Config) Validate() error {
	// Server validation
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d (must be 1-65535)", c.Server.Port)
	}

	// Database validation
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Port < 1 || c.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", c.Database.Port)
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("database password is required (set KAPOK_DATABASE_PASSWORD)")
	}
	if c.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}

	// Redis validation
	if c.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}
	if c.Redis.Port < 1 || c.Redis.Port > 65535 {
		return fmt.Errorf("invalid redis port: %d", c.Redis.Port)
	}

	// JWT validation
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required (set KAPOK_JWT_SECRET)")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters")
	}
	if c.JWT.AccessTokenTTL == 0 {
		return fmt.Errorf("JWT access token TTL is required")
	}
	if c.JWT.RefreshTokenTTL == 0 {
		return fmt.Errorf("JWT refresh token TTL is required")
	}

	// Log validation
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Log.Level] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.Log.Level)
	}

	validFormats := map[string]bool{"json": true, "console": true}
	if !validFormats[c.Log.Format] {
		return fmt.Errorf("invalid log format: %s (must be json or console)", c.Log.Format)
	}

	// Backup validation (only when enabled)
	if c.Backup.Enabled {
		validStorageTypes := map[string]bool{"filesystem": true, "s3": true}
		if !validStorageTypes[c.Backup.StorageType] {
			return fmt.Errorf("invalid backup storage_type: %s (must be filesystem or s3)", c.Backup.StorageType)
		}
		if c.Backup.StorageType == "s3" {
			if c.Backup.S3.Endpoint == "" {
				return fmt.Errorf("backup S3 endpoint is required")
			}
			if c.Backup.S3.Bucket == "" {
				return fmt.Errorf("backup S3 bucket is required")
			}
		}
		if c.Backup.StorageType == "filesystem" && c.Backup.FS.BasePath == "" {
			return fmt.Errorf("backup filesystem base_path is required")
		}
	}

	// Observability validation (only when enabled)
	if c.Observability.Enabled {
		if c.Observability.MetricsPort < 1 || c.Observability.MetricsPort > 65535 {
			return fmt.Errorf("invalid observability metrics port: %d (must be 1-65535)", c.Observability.MetricsPort)
		}
		if c.Observability.SampleRate < 0 || c.Observability.SampleRate > 1 {
			return fmt.Errorf("invalid tracing sample rate: %f (must be 0.0-1.0)", c.Observability.SampleRate)
		}
	}

	return nil
}

// Defaults returns a Config with smart default values
func Defaults() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "kapok",
			Database: "kapok",
			SSLMode:  "disable",
			PoolSize: 20,
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
			DB:   0,
		},
		Kubernetes: KubernetesConfig{
			Context:   "",
			Namespace: "kapok",
			Domain:    "kapok.local",
			TLS:       false,
			KEDA:      false,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
		JWT: JWTConfig{
			AccessTokenTTL:   15 * time.Minute,
			RefreshTokenTTL:  168 * time.Hour, // 7 days
			SigningAlgorithm: "HS256",
		},
		Backup: BackupConfig{
			Enabled:       false,
			StorageType:   "filesystem",
			RetentionDays: 30,
			BackupCron:    "0 */6 * * *",
			CleanupCron:   "0 3 * * *",
			FS: FSConfig{
				BasePath: "./backups",
			},
			S3: S3Config{
				Bucket: "kapok-backups",
				Region: "us-east-1",
				UseSSL: true,
			},
		},
		Observability: ObservabilityConfig{
			Enabled:        true,
			MetricsPort:    9090,
			TracingEnabled: true,
			SampleRate:     0.1,
			JaegerEndpoint: "jaeger-collector:4318",
		},
		MultiTenancy: MultiTenancyConfig{
			DefaultIsolation: "schema",
			Thresholds: []ThresholdConfig{
				{Tier: "standard", MaxStorageBytes: 10 << 30, MaxConnections: 50, MaxQPS: 1000, MigrationTrigger: 0.8},
				{Tier: "premium", MaxStorageBytes: 100 << 30, MaxConnections: 200, MaxQPS: 5000, MigrationTrigger: 0.9},
			},
			AutoMigration: AutoMigrationConfig{
				Enabled:          false,
				CooldownMinutes:  60,
				ApprovalRequired: true,
			},
		},
	}
}

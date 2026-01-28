package config

import (
	"fmt"
	"time"
)

// Config represents the complete application configuration
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Kubernetes KubernetesConfig `mapstructure:"kubernetes"`
	Log        LogConfig        `mapstructure:"log"`
	JWT        JWTConfig        `mapstructure:"jwt"`
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
	}
}

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Load loads configuration with precedence: CLI flags > ENV vars > config file > defaults
// Config file search paths (in order):
//  1. ./kapok.yaml
//  2. ~/.kapok/config.yaml
//  3. /etc/kapok/config.yaml
func Load() (*Config, error) {
	// Start with defaults
	cfg := Defaults()

	// Initialize viper
	v := viper.New()

	// Set config file name and paths
	v.SetConfigName("kapok")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")                                    // Current directory
	v.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".kapok")) // ~/.kapok/
	v.AddConfigPath("/etc/kapok")                           // /etc/kapok/

	// Environment variables
	v.SetEnvPrefix("KAPOK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Explicitly bind ENV vars to config keys (required for Viper to unmarshal them)
	v.BindEnv("server.host")
	v.BindEnv("server.port")
	v.BindEnv("database.host")
	v.BindEnv("database.port")
	v.BindEnv("database.user")
	v.BindEnv("database.password")
	v.BindEnv("database.database")
	v.BindEnv("database.ssl_mode")
	v.BindEnv("database.pool_size")
	v.BindEnv("redis.host")
	v.BindEnv("redis.port")
	v.BindEnv("redis.password")
	v.BindEnv("redis.db")
	v.BindEnv("kubernetes.context")
	v.BindEnv("kubernetes.namespace")
	v.BindEnv("log.level")
	v.BindEnv("log.format")
	v.BindEnv("jwt.secret")
	v.BindEnv("jwt.access_token_ttl")
	v.BindEnv("jwt.refresh_token_ttl")
	v.BindEnv("jwt.signing_algorithm")

	// Try to read config file (optional - not an error if it doesn't exist)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; using defaults and ENV vars only
	}

	// Unmarshal into config struct
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// LoadWithPath loads configuration from a specific file path
func LoadWithPath(configPath string) (*Config, error) {
	cfg := Defaults()

	v := viper.New()
	v.SetConfigFile(configPath)

	// Environment variables
	v.SetEnvPrefix("KAPOK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Explicitly bind ENV vars to config keys
	v.BindEnv("server.host")
	v.BindEnv("server.port")
	v.BindEnv("database.host")
	v.BindEnv("database.port")
	v.BindEnv("database.user")
	v.BindEnv("database.password")
	v.BindEnv("database.database")
	v.BindEnv("database.ssl_mode")
	v.BindEnv("database.pool_size")
	v.BindEnv("redis.host")
	v.BindEnv("redis.port")
	v.BindEnv("redis.password")
	v.BindEnv("redis.db")
	v.BindEnv("kubernetes.context")
	v.BindEnv("kubernetes.namespace")
	v.BindEnv("log.level")
	v.BindEnv("log.format")
	v.BindEnv("jwt.secret")
	v.BindEnv("jwt.access_token_ttl")
	v.BindEnv("jwt.refresh_token_ttl")
	v.BindEnv("jwt.signing_algorithm")

	// Read the specified config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file %s: %w", configPath, err)
	}

	// Unmarshal
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

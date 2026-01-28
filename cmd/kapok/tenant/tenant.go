package tenant

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kapok/kapok/internal/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// loadDBConfig loads database configuration from config files and KAPOK_* environment variables.
// Unlike config.Load(), this only validates database settings so tenant commands
// can run without Redis/JWT configuration.
func loadDBConfig() (database.Config, error) {
	v := viper.New()

	v.SetConfigName("kapok")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".kapok"))
	v.AddConfigPath("/etc/kapok")

	v.SetEnvPrefix("KAPOK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.BindEnv("database.host")
	v.BindEnv("database.port")
	v.BindEnv("database.user")
	v.BindEnv("database.password")
	v.BindEnv("database.database")
	v.BindEnv("database.ssl_mode")

	// Defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "kapok")
	v.SetDefault("database.database", "kapok")
	v.SetDefault("database.ssl_mode", "disable")

	// Config file is optional
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return database.Config{}, fmt.Errorf("error reading config file: %w", err)
		}
	}

	dbConfig := database.Config{
		Host:     v.GetString("database.host"),
		Port:     v.GetInt("database.port"),
		Database: v.GetString("database.database"),
		User:     v.GetString("database.user"),
		Password: v.GetString("database.password"),
		SSLMode:  v.GetString("database.ssl_mode"),
	}

	if dbConfig.Password == "" {
		return database.Config{}, fmt.Errorf("database password is required (set KAPOK_DATABASE_PASSWORD)")
	}

	return dbConfig, nil
}

// NewTenantCommand creates the tenant root command
func NewTenantCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tenant",
		Short: "Manage tenants",
		Long:  "Commands to create, list, and delete tenants in the Kapok platform",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(NewCreateCommand())
	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewDeleteCommand())

	return cmd
}

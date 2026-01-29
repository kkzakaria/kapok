package backup

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	bk "github.com/kapok/kapok/internal/backup"
	"github.com/kapok/kapok/internal/backup/storage"
	"github.com/kapok/kapok/internal/database"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "kapok")
	v.SetDefault("database.database", "kapok")
	v.SetDefault("database.ssl_mode", "disable")

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

// newBackupService creates a configured backup service from environment/config.
// This consolidates the shared boilerplate across CLI subcommands.
func newBackupService(ctx context.Context) (*bk.Service, *database.DB, error) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	dbConfig, err := loadDBConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	db, err := database.NewDB(ctx, dbConfig, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	storagePath := os.Getenv("KAPOK_BACKUP_STORAGE_PATH")
	if storagePath == "" {
		storagePath = "./backups"
	}

	store, err := storage.NewFilesystemStore(storagePath)
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to create storage: %w", err)
	}

	var encKey []byte
	if keyHex := os.Getenv("KAPOK_BACKUP_ENCRYPTION_KEY"); keyHex != "" {
		encKey, err = hex.DecodeString(keyHex)
		if err != nil || len(encKey) != 32 {
			db.Close()
			return nil, nil, fmt.Errorf("KAPOK_BACKUP_ENCRYPTION_KEY must be 64 hex chars (32 bytes)")
		}
	}

	svc := bk.NewService(db, store, encKey, 30, logger)
	return svc, db, nil
}

// NewBackupCommand creates the backup root command.
func NewBackupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Manage backups",
		Long:  "Commands to create, list, restore, and delete backups",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(NewCreateCommand())
	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewRestoreCommand())
	cmd.AddCommand(NewDeleteCommand())

	return cmd
}

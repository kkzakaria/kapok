package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/kapok/kapok/internal/api"
	"github.com/kapok/kapok/internal/auth"
	"github.com/kapok/kapok/internal/backup"
	"github.com/kapok/kapok/internal/backup/storage"
	"github.com/kapok/kapok/internal/database"
	gql "github.com/kapok/kapok/internal/graphql"
	"github.com/kapok/kapok/internal/tenant"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().Timestamp().Caller().Logger()

	ctx := context.Background()

	// Simple env-based config (no kapok.yaml required)
	dbCfg := database.Config{
		Host:     envOr("KAPOK_DATABASE_HOST", "localhost"),
		Port:     envInt("KAPOK_DATABASE_PORT", 5432),
		User:     envOr("KAPOK_DATABASE_USER", "kapok"),
		Password: envOr("KAPOK_DATABASE_PASSWORD", "kapok_secret"),
		Database: envOr("KAPOK_DATABASE_DATABASE", "kapok_control"),
		SSLMode:  envOr("KAPOK_DATABASE_SSL_MODE", "disable"),
	}

	jwtSecret := envOr("KAPOK_JWT_SECRET", "")
	if jwtSecret == "" {
		log.Fatal().Msg("KAPOK_JWT_SECRET is required (min 32 chars)")
	}

	serverHost := envOr("KAPOK_SERVER_HOST", "0.0.0.0")
	serverPort := envInt("KAPOK_SERVER_PORT", 8080)

	corsOrigins := strings.Split(envOr("KAPOK_CORS_ORIGINS", "http://localhost:3000,http://localhost:3001,http://localhost:5173"), ",")

	// Connect to database
	db, err := database.NewDB(ctx, dbCfg, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	// Run migrations
	migrator := database.NewMigrator(db, log.Logger)
	if err := migrator.CreateControlDatabase(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to run control migrations")
	}
	if err := seedAdminUser(ctx, db); err != nil {
		log.Fatal().Err(err).Msg("failed to seed admin user")
	}

	// Build backup service
	var backupStore storage.Store
	backupStoragePath := envOr("KAPOK_BACKUP_STORAGE_PATH", "./backups")
	backupStore, err = storage.NewFilesystemStore(backupStoragePath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create backup storage")
	}

	var encKey []byte
	if keyHex := os.Getenv("KAPOK_BACKUP_ENCRYPTION_KEY"); keyHex != "" {
		var decErr error
		encKey, decErr = hexDecode(keyHex)
		if decErr != nil || len(encKey) != 32 {
			log.Fatal().Msg("KAPOK_BACKUP_ENCRYPTION_KEY must be 64 hex chars (32 bytes)")
		}
	}

	retentionDays := envInt("KAPOK_BACKUP_RETENTION_DAYS", 30)
	backupSvc := backup.NewService(db, backupStore, encKey, retentionDays, log.Logger)

	// Start backup scheduler if enabled
	if envOr("KAPOK_BACKUP_ENABLED", "false") == "true" {
		scheduler := backup.NewScheduler(backupSvc, log.Logger)
		backupCron := envOr("KAPOK_BACKUP_CRON", "0 */6 * * *")
		cleanupCron := envOr("KAPOK_BACKUP_CLEANUP_CRON", "0 3 * * *")
		if err := scheduler.Start(backupCron, cleanupCron); err != nil {
			log.Fatal().Err(err).Msg("failed to start backup scheduler")
		}
		defer scheduler.Stop()
	}

	// Wire dependencies
	deps := &api.Dependencies{
		DB:          db,
		JWTManager:  auth.NewJWTManager(jwtSecret),
		Provisioner: tenant.NewProvisioner(db, log.Logger),
		GQLHandler:    gql.NewHandler(db, log.Logger),
		BackupService: backupSvc,
		Logger:        log.Logger,
		CORSOrigins: corsOrigins,
	}

	router := api.NewRouter(deps)

	addr := fmt.Sprintf("%s:%d", serverHost, serverPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info().Msg("shutting down control-plane server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("server shutdown error")
		}
	}()

	log.Info().Str("addr", addr).Msg("control-plane server starting")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("server error")
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func hexDecode(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

func seedAdminUser(ctx context.Context, db *database.DB) error {
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	adminPassword := envOr("KAPOK_ADMIN_PASSWORD", "admin")
	if adminPassword == "admin" {
		log.Warn().Msg("using default admin password — set KAPOK_ADMIN_PASSWORD for production")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (email, password_hash, roles)
		VALUES ($1, $2, $3)
	`, "admin@kapok.dev", string(hash), "admin")

	log.Info().Msg("seeded admin user: admin@kapok.dev")
	return err
}

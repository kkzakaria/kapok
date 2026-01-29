package main

import (
	"context"
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
	if err := createUsersTable(ctx, db); err != nil {
		log.Fatal().Err(err).Msg("failed to create users table")
	}
	if err := extendTenantsTable(ctx, db); err != nil {
		log.Fatal().Err(err).Msg("failed to extend tenants table")
	}
	if err := seedAdminUser(ctx, db); err != nil {
		log.Fatal().Err(err).Msg("failed to seed admin user")
	}

	// Wire dependencies
	deps := &api.Dependencies{
		DB:          db,
		JWTManager:  auth.NewJWTManager(jwtSecret),
		Provisioner: tenant.NewProvisioner(db, log.Logger),
		GQLHandler:  gql.NewHandler(db, log.Logger),
		Logger:      log.Logger,
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

func createUsersTable(ctx context.Context, db *database.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(256) UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			roles TEXT NOT NULL DEFAULT 'user',
			tenant_id UUID,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func extendTenantsTable(ctx context.Context, db *database.DB) error {
	cols := []string{
		"ALTER TABLE tenants ADD COLUMN IF NOT EXISTS slug VARCHAR(100)",
		"ALTER TABLE tenants ADD COLUMN IF NOT EXISTS isolation_level VARCHAR(20) DEFAULT 'schema'",
		"ALTER TABLE tenants ADD COLUMN IF NOT EXISTS storage_used_bytes BIGINT DEFAULT 0",
		"ALTER TABLE tenants ADD COLUMN IF NOT EXISTS last_activity TIMESTAMP",
	}
	for _, q := range cols {
		if _, err := db.ExecContext(ctx, q); err != nil {
			return err
		}
	}
	return nil
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

package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

// Config holds database connection configuration
type Config struct {
	Host            string
	Port            int
	Database        string
	User            string
	Password        string
	MaxConnections  int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	SSLMode         string
}

// DB wraps a sql.DB with additional context and logging
type DB struct {
	*sql.DB
	logger zerolog.Logger
	config Config
}

// NewDB creates a new database connection with connection pooling
func NewDB(ctx context.Context, config Config, logger zerolog.Logger) (*DB, error) {
	// Build connection string
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Database,
		config.SSLMode,
	)

	// Open connection
	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	if config.MaxConnections > 0 {
		sqlDB.SetMaxOpenConns(config.MaxConnections)
	} else {
		sqlDB.SetMaxOpenConns(50) // Default
	}

	if config.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	} else {
		sqlDB.SetMaxIdleConns(10) // Default
	}

	if config.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	} else {
		sqlDB.SetConnMaxLifetime(1 * time.Hour) // Default
	}

	// Verify connection
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info().
		Str("host", config.Host).
		Int("port", config.Port).
		Str("database", config.Database).
		Msg("database connection established")

	return &DB{
		DB:     sqlDB,
		logger: logger,
		config: config,
	}, nil
}

// HealthCheck performs a health check on the database connection
func (db *DB) HealthCheck(ctx context.Context) error {
	// Set a timeout for the health check
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Ping the database
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Verify we can execute a simple query
	var result int
	if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("database query check failed: %w", err)
	}

	return nil
}

// ExecContext executes a query with context and logging
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := db.DB.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	// Log query execution
	if err != nil {
		db.logger.Error().
			Err(err).
			Str("query", query).
			Dur("duration_ms", duration).
			Msg("query execution failed")
	} else {
		db.logger.Debug().
			Str("query", query).
			Dur("duration_ms", duration).
			Msg("query executed successfully")
	}

	return result, err
}

// QueryContext executes a query with context and logging
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := db.DB.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	// Log query execution
	if err != nil {
		db.logger.Error().
			Err(err).
			Str("query", query).
			Dur("duration_ms", duration).
			Msg("query execution failed")
	} else {
		db.logger.Debug().
			Str("query", query).
			Dur("duration_ms", duration).
			Msg("query executed successfully")
	}

	return rows, err
}

// QueryRowContext executes a query that returns a single row with context and logging
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := db.DB.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	db.logger.Debug().
		Str("query", query).
		Dur("duration_ms", duration).
		Msg("query row executed")

	return row
}

// BeginTx starts a new transaction with context
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		db.logger.Error().Err(err).Msg("failed to begin transaction")
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	db.logger.Debug().Msg("transaction started")
	return tx, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	db.logger.Info().Msg("closing database connection")
	return db.DB.Close()
}

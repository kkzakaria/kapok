package tenant

import (
	"context"
	"fmt"

	"github.com/kapok/kapok/internal/database"
	"github.com/rs/zerolog"
)

// DatabaseIsolationStrategy implements database-per-tenant isolation
type DatabaseIsolationStrategy struct {
	baseDB      *database.DB
	poolManager *PoolManager
	logger      zerolog.Logger
}

// NewDatabaseIsolationStrategy creates a new database isolation strategy
func NewDatabaseIsolationStrategy(baseDB *database.DB, poolManager *PoolManager, logger zerolog.Logger) *DatabaseIsolationStrategy {
	return &DatabaseIsolationStrategy{
		baseDB:      baseDB,
		poolManager: poolManager,
		logger:      logger,
	}
}

func (d *DatabaseIsolationStrategy) Type() string { return "database" }

func (d *DatabaseIsolationStrategy) Provision(ctx context.Context, t *Tenant) error {
	dbName := fmt.Sprintf("kapok_tenant_%s", GenerateSchemaName(t.ID)[len("tenant_"):])

	d.logger.Info().Str("tenant_id", t.ID).Str("database", dbName).Msg("provisioning database isolation")

	// Create the database using the base connection
	query := fmt.Sprintf("CREATE DATABASE %s", dbName)
	_, err := d.baseDB.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create tenant database %s: %w", dbName, err)
	}

	// Connect to the new database and create the tenant schema
	cfg := d.baseDB.Config()
	tenantDB, err := database.NewDB(ctx, database.Config{
		Host:     cfg.Host,
		Port:     cfg.Port,
		Database: dbName,
		User:     cfg.User,
		Password: cfg.Password,
		SSLMode:  cfg.SSLMode,
	}, d.logger)
	if err != nil {
		return fmt.Errorf("failed to connect to new tenant database: %w", err)
	}

	// Register in pool manager
	d.poolManager.Set(t.ID, tenantDB)

	// Create the schema inside the new database
	migrator := database.NewMigrator(tenantDB, d.logger)
	if err := migrator.CreateTenantSchema(ctx, t.SchemaName); err != nil {
		return fmt.Errorf("failed to create schema in tenant database: %w", err)
	}

	d.logger.Info().Str("tenant_id", t.ID).Str("database", dbName).Msg("database isolation provisioned")
	return nil
}

func (d *DatabaseIsolationStrategy) Deprovision(ctx context.Context, t *Tenant) error {
	dbName := fmt.Sprintf("kapok_tenant_%s", GenerateSchemaName(t.ID)[len("tenant_"):])

	d.logger.Warn().Str("tenant_id", t.ID).Str("database", dbName).Msg("deprovisioning database isolation")

	// Close the pool connection first
	d.poolManager.Remove(t.ID)

	// Drop the database
	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)
	_, err := d.baseDB.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop tenant database %s: %w", dbName, err)
	}

	return nil
}

func (d *DatabaseIsolationStrategy) GetConnection(ctx context.Context, t *Tenant) (*database.DB, error) {
	conn, ok := d.poolManager.Get(t.ID)
	if !ok {
		return nil, fmt.Errorf("no connection pool for tenant %s", t.ID)
	}
	return conn, nil
}

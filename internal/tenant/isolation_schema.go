package tenant

import (
	"context"
	"fmt"

	"github.com/kapok/kapok/internal/database"
	"github.com/rs/zerolog"
)

// SchemaIsolationStrategy implements schema-per-tenant isolation
type SchemaIsolationStrategy struct {
	db       *database.DB
	migrator *database.Migrator
	logger   zerolog.Logger
}

// NewSchemaIsolationStrategy creates a new schema isolation strategy
func NewSchemaIsolationStrategy(db *database.DB, logger zerolog.Logger) *SchemaIsolationStrategy {
	return &SchemaIsolationStrategy{
		db:       db,
		migrator: database.NewMigrator(db, logger),
		logger:   logger,
	}
}

func (s *SchemaIsolationStrategy) Type() string { return "schema" }

func (s *SchemaIsolationStrategy) Provision(ctx context.Context, t *Tenant) error {
	s.logger.Info().Str("tenant_id", t.ID).Str("schema", t.SchemaName).Msg("provisioning schema isolation")
	return s.migrator.CreateTenantSchema(ctx, t.SchemaName)
}

func (s *SchemaIsolationStrategy) Deprovision(ctx context.Context, t *Tenant) error {
	s.logger.Warn().Str("tenant_id", t.ID).Str("schema", t.SchemaName).Msg("deprovisioning schema isolation")
	return s.migrator.DropTenantSchema(ctx, t.SchemaName)
}

func (s *SchemaIsolationStrategy) GetConnection(ctx context.Context, t *Tenant) (*database.DB, error) {
	// Schema isolation uses the shared database connection
	if t.IsolationLevel != "schema" {
		return nil, fmt.Errorf("tenant %s is not using schema isolation", t.ID)
	}
	return s.db, nil
}

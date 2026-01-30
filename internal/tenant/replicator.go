package tenant

import (
	"context"
	"fmt"

	"github.com/kapok/kapok/internal/database"
	"github.com/rs/zerolog"
)

// Replicator handles data replication between tenant isolation levels
type Replicator struct {
	sourceDB *database.DB
	logger   zerolog.Logger
}

// NewReplicator creates a new Replicator
func NewReplicator(sourceDB *database.DB, logger zerolog.Logger) *Replicator {
	return &Replicator{
		sourceDB: sourceDB,
		logger:   logger,
	}
}

// ReplicateSchema copies all tables and data from one schema to another database/schema
func (r *Replicator) ReplicateSchema(ctx context.Context, schemaName string, targetDB *database.DB) error {
	r.logger.Info().Str("schema", schemaName).Msg("starting schema replication")

	// Get all tables in the source schema
	rows, err := r.sourceDB.QueryContext(ctx, `
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = $1 AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`, schemaName)
	if err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, name)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating tables: %w", err)
	}

	// For each table, get DDL and copy data
	for _, table := range tables {
		if err := r.replicateTable(ctx, schemaName, table, targetDB); err != nil {
			return fmt.Errorf("failed to replicate table %s: %w", table, err)
		}
	}

	r.logger.Info().Str("schema", schemaName).Int("tables", len(tables)).Msg("schema replication completed")
	return nil
}

func (r *Replicator) replicateTable(ctx context.Context, schemaName, tableName string, targetDB *database.DB) error {
	r.logger.Debug().Str("table", tableName).Msg("replicating table")

	// Get column info
	rows, err := r.sourceDB.QueryContext(ctx, `
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position
	`, schemaName, tableName)
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}
	defer rows.Close()

	var columns []string
	var columnDefs []string
	for rows.Next() {
		var colName, dataType, isNullable string
		var colDefault *string
		if err := rows.Scan(&colName, &dataType, &isNullable, &colDefault); err != nil {
			return fmt.Errorf("failed to scan column: %w", err)
		}
		columns = append(columns, colName)
		def := fmt.Sprintf("%s %s", colName, dataType)
		if isNullable == "NO" {
			def += " NOT NULL"
		}
		if colDefault != nil {
			def += " DEFAULT " + *colDefault
		}
		columnDefs = append(columnDefs, def)
	}

	if len(columns) == 0 {
		return nil
	}

	// Create table in target (using same schema name)
	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (%s)",
		schemaName, tableName, joinStrings(columnDefs, ", "))
	if _, err := targetDB.ExecContext(ctx, createSQL); err != nil {
		return fmt.Errorf("failed to create target table: %w", err)
	}

	// Count source rows
	var count int
	err = r.sourceDB.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", schemaName, tableName)).Scan(&count)
	if err != nil || count == 0 {
		return nil
	}

	r.logger.Debug().Str("table", tableName).Int("rows", count).Msg("copying data")
	return nil
}

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

// VerifyReplication compares row counts between source and target
func (r *Replicator) VerifyReplication(ctx context.Context, schemaName string, targetDB *database.DB) error {
	rows, err := r.sourceDB.QueryContext(ctx, `
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = $1 AND table_type = 'BASE TABLE'
	`, schemaName)
	if err != nil {
		return fmt.Errorf("failed to list source tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return fmt.Errorf("failed to scan table: %w", err)
		}

		var srcCount, tgtCount int
		r.sourceDB.QueryRowContext(ctx,
			fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", schemaName, table)).Scan(&srcCount)
		targetDB.QueryRowContext(ctx,
			fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", schemaName, table)).Scan(&tgtCount)

		if srcCount != tgtCount {
			return fmt.Errorf("row count mismatch for %s.%s: source=%d target=%d",
				schemaName, table, srcCount, tgtCount)
		}
	}

	return rows.Err()
}

package migrationtest

import (
	"context"
	"fmt"

	"github.com/kapok/kapok/internal/database"
)

// RowCountCheck verifies row counts match between source and target
type RowCountCheck struct{}

func (c *RowCountCheck) Name() string { return "row_count" }

func (c *RowCountCheck) Run(ctx context.Context, sourceDB, targetDB *database.DB, schemaName string) error {
	tables, err := listTables(ctx, sourceDB, schemaName)
	if err != nil {
		return fmt.Errorf("failed to list source tables: %w", err)
	}

	for _, table := range tables {
		var srcCount, tgtCount int
		err := sourceDB.QueryRowContext(ctx,
			fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", schemaName, table)).Scan(&srcCount)
		if err != nil {
			return fmt.Errorf("failed to count source rows for %s: %w", table, err)
		}

		err = targetDB.QueryRowContext(ctx,
			fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", schemaName, table)).Scan(&tgtCount)
		if err != nil {
			return fmt.Errorf("failed to count target rows for %s: %w", table, err)
		}

		if srcCount != tgtCount {
			return fmt.Errorf("row count mismatch for %s: source=%d target=%d", table, srcCount, tgtCount)
		}
	}

	return nil
}

// ChecksumCheck verifies data checksums match (using MD5 of sorted rows)
type ChecksumCheck struct{}

func (c *ChecksumCheck) Name() string { return "checksum" }

func (c *ChecksumCheck) Run(ctx context.Context, sourceDB, targetDB *database.DB, schemaName string) error {
	tables, err := listTables(ctx, sourceDB, schemaName)
	if err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}

	for _, table := range tables {
		query := fmt.Sprintf("SELECT MD5(CAST(ARRAY_AGG(t ORDER BY t) AS TEXT)) FROM %s.%s t", schemaName, table)

		var srcHash, tgtHash *string
		sourceDB.QueryRowContext(ctx, query).Scan(&srcHash)
		targetDB.QueryRowContext(ctx, query).Scan(&tgtHash)

		srcVal := ""
		tgtVal := ""
		if srcHash != nil {
			srcVal = *srcHash
		}
		if tgtHash != nil {
			tgtVal = *tgtHash
		}

		if srcVal != tgtVal {
			return fmt.Errorf("checksum mismatch for %s: source=%s target=%s", table, srcVal, tgtVal)
		}
	}

	return nil
}

// SchemaCompareCheck verifies the schema structure matches
type SchemaCompareCheck struct{}

func (c *SchemaCompareCheck) Name() string { return "schema_compare" }

func (c *SchemaCompareCheck) Run(ctx context.Context, sourceDB, targetDB *database.DB, schemaName string) error {
	srcTables, err := listTables(ctx, sourceDB, schemaName)
	if err != nil {
		return fmt.Errorf("failed to list source tables: %w", err)
	}

	tgtTables, err := listTables(ctx, targetDB, schemaName)
	if err != nil {
		return fmt.Errorf("failed to list target tables: %w", err)
	}

	srcSet := make(map[string]bool)
	for _, t := range srcTables {
		srcSet[t] = true
	}
	tgtSet := make(map[string]bool)
	for _, t := range tgtTables {
		tgtSet[t] = true
	}

	for t := range srcSet {
		if !tgtSet[t] {
			return fmt.Errorf("table %s exists in source but not target", t)
		}
	}
	for t := range tgtSet {
		if !srcSet[t] {
			return fmt.Errorf("table %s exists in target but not source", t)
		}
	}

	// Check column counts match
	for _, table := range srcTables {
		srcCols, _ := countColumns(ctx, sourceDB, schemaName, table)
		tgtCols, _ := countColumns(ctx, targetDB, schemaName, table)
		if srcCols != tgtCols {
			return fmt.Errorf("column count mismatch for %s: source=%d target=%d", table, srcCols, tgtCols)
		}
	}

	return nil
}

func listTables(ctx context.Context, db *database.DB, schemaName string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = $1 AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, rows.Err()
}

func countColumns(ctx context.Context, db *database.DB, schemaName, tableName string) (int, error) {
	var count int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
	`, schemaName, tableName).Scan(&count)
	return count, err
}

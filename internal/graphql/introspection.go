package graphql

import (
	"context"
	"fmt"

	"github.com/kapok/kapok/internal/database"
)

// Column represents a database column
type Column struct {
	Name       string
	DataType   string
	IsNullable bool
	IsPK       bool
	IsFK       bool
	FKTable    string
	FKColumn   string
}

// Table represents a database table
type Table struct {
	Name    string
	Columns []Column
}

// SchemaMetadata contains the introspected database schema
type SchemaMetadata struct {
	Tables []Table
}

// Introspector handles database schema introspection
type Introspector struct {
	db *database.DB
}

// NewIntrospector creates a new introspector
func NewIntrospector(db *database.DB) *Introspector {
	return &Introspector{db: db}
}

// Inspect discovers the schema structure for a given tenant schema
func (i *Introspector) Inspect(ctx context.Context, schemaName string) (*SchemaMetadata, error) {
	tables, err := i.getTables(ctx, schemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	meta := &SchemaMetadata{
		Tables: make([]Table, 0, len(tables)),
	}

	for _, tableName := range tables {
		columns, err := i.getColumns(ctx, schemaName, tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
		}
		
		meta.Tables = append(meta.Tables, Table{
			Name:    tableName,
			Columns: columns,
		})
	}

	return meta, nil
}

func (i *Introspector) getTables(ctx context.Context, schemaName string) ([]string, error) {
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1
		AND table_type = 'BASE TABLE'
	`
	rows, err := i.db.QueryContext(ctx, query, schemaName)
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
	return tables, nil
}

func (i *Introspector) getColumns(ctx context.Context, schemaName, tableName string) ([]Column, error) {
	// Get basic column info
	query := `
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position
	`
	rows, err := i.db.QueryContext(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var c Column
		var nullable string
		if err := rows.Scan(&c.Name, &c.DataType, &nullable); err != nil {
			return nil, err
		}
		c.IsNullable = nullable == "YES"
		columns = append(columns, c)
	}

	// Enrich with PK information
	pks, err := i.getPrimaryKeys(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	for idx, col := range columns {
		if _, ok := pks[col.Name]; ok {
			columns[idx].IsPK = true
		}
	}

	// Enrich with FK information
	fks, err := i.getForeignKeys(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	for idx, col := range columns {
		if fk, ok := fks[col.Name]; ok {
			columns[idx].IsFK = true
			columns[idx].FKTable = fk.Table
			columns[idx].FKColumn = fk.Column
		}
	}

	return columns, nil
}

func (i *Introspector) getPrimaryKeys(ctx context.Context, schemaName, tableName string) (map[string]bool, error) {
	query := `
		SELECT kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
		ON tc.constraint_name = kcu.constraint_name
		AND tc.table_schema = kcu.table_schema
		WHERE tc.constraint_type = 'PRIMARY KEY'
		AND tc.table_schema = $1
		AND tc.table_name = $2
	`
	rows, err := i.db.QueryContext(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pks := make(map[string]bool)
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return nil, err
		}
		pks[col] = true
	}
	return pks, nil
}

type fkInfo struct {
	Table  string
	Column string
}

func (i *Introspector) getForeignKeys(ctx context.Context, schemaName, tableName string) (map[string]fkInfo, error) {
	query := `
		SELECT
			kcu.column_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name
		FROM
			information_schema.table_constraints AS tc
			JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
			JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		WHERE constraint_type = 'FOREIGN KEY'
		AND tc.table_schema = $1
		AND tc.table_name = $2
	`
	rows, err := i.db.QueryContext(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fks := make(map[string]fkInfo)
	for rows.Next() {
		var localCol string
		var info fkInfo
		if err := rows.Scan(&localCol, &info.Table, &info.Column); err != nil {
			return nil, err
		}
		fks[localCol] = info
	}
	return fks, nil
}

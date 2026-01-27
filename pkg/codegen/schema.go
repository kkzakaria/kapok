package codegen

import (
	"database/sql"
	"fmt"
	
	"github.com/rs/zerolog/log"
)

// Schema represents the complete database schema
type Schema struct {
	Tables []*Table
}

// Table represents a database table
type Table struct {
	Name       string
	Schema     string
	Columns    []*Column
	PrimaryKey *PrimaryKey
}

// Column represents a table column
type Column struct {
	Name         string
	DataType     string
	IsNullable   bool
	DefaultValue *string
	Position     int
}

// PrimaryKey represents a table's primary key
type PrimaryKey struct {
	ColumnNames []string
}

// SchemaIntrospector queries PostgreSQL information_schema
// Note: The current SDK generator has limited support for composite primary keys.
// Tables with composite PKs will trigger a warning, and SDK generation may not work correctly.
// Consider using a single-column primary key (e.g., surrogate key) for best results.
type SchemaIntrospector struct {
	db *sql.DB
}

// NewSchemaIntrospector creates a new introspector
func NewSchemaIntrospector(db *sql.DB) *SchemaIntrospector {
	return &SchemaIntrospector{db: db}
}

// IntrospectSchema extracts the complete schema from PostgreSQL
func (s *SchemaIntrospector) IntrospectSchema(schemaName string) (*Schema, error) {
	if schemaName == "" {
		schemaName = "public"
	}

	schema := &Schema{
		Tables: make([]*Table, 0),
	}

	// Get all tables
	tables, err := s.getTables(schemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	// For each table, get columns and primary key
	for _, table := range tables {
		columns, err := s.getColumns(schemaName, table.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %w", table.Name, err)
		}
		table.Columns = columns

		pk, err := s.getPrimaryKey(schemaName, table.Name)
		if err != nil {
			log.Debug().
				Str("table", table.Name).
				Err(err).
				Msg("Failed to retrieve primary key, table will have no PK defined")
		}
		
		// Warn about composite primary keys (not fully supported yet)
		if pk != nil && len(pk.ColumnNames) > 1 {
			log.Warn().
				Str("table", table.Name).
				Strs("pk_columns", pk.ColumnNames).
				Msg("Composite primary key detected - SDK generation may not work correctly")
		}
		
		table.PrimaryKey = pk

		schema.Tables = append(schema.Tables, table)
	}

	return schema, nil
}

// getTables retrieves all tables from the schema
func (s *SchemaIntrospector) getTables(schemaName string) ([]*Table, error) {
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := s.db.Query(query, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]*Table, 0)
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}

		tables = append(tables, &Table{
			Name:    tableName,
			Schema:  schemaName,
			Columns: make([]*Column, 0),
		})
	}

	return tables, rows.Err()
}

// getColumns retrieves all columns for a specific table
func (s *SchemaIntrospector) getColumns(schemaName, tableName string) ([]*Column, error) {
	query := `
		SELECT 
			column_name,
			data_type,
			is_nullable,
			column_default,
			ordinal_position
		FROM information_schema.columns
		WHERE table_schema = $1
		AND table_name = $2
		ORDER BY ordinal_position
	`

	rows, err := s.db.Query(query, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make([]*Column, 0)
	for rows.Next() {
		var (
			columnName    string
			dataType      string
			isNullable    string
			columnDefault sql.NullString
			position      int
		)

		if err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault, &position); err != nil {
			return nil, err
		}

		col := &Column{
			Name:       columnName,
			DataType:   dataType,
			IsNullable: isNullable == "YES",
			Position:   position,
		}

		if columnDefault.Valid {
			col.DefaultValue = &columnDefault.String
		}

		columns = append(columns, col)
	}

	return columns, rows.Err()
}

// getPrimaryKey retrieves the primary key for a table
func (s *SchemaIntrospector) getPrimaryKey(schemaName, tableName string) (*PrimaryKey, error) {
	query := `
		SELECT a.attname
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		WHERE i.indrelid = ($1 || '.' || $2)::regclass
		AND i.indisprimary
		ORDER BY array_position(i.indkey, a.attnum)
	`

	rows, err := s.db.Query(query, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columnNames := make([]string, 0)
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, err
		}
		columnNames = append(columnNames, columnName)
	}

	if len(columnNames) == 0 {
		return nil, nil
	}

	return &PrimaryKey{
		ColumnNames: columnNames,
	}, rows.Err()
}

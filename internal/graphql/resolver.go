package graphql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/iancoleman/strcase"
	"github.com/kapok/kapok/internal/database"
)

// Resolver handles GraphQL query execution against the database
type Resolver struct {
	db *database.DB
}

// NewResolver creates a new resolver
func NewResolver(db *database.DB) *Resolver {
	return &Resolver{db: db}
}

// ResolveList returns a function that resolves a list of records from a table
func (r *Resolver) ResolveList(schemaName, tableName string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		// Basic SELECT * FROM schema.table
		// TODO: Add support for filtering, sorting, pagination based on args
		query := fmt.Sprintf(`SELECT * FROM "%s"."%s"`, schemaName, tableName)
		
		limit, _ := p.Args["limit"].(int)
		offset, _ := p.Args["offset"].(int)
		
		if limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", limit)
		}
		if offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", offset)
		}

		rows, err := r.db.QueryContext(p.Context, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		return r.scanRows(rows)
	}
}

// ResolveGet returns a function that resolves a single record by primary key
func (r *Resolver) ResolveGet(schemaName, tableName string, pkName string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		// Arg name is also camelCase usually, but pkName passed here is from DB col (snake).
		// Schema generator logic:
		// pkName arg: pkName (which is col.Name aka snake_case? No wait).
		// SchemaGenerator: pkName := col.Name.
		// Args: pkName: ...  (So arg name IS snake_case).
		// Let's check SchemaGenerator.
		// Args: graphql.FieldConfigArgument{ pkName: ... } -> pkName is col.Name (snake).
		// So ResolveGet is correct for now (uses p.Args[pkName]).
		// BUT ResolveCreate is different.
		
		id, ok := p.Args[pkName]
		if !ok {
			return nil, fmt.Errorf("argument %s is required", pkName)
		}

		query := fmt.Sprintf(`SELECT * FROM "%s"."%s" WHERE "%s" = $1`, schemaName, tableName, pkName)
		
		rows, err := r.db.QueryContext(p.Context, query, id)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		results, err := r.scanRows(rows)
		if err != nil {
			return nil, err
		}

		if len(results) == 0 {
			return nil, nil
		}
		return results[0], nil
	}
}

// ResolveCreate returns a function that inserts a new record
func (r *Resolver) ResolveCreate(schemaName, tableName string, columns []string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var cols []string
		var vals []interface{}
		var placeholders []string
		
		argIdx := 1
		for _, col := range columns {
			// Convert column name (snake_case) to argument name (camelCase)
			argName := strcase.ToLowerCamel(col)
			if val, ok := p.Args[argName]; ok {
				cols = append(cols, fmt.Sprintf(`"%s"`, col))
				vals = append(vals, val)
				placeholders = append(placeholders, fmt.Sprintf("$%d", argIdx))
				argIdx++
			}
		}

		if len(cols) == 0 {
			return nil, fmt.Errorf("no arguments provided")
		}

		query := fmt.Sprintf(
			`INSERT INTO "%s"."%s" (%s) VALUES (%s) RETURNING *`,
			schemaName, tableName,
			strings.Join(cols, ", "),
			strings.Join(placeholders, ", "),
		)

		rows, err := r.db.QueryContext(p.Context, query, vals...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		results, err := r.scanRows(rows)
		if err != nil {
			return nil, err
		}

		if len(results) == 0 {
			return nil, fmt.Errorf("failed to insert record")
		}
		return results[0], nil
	}
}

// helper to scan rows into map[string]interface{}
func (r *Resolver) scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	colNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		// Create a slice of interface{} to hold values
		values := make([]interface{}, len(colNames))
		valuePtrs := make([]interface{}, len(colNames))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		entry := make(map[string]interface{})
		for i, col := range colNames {
			val := values[i]
			// Convert column name to camelCase to match GraphQL fields
			key := strcase.ToLowerCamel(col)
			
			if b, ok := val.([]byte); ok {
				typeName := colTypes[i].DatabaseTypeName()
				switch typeName {
				case "BOOL", "bool":
					s := string(b)
					entry[key] = (s == "t" || s == "true" || s == "1" || s == "YES")
				default:
					entry[key] = string(b)
				}
			} else {
				entry[key] = val
			}
		}
		results = append(results, entry)
	}

	return results, nil
}

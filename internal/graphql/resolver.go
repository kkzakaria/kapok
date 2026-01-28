package graphql

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/iancoleman/strcase"
	"github.com/kapok/kapok/internal/database"
)

const (
	// DefaultLimit is the default number of rows returned for list queries
	DefaultLimit = 100
	// MaxLimit is the maximum number of rows that can be requested
	MaxLimit = 1000
)

// validSQLIdentifier matches valid PostgreSQL identifiers
var validSQLIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// validateIdentifier checks if a string is a valid SQL identifier
func validateIdentifier(name string) error {
	if !validSQLIdentifier.MatchString(name) {
		return fmt.Errorf("invalid identifier: %s", name)
	}
	return nil
}

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
		// Validate identifiers to prevent SQL injection
		if err := validateIdentifier(schemaName); err != nil {
			return nil, fmt.Errorf("invalid schema name")
		}
		if err := validateIdentifier(tableName); err != nil {
			return nil, fmt.Errorf("invalid table name")
		}

		query := fmt.Sprintf(`SELECT * FROM "%s"."%s"`, schemaName, tableName)

		// Apply limit with defaults and maximum cap
		limit, _ := p.Args["limit"].(int)
		offset, _ := p.Args["offset"].(int)

		if limit <= 0 {
			limit = DefaultLimit
		}
		if limit > MaxLimit {
			limit = MaxLimit
		}

		query += fmt.Sprintf(" LIMIT %d", limit)
		if offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", offset)
		}

		rows, err := r.db.QueryContext(p.Context, query)
		if err != nil {
			return nil, fmt.Errorf("query failed")
		}
		defer rows.Close()

		return r.scanRows(rows)
	}
}

// ResolveGet returns a function that resolves a single record by primary key
func (r *Resolver) ResolveGet(schemaName, tableName string, pkName string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		// Validate identifiers to prevent SQL injection
		if err := validateIdentifier(schemaName); err != nil {
			return nil, fmt.Errorf("invalid schema name")
		}
		if err := validateIdentifier(tableName); err != nil {
			return nil, fmt.Errorf("invalid table name")
		}
		if err := validateIdentifier(pkName); err != nil {
			return nil, fmt.Errorf("invalid primary key name")
		}

		id, ok := p.Args[pkName]
		if !ok {
			return nil, fmt.Errorf("argument %s is required", pkName)
		}

		query := fmt.Sprintf(`SELECT * FROM "%s"."%s" WHERE "%s" = $1`, schemaName, tableName, pkName)

		rows, err := r.db.QueryContext(p.Context, query, id)
		if err != nil {
			return nil, fmt.Errorf("query failed")
		}
		defer rows.Close()

		results, err := r.scanRows(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to read results")
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
		// Validate identifiers to prevent SQL injection
		if err := validateIdentifier(schemaName); err != nil {
			return nil, fmt.Errorf("invalid schema name")
		}
		if err := validateIdentifier(tableName); err != nil {
			return nil, fmt.Errorf("invalid table name")
		}

		var cols []string
		var vals []interface{}
		var placeholders []string

		argIdx := 1
		for _, col := range columns {
			// Validate column name
			if err := validateIdentifier(col); err != nil {
				continue // Skip invalid column names
			}
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
			return nil, fmt.Errorf("insert failed")
		}
		defer rows.Close()

		results, err := r.scanRows(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to read results")
		}

		if len(results) == 0 {
			return nil, fmt.Errorf("failed to insert record")
		}
		return results[0], nil
	}
}

// ResolveUpdate returns a function that updates an existing record by primary key
func (r *Resolver) ResolveUpdate(schemaName, tableName, pkName string, columns []string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		// Validate identifiers to prevent SQL injection
		if err := validateIdentifier(schemaName); err != nil {
			return nil, fmt.Errorf("invalid schema name")
		}
		if err := validateIdentifier(tableName); err != nil {
			return nil, fmt.Errorf("invalid table name")
		}
		if err := validateIdentifier(pkName); err != nil {
			return nil, fmt.Errorf("invalid primary key name")
		}

		// Get primary key value
		id, ok := p.Args[pkName]
		if !ok {
			return nil, fmt.Errorf("argument %s is required", pkName)
		}

		var setClauses []string
		var vals []interface{}
		argIdx := 1

		for _, col := range columns {
			// Skip primary key in SET clause
			if col == pkName {
				continue
			}
			// Validate column name
			if err := validateIdentifier(col); err != nil {
				continue
			}
			// Convert column name (snake_case) to argument name (camelCase)
			argName := strcase.ToLowerCamel(col)
			if val, ok := p.Args[argName]; ok {
				setClauses = append(setClauses, fmt.Sprintf(`"%s" = $%d`, col, argIdx))
				vals = append(vals, val)
				argIdx++
			}
		}

		if len(setClauses) == 0 {
			return nil, fmt.Errorf("no fields to update")
		}

		// Add ID as the last parameter
		vals = append(vals, id)

		query := fmt.Sprintf(
			`UPDATE "%s"."%s" SET %s WHERE "%s" = $%d RETURNING *`,
			schemaName, tableName,
			strings.Join(setClauses, ", "),
			pkName, argIdx,
		)

		rows, err := r.db.QueryContext(p.Context, query, vals...)
		if err != nil {
			return nil, fmt.Errorf("update failed")
		}
		defer rows.Close()

		results, err := r.scanRows(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to read results")
		}

		if len(results) == 0 {
			return nil, nil // Record not found
		}
		return results[0], nil
	}
}

// ResolveDelete returns a function that deletes a record by primary key
func (r *Resolver) ResolveDelete(schemaName, tableName, pkName string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		// Validate identifiers to prevent SQL injection
		if err := validateIdentifier(schemaName); err != nil {
			return nil, fmt.Errorf("invalid schema name")
		}
		if err := validateIdentifier(tableName); err != nil {
			return nil, fmt.Errorf("invalid table name")
		}
		if err := validateIdentifier(pkName); err != nil {
			return nil, fmt.Errorf("invalid primary key name")
		}

		id, ok := p.Args[pkName]
		if !ok {
			return nil, fmt.Errorf("argument %s is required", pkName)
		}

		query := fmt.Sprintf(`DELETE FROM "%s"."%s" WHERE "%s" = $1 RETURNING *`, schemaName, tableName, pkName)

		rows, err := r.db.QueryContext(p.Context, query, id)
		if err != nil {
			return nil, fmt.Errorf("delete failed")
		}
		defer rows.Close()

		results, err := r.scanRows(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to read results")
		}

		if len(results) == 0 {
			return nil, nil // Record not found
		}
		return results[0], nil
	}
}

// ResolveRelation returns a function that resolves a belongsTo relation (FK -> parent)
// Example: post.author where post has author_id FK pointing to users.id
func (r *Resolver) ResolveRelation(schemaName, foreignTable, foreignColumn, localColumn string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		// Validate identifiers
		if err := validateIdentifier(schemaName); err != nil {
			return nil, fmt.Errorf("invalid schema name")
		}
		if err := validateIdentifier(foreignTable); err != nil {
			return nil, fmt.Errorf("invalid table name")
		}
		if err := validateIdentifier(foreignColumn); err != nil {
			return nil, fmt.Errorf("invalid column name")
		}

		// Get the FK value from the parent object
		source, ok := p.Source.(map[string]interface{})
		if !ok {
			return nil, nil
		}

		// Get the local column value (e.g., authorId from post)
		localColCamel := strcase.ToLowerCamel(localColumn)
		fkValue, ok := source[localColCamel]
		if !ok || fkValue == nil {
			return nil, nil
		}

		query := fmt.Sprintf(`SELECT * FROM "%s"."%s" WHERE "%s" = $1 LIMIT 1`,
			schemaName, foreignTable, foreignColumn)

		rows, err := r.db.QueryContext(p.Context, query, fkValue)
		if err != nil {
			return nil, fmt.Errorf("relation query failed")
		}
		defer rows.Close()

		results, err := r.scanRows(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to read relation results")
		}

		if len(results) == 0 {
			return nil, nil
		}
		return results[0], nil
	}
}

// ResolveHasMany returns a function that resolves a hasMany relation (parent -> children)
// Example: user.posts where posts have user_id FK pointing to users.id
func (r *Resolver) ResolveHasMany(schemaName, childTable, childColumn, parentColumn string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		// Validate identifiers
		if err := validateIdentifier(schemaName); err != nil {
			return nil, fmt.Errorf("invalid schema name")
		}
		if err := validateIdentifier(childTable); err != nil {
			return nil, fmt.Errorf("invalid table name")
		}
		if err := validateIdentifier(childColumn); err != nil {
			return nil, fmt.Errorf("invalid column name")
		}

		// Get the parent's PK value
		source, ok := p.Source.(map[string]interface{})
		if !ok {
			return nil, nil
		}

		// Get the parent column value (e.g., id from user)
		parentColCamel := strcase.ToLowerCamel(parentColumn)
		pkValue, ok := source[parentColCamel]
		if !ok || pkValue == nil {
			return nil, nil
		}

		// Apply limit with defaults
		limit, _ := p.Args["limit"].(int)
		offset, _ := p.Args["offset"].(int)

		if limit <= 0 {
			limit = DefaultLimit
		}
		if limit > MaxLimit {
			limit = MaxLimit
		}

		query := fmt.Sprintf(`SELECT * FROM "%s"."%s" WHERE "%s" = $1 LIMIT %d`,
			schemaName, childTable, childColumn, limit)

		if offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", offset)
		}

		rows, err := r.db.QueryContext(p.Context, query, pkValue)
		if err != nil {
			return nil, fmt.Errorf("has many query failed")
		}
		defer rows.Close()

		return r.scanRows(rows)
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

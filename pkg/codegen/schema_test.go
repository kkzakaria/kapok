package codegen_test

import (
	"database/sql"
	"testing"

	"github.com/kapok/kapok/pkg/codegen"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSchemaIntrospection tests schema introspection with a real database
// This test requires PostgreSQL to be running (use `kapok dev`)
func TestSchemaIntrospection(t *testing.T) {
	// Skip if CI or no PostgreSQL available
	connStr := "postgres://kapok:kapok_dev_password@localhost:5432/kapok?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skip("PostgreSQL not available:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skip("PostgreSQL not reachable:", err)
	}

	// Create test schema
	_, err = db.Exec(`
		DROP TABLE IF EXISTS test_users CASCADE;
		CREATE TABLE test_users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	require.NoError(t, err)

	defer db.Exec("DROP TABLE IF EXISTS test_users")

	// Test introspection
	introspector := codegen.NewSchemaIntrospector(db)
	schema, err := introspector.IntrospectSchema("public")
	require.NoError(t, err)
	require.NotNil(t, schema)

	// Find test_users table
	var testUsersTable *codegen.Table
	for _, table := range schema.Tables {
		if table.Name == "test_users" {
			testUsersTable = table
			break
		}
	}

	require.NotNil(t, testUsersTable, "test_users table should be found")
	assert.Equal(t, "test_users", testUsersTable.Name)
	assert.Equal(t, "public", testUsersTable.Schema)

	// Verify columns
	assert.Len(t, testUsersTable.Columns, 5)

	// Check specific columns
	columnMap := make(map[string]*codegen.Column)
	for _, col := range testUsersTable.Columns {
		columnMap[col.Name] = col
	}

	// Check id column
	idCol := columnMap["id"]
	require.NotNil(t, idCol)
	assert.Equal(t, "integer", idCol.DataType)
	assert.False(t, idCol.IsNullable)

	// Check username column
	usernameCol := columnMap["username"]
	require.NotNil(t, usernameCol)
	assert.Equal(t, "character varying", usernameCol.DataType)
	assert.False(t, usernameCol.IsNullable)

	// Check is_active column
	isActiveCol := columnMap["is_active"]
	require.NotNil(t, isActiveCol)
	assert.Equal(t, "boolean", isActiveCol.DataType)
	assert.True(t, isActiveCol.IsNullable) // Has default but still nullable

	// Check primary key
	require.NotNil(t, testUsersTable.PrimaryKey)
	assert.Equal(t, []string{"id"}, testUsersTable.PrimaryKey.ColumnNames)
}

func TestSchemaIntrospectionEmpty(t *testing.T) {
	connStr := "postgres://kapok:kapok_dev_password@localhost:5432/kapok?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skip("PostgreSQL not available:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skip("PostgreSQL not reachable:", err)
	}

	// Create empty schema
	_, err = db.Exec("CREATE SCHEMA IF NOT EXISTS test_empty")
	require.NoError(t, err)
	defer db.Exec("DROP SCHEMA IF EXISTS test_empty CASCADE")

	introspector := codegen.NewSchemaIntrospector(db)
	schema, err := introspector.IntrospectSchema("test_empty")
	require.NoError(t, err)
	assert.Empty(t, schema.Tables)
}

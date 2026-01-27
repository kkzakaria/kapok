package codegen_test

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/kapok/kapok/pkg/codegen"
	"github.com/kapok/kapok/pkg/codegen/typescript"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDB  *sql.DB
	cleanup func()
)

func TestMain(m *testing.M) {
	// Skip docker setup in short mode
	// We can't use testing.Short() before m.Run(), so we use a simple check
	var runDocker = true
	for _, arg := range os.Args {
		if arg == "-test.short" || arg == "--test.short" {
			runDocker = false
			break
		}
	}

	if !runDocker {
		os.Exit(m.Run())
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		panic(fmt.Sprintf("Could not connect to docker: %s", err))
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_DB=test",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		panic(fmt.Sprintf("Could not start resource: %s", err))
	}

	cleanup = func() {
		if err := pool.Purge(resource); err != nil {
			panic(fmt.Sprintf("Could not purge resource: %s", err))
		}
	}

	connStr := fmt.Sprintf(
		"host=localhost port=%s user=postgres password=postgres dbname=test sslmode=disable",
		resource.GetPort("5432/tcp"),
	)

	if err = pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("postgres", connStr)
		if err != nil {
			return err
		}
		return testDB.Ping()
	}); err != nil {
		cleanup()
		panic(fmt.Sprintf("Could not connect to database: %s", err))
	}

	code := m.Run()

	if testDB != nil {
		testDB.Close()
	}
	cleanup()

	os.Exit(code)
}

func TestE2E_FullSDKGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Create test schema
	setupTestSchema(t, testDB)
	defer cleanupTestSchema(t, testDB)

	// Introspect schema
	introspector := codegen.NewSchemaIntrospector(testDB)
	schema, err := introspector.IntrospectSchema("public")
	require.NoError(t, err)
	require.NotEmpty(t, schema.Tables)

	// Generate SDK
	tmpDir, err := ioutil.TempDir("", "sdk-e2e-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	clientGen := typescript.NewClientGenerator()
	err = clientGen.WriteSDK(tmpDir, schema, "test-sdk")
	require.NoError(t, err)

	// Verify file structure
	verifySDKStructure(t, tmpDir)

	// Verify file contents
	verifySDKContents(t, tmpDir, schema)
}

func TestE2E_SnakeCaseToCamelCase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Create table with snake_case columns
	_, err := testDB.Exec(`
		CREATE TABLE IF NOT EXISTS test_users (
			user_id SERIAL PRIMARY KEY,
			full_name VARCHAR(255) NOT NULL,
			email_address VARCHAR(255) NOT NULL,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)
	defer testDB.Exec("DROP TABLE IF EXISTS test_users")

	// Introspect and generate
	introspector := codegen.NewSchemaIntrospector(testDB)
	schema, err := introspector.IntrospectSchema("public")
	require.NoError(t, err)

	typeMapper := typescript.NewTypeMapper()
	
	// Find test_users table
	var testUsersTable *codegen.Table
	for _, table := range schema.Tables {
		if table.Name == "test_users" {
			testUsersTable = table
			break
		}
	}
	require.NotNil(t, testUsersTable)

	// Verify interface generation with camelCase
	interfaceCode := typeMapper.GenerateInterface(testUsersTable)
	
	assert.Contains(t, interfaceCode, "userId: number")
	assert.Contains(t, interfaceCode, "fullName: string")
	assert.Contains(t, interfaceCode, "emailAddress: string")
	assert.Contains(t, interfaceCode, "isActive: boolean")
	assert.Contains(t, interfaceCode, "createdAt: Date")

	// Should NOT contain snake_case
	assert.NotContains(t, interfaceCode, "user_id")
	assert.NotContains(t, interfaceCode, "full_name")
}

func TestE2E_NullableTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Create table with nullable and non-nullable fields
	_, err := testDB.Exec(`
		CREATE TABLE IF NOT EXISTS nullable_test (
			id SERIAL PRIMARY KEY,
			required_field VARCHAR(100) NOT NULL,
			optional_field VARCHAR(100),
			nullable_int INTEGER
		)
	`)
	require.NoError(t, err)
	defer testDB.Exec("DROP TABLE IF EXISTS nullable_test")

	introspector := codegen.NewSchemaIntrospector(testDB)
	schema, err := introspector.IntrospectSchema("public")
	require.NoError(t, err)

	typeMapper := typescript.NewTypeMapper()
	
	var table *codegen.Table
	for _, t := range schema.Tables {
		if t.Name == "nullable_test" {
			table = t
			break
		}
	}
	require.NotNil(t, table)

	interfaceCode := typeMapper.GenerateInterface(table)
	
	// Required fields should not have | null
	assert.Contains(t, interfaceCode, "requiredField: string;")
	
	// Nullable fields should have | null
	assert.Contains(t, interfaceCode, "optionalField: string | null;")
	assert.Contains(t, interfaceCode, "nullableInt: number | null;")
}

func TestE2E_PrimaryKeyDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Test with UUID primary key
	_, err := testDB.Exec(`
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
		CREATE TABLE IF NOT EXISTS uuid_test (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(100) NOT NULL
		)
	`)
	require.NoError(t, err)
	defer testDB.Exec("DROP TABLE IF EXISTS uuid_test")

	introspector := codegen.NewSchemaIntrospector(testDB)
	schema, err := introspector.IntrospectSchema("public")
	require.NoError(t, err)

	var table *codegen.Table
	for _, t := range schema.Tables {
		if t.Name == "uuid_test" {
			table = t
			break
		}
	}
	require.NotNil(t, table)

	// Verify primary key is detected
	require.NotNil(t, table.PrimaryKey)
	assert.Equal(t, []string{"id"}, table.PrimaryKey.ColumnNames)

	// Verify CRUD functions use correct type
	crudGen := typescript.NewCRUDGenerator()
	getByIdFunc := crudGen.GenerateGetByIdFunction(table)
	
	// UUID should map to string type
	assert.Contains(t, getByIdFunc, "id: string")
}

func TestE2E_AutoGeneratedFieldExclusion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	_, err := testDB.Exec(`
		CREATE TABLE IF NOT EXISTS auto_fields_test (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)
	defer testDB.Exec("DROP TABLE IF EXISTS auto_fields_test")

	introspector := codegen.NewSchemaIntrospector(testDB)
	schema, err := introspector.IntrospectSchema("public")
	require.NoError(t, err)

	typeMapper := typescript.NewTypeMapper()
	
	var table *codegen.Table
	for _, t := range schema.Tables {
		if t.Name == "auto_fields_test" {
			table = t
			break
		}
	}
	require.NotNil(t, table)

	createInput := typeMapper.GenerateCreateInput(table)
	
	// Auto-generated fields should be excluded from CreateInput
	assert.NotContains(t, createInput, "id:")
	assert.NotContains(t, createInput, "createdAt:")
	assert.NotContains(t, createInput, "updatedAt:")
	
	// Name should be included
	assert.Contains(t, createInput, "name:")
}

func TestE2E_MultipleTablesWithRelations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Create related tables
	_, err := testDB.Exec(`
		CREATE TABLE IF NOT EXISTS authors (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL
		);
		
		CREATE TABLE IF NOT EXISTS books (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			author_id INTEGER REFERENCES authors(id)
		);
	`)
	require.NoError(t, err)
	defer func() {
		testDB.Exec("DROP TABLE IF EXISTS books")
		testDB.Exec("DROP TABLE IF EXISTS authors")
	}()

	introspector := codegen.NewSchemaIntrospector(testDB)
	schema, err := introspector.IntrospectSchema("public")
	require.NoError(t, err)

	// Should have both tables
	assert.GreaterOrEqual(t, len(schema.Tables), 2)

	// Generate complete SDK
	tmpDir, err := ioutil.TempDir("", "sdk-relations-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	clientGen := typescript.NewClientGenerator()
	err = clientGen.WriteSDK(tmpDir, schema, "test-sdk")
	require.NoError(t, err)

	// Verify client has methods for both tables
	clientFile, err := ioutil.ReadFile(filepath.Join(tmpDir, "src", "client.ts"))
	require.NoError(t, err)
	
	clientContent := string(clientFile)
	assert.Contains(t, clientContent, "authors = {")
	assert.Contains(t, clientContent, "books = {")
}

// Helper functions

func setupTestSchema(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS posts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title VARCHAR(255) NOT NULL,
			content TEXT,
			user_id INTEGER REFERENCES users(id),
			published BOOLEAN DEFAULT false,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	require.NoError(t, err)
}

func cleanupTestSchema(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`
		DROP TABLE IF EXISTS posts;
		DROP TABLE IF EXISTS users;
		DROP TABLE IF EXISTS test_users;
		DROP TABLE IF EXISTS nullable_test;
		DROP TABLE IF EXISTS uuid_test;
		DROP TABLE IF EXISTS auto_fields_test;
		DROP TABLE IF EXISTS books;
		DROP TABLE IF EXISTS authors;
	`)
	require.NoError(t, err)
}

func verifySDKStructure(t *testing.T, sdkDir string) {
	expectedFiles := []string{
		"package.json",
		"tsconfig.json",
		"README.md",
		filepath.Join("src", "index.ts"),
		filepath.Join("src", "client.ts"),
		filepath.Join("src", "types", "index.ts"),
		filepath.Join("src", "api", "index.ts"),
	}

	for _, file := range expectedFiles {
		path := filepath.Join(sdkDir, file)
		_, err := os.Stat(path)
		assert.NoError(t, err, "File should exist: %s", file)
	}
}

func verifySDKContents(t *testing.T, sdkDir string, schema *codegen.Schema) {
	// Verify types file contains all tables
	typesFile, err := ioutil.ReadFile(filepath.Join(sdkDir, "src", "types", "index.ts"))
	require.NoError(t, err)
	
	typesContent := string(typesFile)
	for _, table := range schema.Tables {
		typeMapper := typescript.NewTypeMapper()
		typeName := typeMapper.ToTypeName(table.Name)
		assert.Contains(t, typesContent, fmt.Sprintf("export interface %s", typeName))
	}

	// Verify client file contains all CRUD methods
	clientFile, err := ioutil.ReadFile(filepath.Join(sdkDir, "src", "client.ts"))
	require.NoError(t, err)
	
	clientContent := string(clientFile)
	assert.Contains(t, clientContent, "export class KapokClient")
	
	// Verify package.json
	packageFile, err := ioutil.ReadFile(filepath.Join(sdkDir, "package.json"))
	require.NoError(t, err)
	
	packageContent := string(packageFile)
	assert.Contains(t, packageContent, "typescript")
	assert.Contains(t, packageContent, "test-sdk")
}

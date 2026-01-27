package typescript

import (
	"strings"
	"testing"

	"github.com/kapok/kapok/pkg/codegen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCRUDGenerator_GenerateCreateFunction(t *testing.T) {
	generator := NewCRUDGenerator()

	table := &codegen.Table{
		Name:   "users",
		Schema: "public",
		Columns: []*codegen.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
			{Name: "email", DataType: "varchar", IsNullable: false},
			{Name: "name", DataType: "varchar", IsNullable: true},
		},
	}

	result := generator.GenerateCreateFunction(table)

	// Verify function signature
	assert.Contains(t, result, "export async function createUsers(")
	assert.Contains(t, result, "baseUrl: string, input: CreateUsersInput")
	assert.Contains(t, result, "): Promise<Users>")

	// Verify HTTP request
	assert.Contains(t, result, "method: 'POST'")
	assert.Contains(t, result, "headers: { 'Content-Type': 'application/json' }")
	assert.Contains(t, result, "body: JSON.stringify(input)")

	// Verify error handling
	assert.Contains(t, result, "if (!response.ok)")
	assert.Contains(t, result, "throw new Error")
}

func TestCRUDGenerator_GenerateGetByIdFunction(t *testing.T) {
	generator := NewCRUDGenerator()

	defaultVal := "nextval('users_id_seq'::regclass)"
	table := &codegen.Table{
		Name:   "users",
		Schema: "public",
		Columns: []*codegen.Column{
			{Name: "id", DataType: "integer", IsNullable: false, DefaultValue: &defaultVal},
			{Name: "email", DataType: "varchar", IsNullable: false},
		},
		PrimaryKey: &codegen.PrimaryKey{
			ColumnNames: []string{"id"},
		},
	}

	result := generator.GenerateGetByIdFunction(table)

	// Verify function signature with correct PK type
	assert.Contains(t, result, "export async function getUsersById(")
	assert.Contains(t, result, "baseUrl: string, id: number")
	assert.Contains(t, result, "): Promise<Users>")

	// Verify HTTP request
	assert.Contains(t, result, "fetch(`${baseUrl}/users/${id}`)")
	assert.Contains(t, result, "return response.json()")
}

func TestCRUDGenerator_GenerateListFunction(t *testing.T) {
	generator := NewCRUDGenerator()

	table := &codegen.Table{
		Name:   "products",
		Schema: "public",
		Columns: []*codegen.Column{
			{Name: "id", DataType: "uuid", IsNullable: false},
			{Name: "name", DataType: "text", IsNullable: false},
		},
	}

	result := generator.GenerateListFunction(table)

	// Verify function signature
	assert.Contains(t, result, "export async function listProducts(")
	assert.Contains(t, result, "options?: { limit?: number; offset?: number }")
	assert.Contains(t, result, "): Promise<Products[]>")

	// Verify pagination support
	assert.Contains(t, result, "URLSearchParams()")
	assert.Contains(t, result, "params.append('limit'")
	assert.Contains(t, result, "params.append('offset'")
}

func TestCRUDGenerator_GenerateUpdateFunction(t *testing.T) {
	generator := NewCRUDGenerator()

	table := &codegen.Table{
		Name:   "users",
		Schema: "public",
		Columns: []*codegen.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
			{Name: "email", DataType: "varchar", IsNullable: false},
		},
		PrimaryKey: &codegen.PrimaryKey{
			ColumnNames: []string{"id"},
		},
	}

	result := generator.GenerateUpdateFunction(table)

	// Verify function signature
	assert.Contains(t, result, "export async function updateUsers(")
	assert.Contains(t, result, "baseUrl: string, id: number, input: UpdateUsersInput")
	assert.Contains(t, result, "): Promise<Users>")

	// Verify HTTP request
	assert.Contains(t, result, "method: 'PUT'")
	assert.Contains(t, result, "fetch(`${baseUrl}/users/${id}`")
}

func TestCRUDGenerator_GenerateDeleteFunction(t *testing.T) {
	generator := NewCRUDGenerator()

	table := &codegen.Table{
		Name:   "users",
		Schema: "public",
		Columns: []*codegen.Column{
			{Name: "id", DataType: "uuid", IsNullable: false},
		},
		PrimaryKey: &codegen.PrimaryKey{
			ColumnNames: []string{"id"},
		},
	}

	result := generator.GenerateDeleteFunction(table)

	// Verify function signature
	assert.Contains(t, result, "export async function deleteUsers(")
	assert.Contains(t, result, "baseUrl: string, id: string") // UUID maps to string
	assert.Contains(t, result, "): Promise<void>")

	// Verify HTTP request
	assert.Contains(t, result, "method: 'DELETE'")
	assert.Contains(t, result, "fetch(`${baseUrl}/users/${id}`")
}

func TestCRUDGenerator_GenerateAllCRUD(t *testing.T) {
	generator := NewCRUDGenerator()

	table := &codegen.Table{
		Name:   "users",
		Schema: "public",
		Columns: []*codegen.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
			{Name: "email", DataType: "varchar", IsNullable: false},
		},
		PrimaryKey: &codegen.PrimaryKey{
			ColumnNames: []string{"id"},
		},
	}

	result := generator.GenerateAllCRUD(table)

	// Verify all CRUD functions are present
	assert.Contains(t, result, "createUsers(")
	assert.Contains(t, result, "getUsersById(")
	assert.Contains(t, result, "listUsers(")
	assert.Contains(t, result, "updateUsers(")
	assert.Contains(t, result, "deleteUsers(")
}

func TestCRUDGenerator_Pluralize(t *testing.T) {
	generator := NewCRUDGenerator()

	tests := []struct {
		input    string
		expected string
	}{
		{"User", "Users"},
		{"Category", "Categories"},
		{"Box", "Boxes"},
		{"Address", "Addresses"},
		{"Person", "Persons"}, // Simple pluralization
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := generator.pluralize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCRUDGenerator_TypeScriptSyntax(t *testing.T) {
	generator := NewCRUDGenerator()

	table := &codegen.Table{
		Name:   "test_table",
		Schema: "public",
		Columns: []*codegen.Column{
			{Name: "id", DataType: "integer", IsNullable: false},
		},
		PrimaryKey: &codegen.PrimaryKey{ColumnNames: []string{"id"}},
	}

	result := generator.GenerateAllCRUD(table)

	// Verify proper TypeScript syntax
	require.NotEmpty(t, result)

	// Check for proper async/await syntax
	assert.True(t, strings.Contains(result, "async function"))
	assert.True(t, strings.Contains(result, "await fetch"))

	// Check for proper Promise types
	assert.True(t, strings.Contains(result, "Promise<"))

	// Check for proper export statements
	count := strings.Count(result, "export async function")
	assert.Equal(t, 5, count, "Should have 5 exported functions (create, get, list, update, delete)")
}

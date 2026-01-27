package react

import (
	"fmt"
	"testing"

	"github.com/kapok/kapok/pkg/codegen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateListHook(t *testing.T) {
	gen := NewHooksGenerator()

	table := &codegen.Table{
		Name: "users",
		Columns: []*codegen.Column{
			{Name: "id", DataType: "integer"},
			{Name: "email", DataType: "varchar"},
		},
	}

	result := gen.GenerateListHook(table)

	// Verify hook name
	assert.Contains(t, result, "export function useListUsers")
	
	// Verify React Query usage
	assert.Contains(t, result, "useQuery")
	assert.Contains(t, result, "queryKey: ['users', 'list', options]")
	assert.Contains(t, result, "queryFn: () => client.users.list(options)")
	
	// Verify client usage
	assert.Contains(t, result, "useKapokClient()")
}

func TestGenerateGetByIdHook(t *testing.T) {
	gen := NewHooksGenerator()

	tests := []struct {
		name     string
		table    *codegen.Table
		wantType string
	}{
		{
			name: "integer primary key",
			table: &codegen.Table{
				Name: "users",
				Columns: []*codegen.Column{
					{Name: "id", DataType: "integer"},
				},
				PrimaryKey: &codegen.PrimaryKey{ColumnNames: []string{"id"}},
			},
			wantType: "number",
		},
		{
			name: "uuid primary key",
			table: &codegen.Table{
				Name: "posts",
				Columns: []*codegen.Column{
					{Name: "id", DataType: "uuid"},
				},
				PrimaryKey: &codegen.PrimaryKey{ColumnNames: []string{"id"}},
			},
			wantType: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.GenerateGetByIdHook(tt.table)

			// Verify hook signature
			assert.Contains(t, result, fmt.Sprintf("(id: %s)", tt.wantType))
			
			// Verify query key
			assert.Contains(t, result, fmt.Sprintf("queryKey: ['%s', 'detail', id]", tt.table.Name))
			
			// Verify enabled check
			assert.Contains(t, result, "enabled: !!id")
		})
	}
}

func TestGenerateCreateMutationHook(t *testing.T) {
	gen := NewHooksGenerator()

	table := &codegen.Table{
		Name: "users",
		Columns: []*codegen.Column{
			{Name: "email", DataType: "varchar"},
		},
	}

	result := gen.GenerateCreateMutationHook(table)

	// Verify hook name
	assert.Contains(t, result, "export function useCreateUsers")
	
	// Verify mutation setup
	assert.Contains(t, result, "useMutation")
	assert.Contains(t, result, "mutationFn: (input: CreateUsersInput)")
	assert.Contains(t, result, "client.users.create(input)")
	
	// Verify cache invalidation
	assert.Contains(t, result, "useQueryClient()")
	assert.Contains(t, result, "queryClient.invalidateQueries({ queryKey: ['users'] })")
}

func TestGenerateUpdateMutationHook(t *testing.T) {
	gen := NewHooksGenerator()

	table := &codegen.Table{
		Name: "users",
		Columns: []*codegen.Column{
			{Name: "id", DataType: "integer"},
			{Name: "email", DataType: "varchar"},
		},
		PrimaryKey: &codegen.PrimaryKey{ColumnNames: []string{"id"}},
	}

	result := gen.GenerateUpdateMutationHook(table)

	// Verify hook name
	assert.Contains(t, result, "export function useUpdateUsers")
	
	// Verify mutation with both id and input
	assert.Contains(t, result, "{ id, input }: { id: number; input: UpdateUsersInput }")
	
	// Verify double cache invalidation (list + detail)
	assert.Contains(t, result, "queryClient.invalidateQueries({ queryKey: ['users'] })")
	assert.Contains(t, result, "queryClient.invalidateQueries({ queryKey: ['users', 'detail', variables.id] })")
}

func TestGenerateDeleteMutationHook(t *testing.T) {
	gen := NewHooksGenerator()

	table := &codegen.Table{
		Name: "users",
		Columns: []*codegen.Column{
			{Name: "id", DataType: "uuid"},
		},
		PrimaryKey: &codegen.PrimaryKey{ColumnNames: []string{"id"}},
	}

	result := gen.GenerateDeleteMutationHook(table)

	// Verify hook name
	assert.Contains(t, result, "export function useDeleteUsers")
	
	// Verify mutation signature with string type (uuid)
	assert.Contains(t, result, "mutationFn: (id: string)")
	
	// Verify cache invalidation
	assert.Contains(t, result, "queryClient.invalidateQueries({ queryKey: ['users'] })")
	assert.Contains(t, result, "queryClient.invalidateQueries({ queryKey: ['users', 'detail', id] })")
}

func TestGenerateAllHooks(t *testing.T) {
	gen := NewHooksGenerator()

	table := &codegen.Table{
		Name: "users",
		Columns: []*codegen.Column{
			{Name: "id", DataType: "integer"},
			{Name: "email", DataType: "varchar"},
		},
		PrimaryKey: &codegen.PrimaryKey{ColumnNames: []string{"id"}},
	}

	result := gen.GenerateAllHooks(table)

	// Verify imports
	assert.Contains(t, result, "import { useQuery, useMutation, useQueryClient }")
	assert.Contains(t, result, "import { useKapokClient }")
	assert.Contains(t, result, "import type { Users, CreateUsersInput, UpdateUsersInput }")
	
	// Verify all hooks are generated
	assert.Contains(t, result, "export function useListUsers")
	assert.Contains(t, result, "export function useUsersById")
	assert.Contains(t, result, "export function useCreateUsers")
	assert.Contains(t, result, "export function useUpdateUsers")
	assert.Contains(t, result, "export function useDeleteUsers")
}

func TestGetPrimaryKeyType(t *testing.T) {
	gen := NewHooksGenerator()

	tests := []struct {
		name     string
		table    *codegen.Table
		expected string
	}{
		{
			name: "integer pk",
			table: &codegen.Table{
				Columns:    []*codegen.Column{{Name: "id", DataType: "integer"}},
				PrimaryKey: &codegen.PrimaryKey{ColumnNames: []string{"id"}},
			},
			expected: "number",
		},
		{
			name: "uuid pk",
			table: &codegen.Table{
				Columns:    []*codegen.Column{{Name: "id", DataType: "uuid"}},
				PrimaryKey: &codegen.PrimaryKey{ColumnNames: []string{"id"}},
			},
			expected: "string",
		},
		{
			name:     "no pk",
			table:    &codegen.Table{Columns: []*codegen.Column{}},
			expected: "number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.getPrimaryKeyType(tt.table)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypeMapper_ToHookName(t *testing.T) {
	tm := NewTypeMapper()

	tests := []struct {
		tableName string
		operation string
		expected  string
	}{
		{"users", "List", "useListUsers"},
		{"blog_posts", "Create", "useCreateBlogPosts"},
		{"user_profiles", "Update", "useUpdateUserProfiles"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tm.ToHookName(tt.tableName, tt.operation)
			require.Equal(t, tt.expected, result)
		})
	}
}

package react

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/kapok/kapok/pkg/codegen"
)

// HooksGenerator generates React Query hooks for database tables
type HooksGenerator struct {
	typeMapper *TypeMapper
}

// NewHooksGenerator creates a new hooks generator
func NewHooksGenerator() *HooksGenerator {
	return &HooksGenerator{
		typeMapper: NewTypeMapper(),
	}
}

// TypeMapper helps convert types for React/TypeScript
type TypeMapper struct{}

func NewTypeMapper() *TypeMapper {
	return &TypeMapper{}
}

// ToTypeName converts table name to TypeScript type name (PascalCase)
func (tm *TypeMapper) ToTypeName(tableName string) string {
	return strcase.ToCamel(tableName)
}

// ToHookName converts table name to hook name (camelCase with 'use' prefix)
func (tm *TypeMapper) ToHookName(tableName string, operation string) string {
	typeName := tm.ToTypeName(tableName)
	return fmt.Sprintf("use%s%s", operation, typeName)
}

// GenerateListHook generates a React Query hook for listing records
func (g *HooksGenerator) GenerateListHook(table *codegen.Table) string {
	var sb strings.Builder

	hookName := g.typeMapper.ToHookName(table.Name, "List")

	sb.WriteString(fmt.Sprintf("export function %s(options?: { limit?: number; offset?: number }) {\n", hookName))
	sb.WriteString("  const client = useKapokClient();\n")
	sb.WriteString("  \n")
	sb.WriteString("  return useQuery({\n")
	sb.WriteString(fmt.Sprintf("    queryKey: ['%s', 'list', options],\n", table.Name))
	sb.WriteString(fmt.Sprintf("    queryFn: () => client.%s.list(options),\n", strcase.ToLowerCamel(table.Name)))
	sb.WriteString("  });\n")
	sb.WriteString("}\n")

	return sb.String()
}

// GenerateGetByIdHook generates a React Query hook for fetching by ID
func (g *HooksGenerator) GenerateGetByIdHook(table *codegen.Table) string {
	var sb strings.Builder

	typeName := g.typeMapper.ToTypeName(table.Name)
	hookName := fmt.Sprintf("use%sById", typeName)
	
	// Determine primary key type
	pkType := g.getPrimaryKeyType(table)

	sb.WriteString(fmt.Sprintf("export function %s(id: %s) {\n", hookName, pkType))
	sb.WriteString("  const client = useKapokClient();\n")
	sb.WriteString("  \n")
	sb.WriteString("  return useQuery({\n")
	sb.WriteString(fmt.Sprintf("    queryKey: ['%s', 'detail', id],\n", table.Name))
	sb.WriteString(fmt.Sprintf("    queryFn: () => client.%s.getById(id),\n", strcase.ToLowerCamel(table.Name)))
	sb.WriteString("    enabled: !!id,\n")
	sb.WriteString("  });\n")
	sb.WriteString("}\n")

	return sb.String()
}

// GenerateCreateMutationHook generates a mutation hook for creating records
func (g *HooksGenerator) GenerateCreateMutationHook(table *codegen.Table) string {
	var sb strings.Builder

	typeName := g.typeMapper.ToTypeName(table.Name)
	hookName := g.typeMapper.ToHookName(table.Name, "Create")

	sb.WriteString(fmt.Sprintf("export function %s() {\n", hookName))
	sb.WriteString("  const client = useKapokClient();\n")
	sb.WriteString("  const queryClient = useQueryClient();\n")
	sb.WriteString("  \n")
	sb.WriteString("  return useMutation({\n")
	sb.WriteString(fmt.Sprintf("    mutationFn: (input: Create%sInput) => client.%s.create(input),\n", 
		typeName, strcase.ToLowerCamel(table.Name)))
	sb.WriteString("    onSuccess: () => {\n")
	sb.WriteString(fmt.Sprintf("      queryClient.invalidateQueries({ queryKey: ['%s'] });\n", table.Name))
	sb.WriteString("    },\n")
	sb.WriteString("  });\n")
	sb.WriteString("}\n")

	return sb.String()
}

// GenerateUpdateMutationHook generates a mutation hook for updating records
func (g *HooksGenerator) GenerateUpdateMutationHook(table *codegen.Table) string {
	var sb strings.Builder

	typeName := g.typeMapper.ToTypeName(table.Name)
	hookName := g.typeMapper.ToHookName(table.Name, "Update")
	pkType := g.getPrimaryKeyType(table)

	sb.WriteString(fmt.Sprintf("export function %s() {\n", hookName))
	sb.WriteString("  const client = useKapokClient();\n")
	sb.WriteString("  const queryClient = useQueryClient();\n")
	sb.WriteString("  \n")
	sb.WriteString("  return useMutation({\n")
	sb.WriteString(fmt.Sprintf("    mutationFn: ({ id, input }: { id: %s; input: Update%sInput }) => client.%s.update(id, input),\n",
		pkType, typeName, strcase.ToLowerCamel(table.Name)))
	sb.WriteString("    onSuccess: (_, variables) => {\n")
	sb.WriteString(fmt.Sprintf("      queryClient.invalidateQueries({ queryKey: ['%s'] });\n", table.Name))
	sb.WriteString(fmt.Sprintf("      queryClient.invalidateQueries({ queryKey: ['%s', 'detail', variables.id] });\n", table.Name))
	sb.WriteString("    },\n")
	sb.WriteString("  });\n")
	sb.WriteString("}\n")

	return sb.String()
}

// GenerateDeleteMutationHook generates a mutation hook for deleting records
func (g *HooksGenerator) GenerateDeleteMutationHook(table *codegen.Table) string {
	var sb strings.Builder

	hookName := g.typeMapper.ToHookName(table.Name, "Delete")
	pkType := g.getPrimaryKeyType(table)

	sb.WriteString(fmt.Sprintf("export function %s() {\n", hookName))
	sb.WriteString("  const client = useKapokClient();\n")
	sb.WriteString("  const queryClient = useQueryClient();\n")
	sb.WriteString("  \n")
	sb.WriteString("  return useMutation({\n")
	sb.WriteString(fmt.Sprintf("    mutationFn: (id: %s) => client.%s.delete(id),\n",
		pkType, strcase.ToLowerCamel(table.Name)))
	sb.WriteString("    onSuccess: (_, id) => {\n")
	sb.WriteString(fmt.Sprintf("      queryClient.invalidateQueries({ queryKey: ['%s'] });\n", table.Name))
	sb.WriteString(fmt.Sprintf("      queryClient.invalidateQueries({ queryKey: ['%s', 'detail', id] });\n", table.Name))
	sb.WriteString("    },\n")
	sb.WriteString("  });\n")
	sb.WriteString("}\n")

	return sb.String()
}

// GenerateAllHooks generates all hooks for a table
func (g *HooksGenerator) GenerateAllHooks(table *codegen.Table) string {
	var sb strings.Builder

	// Add imports
	sb.WriteString("import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';\n")
	sb.WriteString("import { useKapokClient } from '../provider';\n")
	sb.WriteString(fmt.Sprintf("import type { %s, Create%sInput, Update%sInput } from '../types';\n\n",
		g.typeMapper.ToTypeName(table.Name),
		g.typeMapper.ToTypeName(table.Name),
		g.typeMapper.ToTypeName(table.Name)))

	// Generate hooks
	sb.WriteString(g.GenerateListHook(table))
	sb.WriteString("\n")
	sb.WriteString(g.GenerateGetByIdHook(table))
	sb.WriteString("\n")
	sb.WriteString(g.GenerateCreateMutationHook(table))
	sb.WriteString("\n")
	sb.WriteString(g.GenerateUpdateMutationHook(table))
	sb.WriteString("\n")
	sb.WriteString(g.GenerateDeleteMutationHook(table))

	return sb.String()
}

// getPrimaryKeyType returns the TypeScript type for the primary key
func (g *HooksGenerator) getPrimaryKeyType(table *codegen.Table) string {
	if table.PrimaryKey == nil || len(table.PrimaryKey.ColumnNames) == 0 {
		return "number" // default
	}

	pkColName := table.PrimaryKey.ColumnNames[0]
	for _, col := range table.Columns {
		if col.Name == pkColName {
			// Simple type mapping
			dataType := strings.ToLower(col.DataType)
			if strings.Contains(dataType, "uuid") || strings.Contains(dataType, "char") || strings.Contains(dataType, "text") {
				return "string"
			}
			return "number"
		}
	}
	return "number"
}

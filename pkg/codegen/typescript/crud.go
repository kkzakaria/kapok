package typescript

import (
	"fmt"
	"strings"

	"github.com/kapok/kapok/pkg/codegen"
)

// CRUDGenerator generates TypeScript CRUD functions for database tables
type CRUDGenerator struct {
	typeMapper *TypeMapper
}

// NewCRUDGenerator creates a new CRUD generator
func NewCRUDGenerator() *CRUDGenerator {
	return &CRUDGenerator{
		typeMapper: NewTypeMapper(),
	}
}

// GenerateCreateFunction generates an async create function
func (g *CRUDGenerator) GenerateCreateFunction(table *codegen.Table) string {
	var sb strings.Builder

	typeName := g.typeMapper.ToTypeName(table.Name)
	functionName := fmt.Sprintf("create%s", typeName)

	sb.WriteString(fmt.Sprintf("export async function %s(", functionName))
	sb.WriteString(fmt.Sprintf("baseUrl: string, input: Create%sInput", typeName))
	sb.WriteString(fmt.Sprintf("): Promise<%s> {\n", typeName))
	sb.WriteString(fmt.Sprintf("  const response = await fetch(`${baseUrl}/%s`, {\n", table.Name))
	sb.WriteString("    method: 'POST',\n")
	sb.WriteString("    headers: { 'Content-Type': 'application/json' },\n")
	sb.WriteString("    body: JSON.stringify(input),\n")
	sb.WriteString("  });\n")
	sb.WriteString("  if (!response.ok) {\n")
	sb.WriteString("    const error = await response.json().catch(() => ({ message: response.statusText }));\n")
	sb.WriteString("    throw new Error(error.message || `Failed to create: ${response.statusText}`);\n")
	sb.WriteString("  }\n")
	sb.WriteString("  return response.json();\n")
	sb.WriteString("}\n")

	return sb.String()
}

// GenerateGetByIdFunction generates a get by ID function
func (g *CRUDGenerator) GenerateGetByIdFunction(table *codegen.Table) string {
	var sb strings.Builder

	typeName := g.typeMapper.ToTypeName(table.Name)
	functionName := fmt.Sprintf("get%sById", typeName)

	// Determine primary key type
	pkType := g.getPrimaryKeyTSType(table)

	sb.WriteString(fmt.Sprintf("export async function %s(", functionName))
	sb.WriteString(fmt.Sprintf("baseUrl: string, id: %s", pkType))
	sb.WriteString(fmt.Sprintf("): Promise<%s> {\n", typeName))
	sb.WriteString(fmt.Sprintf("  const response = await fetch(`${baseUrl}/%s/${id}`);\n", table.Name))
	sb.WriteString("  if (!response.ok) {\n")
	sb.WriteString("    const error = await response.json().catch(() => ({ message: response.statusText }));\n")
	sb.WriteString("    throw new Error(error.message || `Failed to fetch: ${response.statusText}`);\n")
	sb.WriteString("  }\n")
	sb.WriteString("  return response.json();\n")
	sb.WriteString("}\n")

	return sb.String()
}

// GenerateListFunction generates a list all function with optional pagination
func (g *CRUDGenerator) GenerateListFunction(table *codegen.Table) string {
	var sb strings.Builder

	typeName := g.typeMapper.ToTypeName(table.Name)
	// Use the type name directly as table names are typically already plural
	functionName := fmt.Sprintf("list%s", typeName)

	sb.WriteString(fmt.Sprintf("export async function %s(", functionName))
	sb.WriteString("baseUrl: string, options?: { limit?: number; offset?: number }")
	sb.WriteString(fmt.Sprintf("): Promise<%s[]> {\n", typeName))
	sb.WriteString("  let url = `${baseUrl}/" + table.Name + "`;\n")
	sb.WriteString("  if (options) {\n")
	sb.WriteString("    const params = new URLSearchParams();\n")
	sb.WriteString("    if (options.limit) params.append('limit', options.limit.toString());\n")
	sb.WriteString("    if (options.offset) params.append('offset', options.offset.toString());\n")
	sb.WriteString("    const queryString = params.toString();\n")
	sb.WriteString("    if (queryString) url += `?${queryString}`;\n")
	sb.WriteString("  }\n")
	sb.WriteString("  const response = await fetch(url);\n")
	sb.WriteString("  if (!response.ok) {\n")
	sb.WriteString("    const error = await response.json().catch(() => ({ message: response.statusText }));\n")
	sb.WriteString("    throw new Error(error.message || `Failed to fetch list: ${response.statusText}`);\n")
	sb.WriteString("  }\n")
	sb.WriteString("  return response.json();\n")
	sb.WriteString("}\n")

	return sb.String()
}

// GenerateUpdateFunction generates an update by ID function
func (g *CRUDGenerator) GenerateUpdateFunction(table *codegen.Table) string {
	var sb strings.Builder

	typeName := g.typeMapper.ToTypeName(table.Name)
	functionName := fmt.Sprintf("update%s", typeName)

	// Determine primary key type
	pkType := g.getPrimaryKeyTSType(table)

	sb.WriteString(fmt.Sprintf("export async function %s(", functionName))
	sb.WriteString(fmt.Sprintf("baseUrl: string, id: %s, input: Update%sInput", pkType, typeName))
	sb.WriteString(fmt.Sprintf("): Promise<%s> {\n", typeName))
	sb.WriteString(fmt.Sprintf("  const response = await fetch(`${baseUrl}/%s/${id}`, {\n", table.Name))
	sb.WriteString("    method: 'PUT',\n")
	sb.WriteString("    headers: { 'Content-Type': 'application/json' },\n")
	sb.WriteString("    body: JSON.stringify(input),\n")
	sb.WriteString("  });\n")
	sb.WriteString("  if (!response.ok) {\n")
	sb.WriteString("    const error = await response.json().catch(() => ({ message: response.statusText }));\n")
	sb.WriteString("    throw new Error(error.message || `Failed to update: ${response.statusText}`);\n")
	sb.WriteString("  }\n")
	sb.WriteString("  return response.json();\n")
	sb.WriteString("}\n")

	return sb.String()
}

// GenerateDeleteFunction generates a delete by ID function
func (g *CRUDGenerator) GenerateDeleteFunction(table *codegen.Table) string {
	var sb strings.Builder

	typeName := g.typeMapper.ToTypeName(table.Name)
	functionName := fmt.Sprintf("delete%s", typeName)

	// Determine primary key type
	pkType := g.getPrimaryKeyTSType(table)

	sb.WriteString(fmt.Sprintf("export async function %s(", functionName))
	sb.WriteString(fmt.Sprintf("baseUrl: string, id: %s", pkType))
	sb.WriteString("): Promise<void> {\n")
	sb.WriteString(fmt.Sprintf("  const response = await fetch(`${baseUrl}/%s/${id}`, {\n", table.Name))
	sb.WriteString("    method: 'DELETE',\n")
	sb.WriteString("  });\n")
	sb.WriteString("  if (!response.ok) {\n")
	sb.WriteString("    const error = await response.json().catch(() => ({ message: response.statusText }));\n")
	sb.WriteString("    throw new Error(error.message || `Failed to delete: ${response.statusText}`);\n")
	sb.WriteString("  }\n")
	sb.WriteString("}\n")

	return sb.String()
}

// getPrimaryKeyTSType returns the TypeScript type for the table's primary key
// Returns "string" for UUID/TEXT types, "number" for integer types
func (g *CRUDGenerator) getPrimaryKeyTSType(table *codegen.Table) string {
	if table.PrimaryKey == nil || len(table.PrimaryKey.ColumnNames) == 0 {
		return "number" // default fallback
	}
	
	pkColName := table.PrimaryKey.ColumnNames[0]
	for _, col := range table.Columns {
		if col.Name == pkColName {
			return g.typeMapper.mapBaseType(col.DataType)
		}
	}
	return "number" // fallback if PK column not found
}

// GenerateAllCRUD generates all CRUD functions for a table
func (g *CRUDGenerator) GenerateAllCRUD(table *codegen.Table) string {
	var sb strings.Builder

	sb.WriteString(g.GenerateCreateFunction(table))
	sb.WriteString("\n")
	sb.WriteString(g.GenerateGetByIdFunction(table))
	sb.WriteString("\n")
	sb.WriteString(g.GenerateListFunction(table))
	sb.WriteString("\n")
	sb.WriteString(g.GenerateUpdateFunction(table))
	sb.WriteString("\n")
	sb.WriteString(g.GenerateDeleteFunction(table))

	return sb.String()
}

// pluralize simple pluralization (adding 's')
func (g *CRUDGenerator) pluralize(word string) string {
	// Simple pluralization - can be enhanced with library if needed
	if strings.HasSuffix(word, "s") || strings.HasSuffix(word, "x") || 
	   strings.HasSuffix(word, "z") || strings.HasSuffix(word, "ch") || 
	   strings.HasSuffix(word, "sh") {
		return word + "es"
	}
	if strings.HasSuffix(word, "y") && len(word) > 1 {
		// Check if preceded by consonant
		if !g.isVowel(rune(word[len(word)-2])) {
			return word[:len(word)-1] + "ies"
		}
	}
	return word + "s"
}

// isVowel checks if character is a vowel
func (g *CRUDGenerator) isVowel(r rune) bool {
	vowels := "aeiouAEIOU"
	return strings.ContainsRune(vowels, r)
}

package typescript_test

import (
	"testing"

	"github.com/kapok/kapok/pkg/codegen"
	"github.com/kapok/kapok/pkg/codegen/typescript"
	"github.com/stretchr/testify/assert"
)

func TestTypeMapper_MapType(t *testing.T) {
	tm := typescript.NewTypeMapper()

	tests := []struct {
		name       string
		pgType     string
		isNullable bool
		want       string
	}{
		{"varchar not null", "character varying", false, "string"},
		{"varchar nullable", "character varying", true, "string | null"},
		{"text", "text", false, "string"},
		{"integer", "integer", false, "number"},
		{"bigint", "bigint", false, "number"},
		{"serial", "serial", false, "number"},
		{"boolean", "boolean", false, "boolean"},
		{"timestamp", "timestamp without time zone", false, "Date"},
		{"timestamptz", "timestamp with time zone", false, "Date"},
		{"date", "date", false, "Date"},
		{"json", "json", false, "Record<string, any>"},
		{"jsonb", "jsonb", false, "Record<string, any>"},
		{"uuid", "uuid", false, "string"},
		{"numeric", "numeric", false, "number"},
		{"decimal", "decimal", false, "number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tm.MapType(tt.pgType, tt.isNullable)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTypeMapper_ToTypeName(t *testing.T) {
	tm := typescript.NewTypeMapper()

	tests := []struct {
		tableName string
		want      string
	}{
		{"users", "Users"},
		{"user_profiles", "UserProfiles"},
		{"oauth_tokens", "OauthTokens"},
		{"api_keys", "ApiKeys"},
	}

	for _, tt := range tests {
		t.Run(tt.tableName, func(t *testing.T) {
			got := tm.ToTypeName(tt.tableName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTypeMapper_ToFieldName(t *testing.T) {
	tm := typescript.NewTypeMapper()

	tests := []struct {
		columnName string
		want       string
	}{
		{"id", "id"},
		{"user_id", "userId"},
		{"created_at", "createdAt"},
		{"is_active", "isActive"},
		{"full_name", "fullName"},
	}

	for _, tt := range tests {
		t.Run(tt.columnName, func(t *testing.T) {
			got := tm.ToFieldName(tt.columnName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTypeMapper_GenerateInterface(t *testing.T) {
	tm := typescript.NewTypeMapper()

	table := &codegen.Table{
		Name:   "users",
		Schema: "public",
		Columns: []*codegen.Column{
			{
				Name:       "id",
				DataType:   "integer",
				IsNullable: false,
				Position:   1,
			},
			{
				Name:       "username",
				DataType:   "character varying",
				IsNullable: false,
				Position:   2,
			},
			{
				Name:       "email",
				DataType:   "character varying",
				IsNullable: true,
				Position:   3,
			},
			{
				Name:       "is_active",
				DataType:   "boolean",
				IsNullable: false,
				Position:   4,
			},
		},
	}

	got := tm.GenerateInterface(table)

	assert.Contains(t, got, "export interface Users {")
	assert.Contains(t, got, "id: number;")
	assert.Contains(t, got, "username: string;")
	assert.Contains(t, got, "email: string | null;")
	assert.Contains(t, got, "isActive: boolean;")
}

func TestTypeMapper_GenerateCreateInput(t *testing.T) {
	tm := typescript.NewTypeMapper()

	defaultTrue := "true"
	defaultNow := "CURRENT_TIMESTAMP"

	table := &codegen.Table{
		Name:   "users",
		Schema: "public",
		Columns: []*codegen.Column{
			{
				Name:         "id",
				DataType:     "serial",
				IsNullable:   false,
				DefaultValue: nil,
				Position:     1,
			},
			{
				Name:         "username",
				DataType:     "character varying",
				IsNullable:   false,
				DefaultValue: nil,
				Position:     2,
			},
			{
				Name:         "email",
				DataType:     "character varying",
				IsNullable:   true,
				DefaultValue: nil,
				Position:     3,
			},
			{
				Name:         "is_active",
				DataType:     "boolean",
				IsNullable:   false,
				DefaultValue: &defaultTrue,
				Position:     4,
			},
			{
				Name:         "created_at",
				DataType:     "timestamp without time zone",
				IsNullable:   false,
				DefaultValue: &defaultNow,
				Position:     5,
			},
		},
	}

	got := tm.GenerateCreateInput(table)

	assert.Contains(t, got, "export interface CreateUsersInput {")
	assert.NotContains(t, got, "id:") // Auto-generated, should be excluded
	assert.Contains(t, got, "username: string;")
	assert.Contains(t, got, "email?: string | null;") // Nullable = optional
	assert.Contains(t, got, "isActive?: boolean;")     // Has default = optional
	assert.NotContains(t, got, "createdAt")            // Auto-generated timestamp
}

func TestTypeMapper_GenerateUpdateInput(t *testing.T) {
	tm := typescript.NewTypeMapper()

	table := &codegen.Table{
		Name:   "users",
		Schema: "public",
		Columns: []*codegen.Column{
			{Name: "id", DataType: "integer", IsNullable: false, Position: 1},
			{Name: "username", DataType: "character varying", IsNullable: false, Position: 2},
			{Name: "email", DataType: "character varying", IsNullable: true, Position: 3},
			{Name: "is_active", DataType: "boolean", IsNullable: false, Position: 4},
			{Name: "created_at", DataType: "timestamp without time zone", IsNullable: false, Position: 5},
			{Name: "updated_at", DataType: "timestamp without time zone", IsNullable: false, Position: 6},
		},
		PrimaryKey: &codegen.PrimaryKey{
			ColumnNames: []string{"id"},
		},
	}

	got := tm.GenerateUpdateInput(table)

	assert.Contains(t, got, "export interface UpdateUsersInput {")
	assert.NotContains(t, got, "id?:") // PK excluded
	assert.Contains(t, got, "username?: string;")
	assert.Contains(t, got, "email?: string | null;")
	assert.Contains(t, got, "isActive?: boolean;")
	assert.NotContains(t, got, "createdAt") // Auto-timestamp excluded
	assert.NotContains(t, got, "updatedAt") // Auto-timestamp excluded
}

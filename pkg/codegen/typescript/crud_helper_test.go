package typescript

import (
	"testing"

	"github.com/kapok/kapok/pkg/codegen"
	"github.com/stretchr/testify/assert"
)

func TestGetPrimaryKeyTSType(t *testing.T) {
	gen := NewCRUDGenerator()

	tests := []struct {
		name     string
		table    *codegen.Table
		expected string
	}{
		{
			name: "integer primary key",
			table: &codegen.Table{
				Name: "users",
				Columns: []*codegen.Column{
					{Name: "id", DataType: "integer", IsNullable: false},
				},
				PrimaryKey: &codegen.PrimaryKey{
					ColumnNames: []string{"id"},
				},
			},
			expected: "number",
		},
		{
			name: "uuid primary key",
			table: &codegen.Table{
				Name: "posts",
				Columns: []*codegen.Column{
					{Name: "id", DataType: "uuid", IsNullable: false},
				},
				PrimaryKey: &codegen.PrimaryKey{
					ColumnNames: []string{"id"},
				},
			},
			expected: "string",
		},
		{
			name: "text primary key",
			table: &codegen.Table{
				Name: "codes",
				Columns: []*codegen.Column{
					{Name: "code", DataType: "text", IsNullable: false},
				},
				PrimaryKey: &codegen.PrimaryKey{
					ColumnNames: []string{"code"},
				},
			},
			expected: "string",
		},
		{
			name: "no primary key",
			table: &codegen.Table{
				Name:       "logs",
				Columns:    []*codegen.Column{},
				PrimaryKey: nil,
			},
			expected: "number",
		},
		{
			name: "empty primary key columns",
			table: &codegen.Table{
				Name:    "items",
				Columns: []*codegen.Column{},
				PrimaryKey: &codegen.PrimaryKey{
					ColumnNames: []string{},
				},
			},
			expected: "number",
		},
		{
			name: "bigint primary key",
			table: &codegen.Table{
				Name: "events",
				Columns: []*codegen.Column{
					{Name: "id", DataType: "bigint", IsNullable: false},
				},
				PrimaryKey: &codegen.PrimaryKey{
					ColumnNames: []string{"id"},
				},
			},
			expected: "number",
		},
		{
			name: "varchar primary key",
			table: &codegen.Table{
				Name: "tags",
				Columns: []*codegen.Column{
					{Name: "slug", DataType: "varchar", IsNullable: false},
				},
				PrimaryKey: &codegen.PrimaryKey{
					ColumnNames: []string{"slug"},
				},
			},
			expected: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.getPrimaryKeyTSType(tt.table)
			assert.Equal(t, tt.expected, result)
		})
	}
}

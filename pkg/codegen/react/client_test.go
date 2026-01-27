package react

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kapok/kapok/pkg/codegen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteSDK(t *testing.T) {
	tmpDir := t.TempDir()

	schema := &codegen.Schema{
		Tables: []*codegen.Table{
			{
				Name: "users",
				Columns: []*codegen.Column{
					{Name: "id", DataType: "integer"},
					{Name: "email", DataType: "varchar"},
				},
				PrimaryKey: &codegen.PrimaryKey{ColumnNames: []string{"id"}},
			},
		},
	}

	gen := NewReactClientGenerator()
	err := gen.WriteSDK(schema, tmpDir, "my-app-hooks", "my-app-sdk")
	require.NoError(t, err)

	// Verify directory structure
	assert.DirExists(t, filepath.Join(tmpDir, "src"))
	assert.DirExists(t, filepath.Join(tmpDir, "src", "hooks"))
	assert.DirExists(t, filepath.Join(tmpDir, "src", "types"))

	// Verify files exist
	assert.FileExists(t, filepath.Join(tmpDir, "package.json"))
	assert.FileExists(t, filepath.Join(tmpDir, "tsconfig.json"))
	assert.FileExists(t, filepath.Join(tmpDir, "README.md"))
	assert.FileExists(t, filepath.Join(tmpDir, "src", "provider.tsx"))
	assert.FileExists(t, filepath.Join(tmpDir, "src", "index.ts"))
	assert.FileExists(t, filepath.Join(tmpDir, "src", "hooks", "index.ts"))
	assert.FileExists(t, filepath.Join(tmpDir, "src", "hooks", "useUsers.ts"))
}

func TestGeneratePackageJSON(t *testing.T) {
	tmpDir := t.TempDir()

	gen := NewReactClientGenerator()
	err := gen.writePackageJSON(tmpDir, "my-app-hooks", "../typescript")
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "package.json"))
	require.NoError(t, err)

	// Verify content
	assert.Contains(t, string(content), `"name": "my-app-hooks"`)
	assert.Contains(t, string(content), `"react": "^18.0.0"`)
	assert.Contains(t, string(content), `"@tanstack/react-query": "^5.0.0"`)
	assert.Contains(t, string(content), `"kapok-sdk": "../typescript"`)
}

func TestGenerateTSConfig(t *testing.T) {
	tmpDir := t.TempDir()

	gen := NewReactClientGenerator()
	err := gen.writeTSConfig(tmpDir)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "tsconfig.json"))
	require.NoError(t, err)

	// Verify TS config
	assert.Contains(t, string(content), `"jsx": "react-jsx"`)
	assert.Contains(t, string(content), `"declaration": true`)
	assert.Contains(t, string(content), `"strict": true`)
}

func TestGenerateREADME(t *testing.T) {
	tmpDir := t.TempDir()

	gen := NewReactClientGenerator()
	err := gen.writeREADME(tmpDir, "my-app-hooks")
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	require.NoError(t, err)

	// Verify README content
	assert.Contains(t, string(content), "# my-app-hooks")
	assert.Contains(t, string(content), "QueryClientProvider")
	assert.Contains(t, string(content), "KapokProvider")
	assert.Contains(t, string(content), "useListUsers")
	assert.Contains(t, string(content), "useCreateUsers")
}

func TestGenerateMainIndex(t *testing.T) {
	schema := &codegen.Schema{
		Tables: []*codegen.Table{
			{Name: "users"},
			{Name: "posts"},
		},
	}

	gen := NewReactClientGenerator()
	result := gen.generateMainIndex(schema)

	// Verify exports
	assert.Contains(t, result, "export { KapokProvider, useKapokClient }")
	assert.Contains(t, result, "export * from './hooks'")
	assert.Contains(t, result, "export * from './types'")
}

func TestGenerateHooksIndex(t *testing.T) {
	schema := &codegen.Schema{
		Tables: []*codegen.Table{
			{Name: "users"},
			{Name: "blog_posts"},
		},
	}

	gen := NewReactClientGenerator()
	result := gen.generateHooksIndex(schema)

	// Verify hook exports
	assert.Contains(t, result, "export * from './useUsers'")
	assert.Contains(t, result, "export * from './useBlogPosts'")
}

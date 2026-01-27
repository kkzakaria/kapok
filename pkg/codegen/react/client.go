package react

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kapok/kapok/pkg/codegen"
)

// ReactClientGenerator generates complete React SDK with hooks
type ReactClientGenerator struct {
	hooksGen    *HooksGenerator
	providerGen *ProviderGenerator
}

// NewReactClientGenerator creates a new React client generator
func NewReactClientGenerator() *ReactClientGenerator {
	return &ReactClientGenerator{
		hooksGen:    NewHooksGenerator(),
		providerGen: NewProviderGenerator(),
	}
}

// WriteSDK generates and writes all React SDK files
func (g *ReactClientGenerator) WriteSDK(schema *codegen.Schema, outputDir, projectName, sdkImport string) error {
	// Create directory structure
	if err := g.createDirectoryStructure(outputDir); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Generate provider
	if err := g.writeProvider(outputDir, sdkImport); err != nil {
		return fmt.Errorf("failed to write provider: %w", err)
	}

	// Generate hooks for each table
	if err := g.writeHooks(schema, outputDir); err != nil {
		return fmt.Errorf("failed to write hooks: %w", err)
	}

	// Generate index exports
	if err := g.writeIndexFiles(schema, outputDir); err != nil {
		return fmt.Errorf("failed to write index files: %w", err)
	}

	// Generate package.json
	if err := g.writePackageJSON(outputDir, projectName, sdkImport); err != nil {
		return fmt.Errorf("failed to write package.json: %w", err)
	}

	// Generate tsconfig.json
	if err := g.writeTSConfig(outputDir); err != nil {
		return fmt.Errorf("failed to write tsconfig.json: %w", err)
	}

	// Generate README
	if err := g.writeREADME(outputDir, projectName); err != nil {
		return fmt.Errorf("failed to write README: %w", err)
	}

	return nil
}

func (g *ReactClientGenerator) createDirectoryStructure(outputDir string) error {
	dirs := []string{
		filepath.Join(outputDir, "src"),
		filepath.Join(outputDir, "src", "hooks"),
		filepath.Join(outputDir, "src", "types"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

func (g *ReactClientGenerator) writeProvider(outputDir, sdkImport string) error {
	content := g.providerGen.GenerateProvider(sdkImport)
	path := filepath.Join(outputDir, "src", "provider.tsx")
	return os.WriteFile(path, []byte(content), 0644)
}

func (g *ReactClientGenerator) writeHooks(schema *codegen.Schema, outputDir string) error {
	hooksDir := filepath.Join(outputDir, "src", "hooks")

	for _, table := range schema.Tables {
		content := g.hooksGen.GenerateAllHooks(table)
		filename := fmt.Sprintf("use%s.ts", g.hooksGen.typeMapper.ToTypeName(table.Name))
		path := filepath.Join(hooksDir, filename)

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

func (g *ReactClientGenerator) writeIndexFiles(schema *codegen.Schema, outputDir string) error {
	// Main index
	mainIndex := g.generateMainIndex(schema)
	if err := os.WriteFile(filepath.Join(outputDir, "src", "index.ts"), []byte(mainIndex), 0644); err != nil {
		return err
	}

	// Hooks index
	hooksIndex := g.generateHooksIndex(schema)
	if err := os.WriteFile(filepath.Join(outputDir, "src", "hooks", "index.ts"), []byte(hooksIndex), 0644); err != nil {
		return err
	}

	// Types index (re-export from SDK)
	typesIndex := "export * from '../../types';\n"
	return os.WriteFile(filepath.Join(outputDir, "src", "types", "index.ts"), []byte(typesIndex), 0644)
}

func (g *ReactClientGenerator) generateMainIndex(schema *codegen.Schema) string {
	var sb strings.Builder

	sb.WriteString("// Export provider\n")
	sb.WriteString("export { KapokProvider, useKapokClient } from './provider';\n\n")

	sb.WriteString("// Export all hooks\n")
	sb.WriteString("export * from './hooks';\n\n")

	sb.WriteString("// Export types\n")
	sb.WriteString("export * from './types';\n")

	return sb.String()
}

func (g *ReactClientGenerator) generateHooksIndex(schema *codegen.Schema) string {
	var sb strings.Builder

	for _, table := range schema.Tables {
		typeName := g.hooksGen.typeMapper.ToTypeName(table.Name)
		sb.WriteString(fmt.Sprintf("export * from './use%s';\n", typeName))
	}

	return sb.String()
}

func (g *ReactClientGenerator) writePackageJSON(outputDir, projectName, sdkImport string) error {
	content := fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "Auto-generated React hooks for Kapok backend",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "scripts": {
    "build": "tsc",
    "watch": "tsc --watch"
  },
  "peerDependencies": {
    "react": "^18.0.0",
    "@tanstack/react-query": "^5.0.0"
  },
  "dependencies": {
    "kapok-sdk": "%s"
  },
  "devDependencies": {
    "@types/react": "^18.0.0",
    "typescript": "^5.0.0"
  }
}
`, projectName, sdkImport)

	path := filepath.Join(outputDir, "package.json")
	return os.WriteFile(path, []byte(content), 0644)
}

func (g *ReactClientGenerator) writeTSConfig(outputDir string) error {
	content := `{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "jsx": "react-jsx",
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "moduleResolution": "node",
    "resolveJsonModule": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}
`

	path := filepath.Join(outputDir, "tsconfig.json")
	return os.WriteFile(path, []byte(content), 0644)
}

func (g *ReactClientGenerator) writeREADME(outputDir, projectName string) error {
	content := "# " + projectName + `

Auto-generated React hooks for your Kapok backend using React Query.

## Installation

` + "```bash" + `
npm install
npm run build
` + "```" + `

## Setup

Wrap your app with the required providers:

` + "```tsx" + `
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { KapokProvider } from '` + projectName + `';

const queryClient = new QueryClient();

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <KapokProvider baseUrl="http://localhost:8080/api">
        <YourApp />
      </KapokProvider>
    </QueryClientProvider>
  );
}
` + "```" + `

## Usage

### Query Hooks (Fetch Data)

` + "```tsx" + `
import { useListUsers, useUsersById } from '` + projectName + `';

function UsersList() {
  const { data: users, isLoading, error } = useListUsers({ limit: 10 });
  
  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;
  
  return (
    <ul>
      {users.map(user => (
        <li key={user.id}>{user.name}</li>
      ))}
    </ul>
  );
}

function UserDetail({ userId }: { userId: number }) {
  const { data: user } = useUsersById(userId);
  return <div>{user?.name}</div>;
}
` + "```" + `

### Mutation Hooks (Modify Data)

` + "```tsx" + `
import { useCreateUsers, useUpdateUsers, useDeleteUsers } from '` + projectName + `';

function CreateUserForm() {
  const createUser = useCreateUsers();
  
  const handleSubmit = async (e) => {
    e.preventDefault();
    await createUser.mutateAsync({
      email: 'user@example.com',
      name: 'John Doe',
    });
  };
  
  return (
    <form onSubmit={handleSubmit}>
      <button type="submit" disabled={createUser.isPending}>
        {createUser.isPending ? 'Creating...' : 'Create User'}
      </button>
      {createUser.isError && <p>Error: {createUser.error.message}</p>}
    </form>
  );
}

function UpdateUserButton({ userId }: { userId: number }) {
  const updateUser = useUpdateUsers();
  
  return (
    <button
      onClick={() => updateUser.mutate({ 
        id: userId, 
        input: { name: 'Updated Name' } 
      })}
    >
      Update
    </button>
  );
}

function DeleteUserButton({ userId }: { userId: number }) {
  const deleteUser = useDeleteUsers();
  
  return (
    <button onClick={() => deleteUser.mutate(userId)}>
      Delete
    </button>
  );
}
` + "```" + `

## Features

- ✅ Type-safe hooks with TypeScript
- ✅ Automatic caching and background refetching
- ✅ Loading and error states handled
- ✅ Automatic cache invalidation on mutations
- ✅ Pagination support for list queries

## Development

` + "```bash" + `
npm run watch  # Watch for changes
npm run build  # Build TypeScript
` + "```" + `

---

Generated by [Kapok](https://github.com/kapok/kapok)
`

	path := filepath.Join(outputDir, "README.md")
	return os.WriteFile(path, []byte(content), 0644)
}


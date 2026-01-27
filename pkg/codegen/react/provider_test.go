package react

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateProvider(t *testing.T) {
	gen := NewProviderGenerator()
	projectName := "my-app-sdk"

	result := gen.GenerateProvider(projectName)

	// Verify imports
	assert.Contains(t, result, "import React, { createContext, useContext, useMemo, ReactNode }")
	assert.Contains(t, result, fmt.Sprintf("import { KapokClient } from '%s'", projectName))

	// Verify context creation
	assert.Contains(t, result, "const KapokContext = createContext<KapokClient | null>(null)")

	// Verify provider props interface
	assert.Contains(t, result, "export interface KapokProviderProps")
	assert.Contains(t, result, "children: ReactNode")
	assert.Contains(t, result, "baseUrl: string")

	// Verify provider component
	assert.Contains(t, result, "export function KapokProvider({ children, baseUrl }: KapokProviderProps)")
	assert.Contains(t, result, "const client = useMemo(() => new KapokClient(baseUrl), [baseUrl])")
	assert.Contains(t, result, "<KapokContext.Provider value={client}>")

	// Verify useKapokClient hook
	assert.Contains(t, result, "export function useKapokClient(): KapokClient")
	assert.Contains(t, result, "const client = useContext(KapokContext)")
	assert.Contains(t, result, "if (!client)")
	assert.Contains(t, result, "useKapokClient must be used within a KapokProvider")
}

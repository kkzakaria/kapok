package react

import (
	"fmt"
	"strings"
)

// ProviderGenerator generates React Context provider for SDK client
type ProviderGenerator struct{}

// NewProviderGenerator creates a new provider generator
func NewProviderGenerator() *ProviderGenerator {
	return &ProviderGenerator{}
}

// GenerateProvider generates the KapokProvider component
func (g *ProviderGenerator) GenerateProvider(projectName string) string {
	var sb strings.Builder

	sb.WriteString("import React, { createContext, useContext, useMemo, ReactNode } from 'react';\n")
	sb.WriteString(fmt.Sprintf("import { KapokClient } from '%s';\n\n", projectName))

	// Context definition
	sb.WriteString("const KapokContext = createContext<KapokClient | null>(null);\n\n")

	// Provider props interface
	sb.WriteString("export interface KapokProviderProps {\n")
	sb.WriteString("  children: ReactNode;\n")
	sb.WriteString("  baseUrl: string;\n")
	sb.WriteString("}\n\n")

	// Provider component
	sb.WriteString("export function KapokProvider({ children, baseUrl }: KapokProviderProps) {\n")
	sb.WriteString("  const client = useMemo(() => new KapokClient(baseUrl), [baseUrl]);\n")
	sb.WriteString("  \n")
	sb.WriteString("  return (\n")
	sb.WriteString("    <KapokContext.Provider value={client}>\n")
	sb.WriteString("      {children}\n")
	sb.WriteString("    </KapokContext.Provider>\n")
	sb.WriteString("  );\n")
	sb.WriteString("}\n\n")

	// Hook to access client
	sb.WriteString("export function useKapokClient(): KapokClient {\n")
	sb.WriteString("  const client = useContext(KapokContext);\n")
	sb.WriteString("  \n")
	sb.WriteString("  if (!client) {\n")
	sb.WriteString("    throw new Error(\n")
	sb.WriteString("      'useKapokClient must be used within a KapokProvider. ' +\n")
	sb.WriteString("      'Wrap your component tree with <KapokProvider baseUrl=\"...\">'\n")
	sb.WriteString("    );\n")
	sb.WriteString("  }\n")
	sb.WriteString("  \n")
	sb.WriteString("  return client;\n")
	sb.WriteString("}\n")

	return sb.String()
}

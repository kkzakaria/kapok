package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new Kapok project",
	Long: `Initialize a new Kapok project with zero configuration.

Creates:
  • kapok.yaml configuration file with smart defaults
  • .env.example template
  • README.md with project-specific quick start
  • docs/ folder with basic architecture documentation`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := "my-kapok-project"
		if len(args) > 0 {
			projectName = args[0]
		}

		force, _ := cmd.Flags().GetBool("force")

		fmt.Fprintf(cmd.OutOrStdout(), "🚀 Initializing Kapok project: %s\n\n", projectName)

		// Check if directory is empty (unless --force)
		if !force {
			if err := checkDirectoryEmpty("."); err != nil {
				return fmt.Errorf("directory is not empty (use --force to override): %w", err)
			}
		}

		// Create kapok.yaml
		if err := createKapokConfig(projectName, force); err != nil {
			return fmt.Errorf("failed to create kapok.yaml: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "✓ Created kapok.yaml")

		// Create .env.example  
		if err := createEnvExample(projectName, force); err != nil {
			return fmt.Errorf("failed to create .env.example: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "✓ Created .env.example")

		// Create README.md
		if err := createReadme(projectName, force); err != nil {
			return fmt.Errorf("failed to create README.md: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "✓ Created README.md")

		// Create docs/ folder
		if err := createDocsFolder(projectName, force); err != nil {
			return fmt.Errorf("failed to create docs folder: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "✓ Created docs/ folder")

		// Success message
		fmt.Fprintln(cmd.OutOrStdout(), "\n✨ Project initialized successfully!")
		fmt.Fprintln(cmd.OutOrStdout(), "\n📝 Next steps:")
		fmt.Fprintln(cmd.OutOrStdout(), "  1. Copy .env.example to .env and set your secrets")
		fmt.Fprintln(cmd.OutOrStdout(), "  2. Review and customize kapok.yaml")
		fmt.Fprintln(cmd.OutOrStdout(), "  3. Run: kapok dev")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().Bool("force", false, "Overwrite existing files")
}

// checkDirectoryEmpty returns error if directory is not empty
func checkDirectoryEmpty(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Ignore hidden files (like .git)
	for _, entry := range entries {
		if !isHiddenFile(entry.Name()) {
			return fmt.Errorf("directory contains files")
		}
	}

	return nil
}

// isHiddenFile checks if filename starts with dot
func isHiddenFile(name string) bool {
	return len(name) > 0 && name[0] == '.'
}

// createKapokConfig creates kapok.yaml configuration file
func createKapokConfig(projectName string, force bool) error {
	filename := "kapok.yaml"

	if !force && fileExists(filename) {
		return fmt.Errorf("file already exists")
	}

	content := fmt.Sprintf(`# Kapok Configuration for %s
# See https://github.com/kapok/kapok for full documentation

server:
  host: "0.0.0.0"
  port: 8080

database:
  host: "localhost"
  port: 5432
  user: "kapok"
  # password: set via KAPOK_DATABASE_PASSWORD environment variable
  database: "%s"
  ssl_mode: "disable"
  pool_size: 20

redis:
  host: "localhost"
  port: 6379
  # password: set via KAPOK_REDIS_PASSWORD if needed
  db: 0

log:
  level: "info"    # debug, info, warn, error
  format: "console" # json (production), console (development)

jwt:
  # secret: set via KAPOK_JWT_SECRET (minimum 32 characters)
  access_token_ttl: "15m"
  refresh_token_ttl: "168h"  # 7 days
  signing_algorithm: "HS256"
`, projectName, projectName)

	return os.WriteFile(filename, []byte(content), 0644)
}

// createEnvExample creates .env.example template
func createEnvExample(projectName string, force bool) error {
	filename := ".env.example"

	if !force && fileExists(filename) {
		return fmt.Errorf("file already exists")
	}

	content := `# Kapok Environment Variables Template
# Copy this file to .env and set your actual values
# NEVER commit .env to version control!

# Database Configuration
KAPOK_DATABASE_HOST=localhost
KAPOK_DATABASE_PORT=5432
KAPOK_DATABASE_USER=kapok
KAPOK_DATABASE_PASSWORD=your-secure-database-password-here
KAPOK_DATABASE_DATABASE=kapok

# Redis Configuration (optional)
# KAPOK_REDIS_HOST=localhost
# KAPOK_REDIS_PORT=6379
# KAPOK_REDIS_PASSWORD=your-redis-password-if-needed

# JWT Configuration
KAPOK_JWT_SECRET=your-jwt-secret-minimum-32-characters-long-please

# Server Configuration (optional overrides)
# KAPOK_SERVER_HOST=0.0.0.0
# KAPOK_SERVER_PORT=8080

# Logging (optional)
# KAPOK_LOG_LEVEL=debug
# KAPOK_LOG_FORMAT=console
`

	return os.WriteFile(filename, []byte(content), 0644)
}

// createReadme creates project-specific README.md
func createReadme(projectName string, force bool) error {
	filename := "README.md"

	if !force && fileExists(filename) {
		return fmt.Errorf("file already exists")
	}

	content := fmt.Sprintf(`# %s

Backend powered by [Kapok](https://github.com/kapok/kapok) - Backend-as-a-Service with multi-tenancy.

## Quick Start

### Prerequisites

- Go 1.21+ installed
- PostgreSQL 14+ running
- Redis (optional, for caching)

### Setup

1. **Install dependencies**
   ` + "`" + `bash
   cp .env.example .env
   # Edit .env and set your database password and JWT secret
   ` + "`" + `

2. **Start local development**
   ` + "`" + `bash
   kapok dev
   ` + "`" + `

3. **Access GraphQL Playground**
   ` + "`" + `
   http://localhost:8080/playground
   ` + "`" + `

### Configuration

See ` + "`" + `kapok.yaml` + "`" + ` for all configuration options.

Environment variables take precedence over config file:
- Database password: ` + "`" + `KAPOK_DATABASE_PASSWORD` + "`" + `
- JWT secret: ` + "`" + `KAPOK_JWT_SECRET` + "`" + `

See ` + "`" + `.env.example` + "`" + ` for all available environment variables.

### Project Structure

` + "`" + `
%s/
├── kapok.yaml        # Kapok configuration
├── .env              # Environment variables (not committed)
├── docs/             # Project documentation
└── README.md         # This file
` + "`" + `

### Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Kapok Documentation](https://github.com/kapok/kapok)

### Deployment

` + "`" + `bash
# Deploy to Kubernetes
kapok deploy

# See deployment options
kapok deploy --help
` + "`" + `

## License

_To be determined_
`, projectName, projectName)

	return os.WriteFile(filename, []byte(content), 0644)
}

// createDocsFolder creates docs/ folder with basic documentation
func createDocsFolder(projectName string, force bool) error {
	docsDir := "docs"

	// Create docs directory
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return err
	}

	// Create ARCHITECTURE.md
	archFile := filepath.Join(docsDir, "ARCHITECTURE.md")
	if !force && fileExists(archFile) {
		return fmt.Errorf("docs/ARCHITECTURE.md already exists")
	}

	archContent := fmt.Sprintf(`# %s Architecture

## Overview

This project uses Kapok as its backend infrastructure, providing:
- Multi-tenant PostgreSQL database with schema-per-tenant isolation
- Auto-generated GraphQL API from database schema
- Built-in authentication with JWT
- Redis caching layer

## Components

### Database Layer
- **PostgreSQL**: Primary data store with tenant isolation
- **Schema Design**: Each tenant gets dedicated schema
- **Migrations**: Managed via Kapok

### API Layer
- **GraphQL**: Auto-generated from PostgreSQL schema
- **Authentication**: JWT-based with tenant context
- **Authorization**: Role-based access control (RBAC)

### Infrastructure
- **Kubernetes**: Container orchestration
- **Redis**: Caching and session storage
- **Monitoring**: Built-in observability

## Development

See main [README.md](../README.md) for setup instructions.

## Deployment

Kapok handles deployment to Kubernetes clusters:
- AWS EKS
- GCP GKE
- Azure AKS

## Security

- Secrets managed via environment variables
- Database passwords never in config files
- JWT secrets rotated regularly
- HTTPS/TLS enforced in production
`, projectName)

	return os.WriteFile(archFile, []byte(archContent), 0644)
}

// fileExists checks if file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

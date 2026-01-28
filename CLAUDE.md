# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Kapok is a self-hosted Backend-as-a-Service (BaaS) platform with native multi-tenancy and automatic SDK generation. It provides frontend developers with full infrastructure control without requiring DevOps expertise.

## Build & Development Commands

```bash
# Build
make build                    # Build all binaries (gateway + CLI)
make build-cli                # Build CLI only → bin/kapok-cli
make build-gateway            # Build gateway only → bin/kapok-gateway

# Run
make run                      # Run gateway locally
make dev                      # Start dev environment (Docker: PostgreSQL, Redis, MinIO, Mailhog)
make stop                     # Stop dev environment

# Test
go test ./...                 # Run all tests
go test -v -race ./...        # Run tests with race detection
go test ./internal/tenant/... # Run tests for specific package
make test-coverage            # Generate coverage report → coverage.html

# Lint & Format
make lint                     # Run golangci-lint
make fmt                      # Format code (gofmt + gofumpt)
make vet                      # Run go vet

# Database
make db-shell                 # Open PostgreSQL shell
make db-reset                 # Drop and recreate database
```

## Architecture

### Multi-Tenant Design
- **Schema-per-tenant isolation**: Each tenant gets a dedicated PostgreSQL schema (`tenant_{uuid}`)
- Tenant context is extracted from JWT tokens and propagated through request context
- Row-Level Security (RLS) provides additional data isolation

### Key Internal Packages (`internal/`)
- **tenant/**: Tenant context management, routing, and schema provisioning. `GetTenant(ctx)` retrieves tenant from context.
- **graphql/**: Dynamic GraphQL schema generation from PostgreSQL introspection. Schemas are cached with 5-minute TTL.
- **auth/**: JWT token management (access + refresh tokens), middleware for authentication
- **rbac/**: Role-based access control using Casbin
- **database/**: Connection pooling, migrations, RLS policy management

### Public Packages (`pkg/`)
- **codegen/**: SDK generation from database schema introspection
  - `typescript/`: TypeScript SDK generator (types, CRUD, client)
  - `react/`: React Query hooks generator
- **config/**: Configuration loading with Viper (YAML files + env vars)
- **logger/**: Zerolog-based structured logging

### CLI Structure (`cmd/kapok/`)
Commands: `init`, `dev`, `generate sdk`, `generate react`, `deploy`, `tenant` (create/list/delete)

## Code Conventions

### Files & Naming
- Files: `snake_case.go`
- Packages: lowercase, singular, no underscores
- Exported: `PascalCase`, Private: `camelCase`

### Logging
Always use zerolog structured logging, never `fmt.Println()` or `log.Print()`:
```go
log.Info().Str("tenant_id", tenantID).Msg("tenant created")
```

### Error Handling
Use `%w` for error wrapping:
```go
return fmt.Errorf("failed to create tenant: %w", err)
```

### Testing
- Co-locate tests: `filename_test.go`
- Use table-driven tests with `testify`
- Integration tests use `dockertest` for PostgreSQL containers

## Git Workflow

Branch protection enforces PRs to `main`. Use conventional commits:
```
feat(scope): description (Story X.Y)
fix(cli): resolve config loading issue
```

Branch naming: `feature/epic-X-story-Y-description`

## Configuration

Precedence: CLI flags → Environment variables (`KAPOK_*`) → Config files → Defaults

Required environment variables for production:
- `KAPOK_DATABASE_PASSWORD`
- `KAPOK_JWT_SECRET` (min 32 chars)

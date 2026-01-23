---
storyId: "1.1"
epicId: "1"
epicTitle: "Project Foundation & Local Development"
storyTitle: "Initialize Monorepo Structure"
status: "dev-complete"
assignedTo: "Antigravity"
estimatedComplexity: "small"
completedAt: "2026-01-23"
dependencies: []
---

# Story 1.1: Initialize Monorepo Structure

As a **platform developer**,\
I want to **set up the Go monorepo structure with proper organization**,\
So that **code is organized following best practices and architectural
guidelines**.

## Acceptance Criteria

**AC1: Directory Structure Created**

- **Given** the project repository is empty
- **When** the monorepo structure is created
- **Then** directories `cmd/`, `internal/`, `pkg/`, `deployments/`, `testdata/`,
  `scripts/` exist
- **And** `go.mod` is initialized with Go 1.21+ and module name
  `github.com/kapok/kapok`
- **And** `.gitignore` excludes binaries, `.env`, and IDE files
- **And** `README.md` contains project overview and quick start placeholder

## Context & References

**From Architecture Document:**

- AR01: Monorepo Structure - Single Go monorepo avec `cmd/`, `internal/`,
  `pkg/`, `deployments/` organization
- AR11: Naming Conventions - snake_case (files), camelCase (Go/JSON), PascalCase
  (exported Go)
- AR20: Deployment Target - Go 1.21+

**From PRD:**

- FR05: CLI Developer-Friendly - Foundation for Cobra CLI
- NFR48: Onboarding Time - <5 minutes from install to deployed backend

## Implementation Notes

### Directory Structure:

```
kapok/
├── cmd/                        # Entrypoints (binaries)
│   ├── kapok/                  # CLI binary
│   ├── control-plane/          # Control plane service
│   ├── graphql-engine/         # GraphQL engine service
│   └── provisioner/            # Database provisioner service
├── internal/                   # Private application code
│   ├── auth/                   # Authentication & JWT
│   ├── tenant/                 # Tenant context & routing
│   ├── database/               # Database connections & pooling
│   ├── graphql/                # GraphQL resolver logic
│   ├── rbac/                   # RBAC with Casbin
│   └── k8s/                    # Kubernetes client wrappers
├── pkg/                        # Exported libraries (RARE)
│   ├── api/                    # Shared API types
│   └── config/                 # Configuration structs
├── deployments/                # Deployment configs
│   └── helm/                   # Helm charts
├── testdata/                   # Test fixtures
├── scripts/                    # Build & deployment scripts
├── go.mod                      # Go module file
├── go.sum                      # Go module checksums
├── Makefile                    # Build automation
├── .gitignore                  # Git ignore patterns
└── README.md                   # Project documentation
```

### go.mod Requirements:

- Module name: `github.com/kapok/kapok`
- Go version: 1.21 or higher
- No dependencies yet (will be added in subsequent stories)

### .gitignore Must Include:

- Binary outputs: `cmd/*/kapok`, `/bin/`, `/dist/`
- Environment files: `.env`, `.env.local`
- IDE files: `.vscode/`, `.idea/`, `*.swp`
- OS files: `.DS_Store`, `Thumbs.db`
- Test coverage: `*.out`, `coverage.html`

### README.md Structure:

```markdown
# Kapok

Backend-as-a-Service auto-hébergé avec multi-tenancy native.

## Quick Start

_Coming soon - see docs/ for detailed setup_

## Architecture

See `_bmad-output/planning-artifacts/architecture.md`

## Contributing

_Coming soon_

## License

_To be determined_
```

## Tasks/Subtasks

- [ ] Create root directory structure
- [ ] Initialize go.mod with correct module name and Go version
- [ ] Create .gitignore with comprehensive patterns
- [ ] Create basic README.md
- [ ] Create placeholder directories with .gitkeep files
- [ ] Verify structure matches architectural guidelines

## Test Strategy

- Verify all directories exist
- Validate go.mod format and content
- Check .gitignore patterns are complete
- Ensure README.md follows template

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Directory structure matches architecture document
- [ ] go.mod is valid and uses Go 1.21+
- [ ] .gitignore covers all necessary patterns
- [ ] README.md contains required sections
- [ ] All files committed to git
- [ ] Story marked as `dev-complete` in frontmatter

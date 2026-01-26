# Kapok

Backend-as-a-Service auto-hébergé avec multi-tenancy native.

## Overview

Kapok est une plateforme Backend-as-a-Service (BaaS) auto-hébergée conçue pour
les développeurs frontend qui ont besoin de contrôle total sur leur
infrastructure sans expertise DevOps.

### Features Clés (MVP)

- **Multi-Tenant Foundation**: Schema-per-tenant isolation dans PostgreSQL
- **GraphQL Auto-Generated**: API GraphQL générée automatiquement depuis votre
  schéma PostgreSQL
- **CLI Developer-Friendly**: `kapok init`, `kapok dev`, `kapok deploy`,
  `kapok tenant`
- **Kubernetes Deployment**: Déploiement one-command sur AWS EKS, GCP GKE, Azure
  AKS
- **Zero-Config**: Configuration minimale, smart defaults pour 90% des use cases

## Quick Start

_Documentation complète à venir - voir `_bmad-output/planning-artifacts/` pour
PRD et Architecture_

### Installation

```bash
# Coming soon
go install github.com/kapok/kapok/cmd/kapok@latest
```

### Usage

```bash
# Initialize new project
kapok init my-backend

# Start local development
kapok dev

#Deploy to Kubernetes
kapok deploy
```

## Configuration

Kapok uses a flexible configuration system with multiple sources.

### Configuration Precedence (Highest to Lowest)

1. **CLI flags** - Command-line arguments (e.g., `--port=9090`)
2. **Environment variables** - Prefixed with `KAPOK_` (e.g.,
   `KAPOK_SERVER_PORT=9090`)
3. **Config files** - Searched in order:
   - `./kapok.yaml` (current directory)
   - `~/.kapok/config.yaml` (user home)
   - `/etc/kapok/config.yaml` (system-wide)
4. **Defaults** - Smart defaults for development

### Quick Start Configuration

```bash
# 1. Copy example config
cp kapok.yaml.example kapok.yaml

# 2. Set required secrets via environment variables
export KAPOK_DATABASE_PASSWORD="your-secure-password"
export KAPOK_JWT_SECRET="your-jwt-secret-min-32-characters"

# 3. (Optional) Customize kapok.yaml for your needs
nano kapok.yaml

# 4. Run Kapok
kapok dev
```

### Environment Variables

All configuration can be set via environment variables using the `KAPOK_`
prefix:

```bash
# Server configuration
export KAPOK_SERVER_HOST="0.0.0.0"
export KAPOK_SERVER_PORT="8080"

# Database configuration
export KAPOK_DATABASE_HOST="localhost"
export KAPOK_DATABASE_PORT="5432"
export KAPOK_DATABASE_USER="kapok"
export KAPOK_DATABASE_PASSWORD="secure-password"  # Required!
export KAPOK_DATABASE_DATABASE="kapok"

# Redis configuration
export KAPOK_REDIS_HOST="localhost"
export KAPOK_REDIS_PORT="6379"
export KAPOK_REDIS_PASSWORD="redis-password"      # If required

# JWT configuration
export KAPOK_JWT_SECRET="min-32-chars-secret"     # Required!
export KAPOK_JWT_ACCESS_TOKEN_TTL="15m"
export KAPOK_JWT_REFRESH_TOKEN_TTL="168h"

# Logging
export KAPOK_LOG_LEVEL="info"  # debug, info, warn, error
export KAPOK_LOG_FORMAT="json" # json, console
```

**Note:** Nested config keys use underscores: `database.host` →
`KAPOK_DATABASE_HOST`

### Security

**⚠️ CRITICAL: Never commit secrets to config files!**

Sensitive values (passwords, tokens, secrets) **MUST** be set via environment
variables:

- `KAPOK_DATABASE_PASSWORD` - Database password
- `KAPOK_JWT_SECRET` - JWT signing secret (minimum 32 characters)
- `KAPOK_REDIS_PASSWORD` - Redis password (if authentication enabled)

### Example Config File

See `kapok.yaml.example` for a fully documented configuration template.

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  host: "localhost"
  port: 5432
  user: "kapok"
  # password: set via KAPOK_DATABASE_PASSWORD

log:
  level: "info"
  format: "json"
```

## Architecture

Voir la documentation complète d'architecture :

- [Architecture Decision Document](./_bmad-output/planning-artifacts/architecture.md)
- [Product Requirements Document](./_bmad-output/planning-artifacts/prd.md)
- [Epics & Stories](./_bmad-output/planning-artifacts/epics.md)

## Project Structure

```
kapok/
├── cmd/                    # Binary entry points
│   ├── kapok/              # CLI binary
│   ├── control-plane/      # Control plane service
│   ├── graphql-engine/     # GraphQL engine service
│   └── provisioner/        # Database provisioner service
├── internal/               # Private application code
│   ├── auth/               # Authentication & JWT
│   ├── tenant/             # Tenant context & routing
│   ├── database/           # Database connections
│   ├── graphql/            # GraphQL resolver logic
│   ├── rbac/               # RBAC with Casbin
│   └── k8s/                # Kubernetes client wrappers
├── pkg/                    # Exported libraries
│   ├── api/                # Shared API types
│   └── config/             # Configuration structs
├── deployments/            # Deployment configurations
│   └── helm/               # Helm charts
├── testdata/               # Test fixtures
└── scripts/                # Build and deployment scripts
```

## Development

### Prerequisites

- Go 1.21+
- Docker (for local development)
- Kubernetes cluster (for deployment)

### Building

```bash
# Coming soon
make build
```

### Testing

```bash
# Coming soon
make test
```

## Contributing

Contributions are welcome! Please read our contributing guidelines (coming
soon).

## License

_To be determined_

## Contact

- GitHub: https://github.com/kapok/kapok
- Documentation: _Coming soon_

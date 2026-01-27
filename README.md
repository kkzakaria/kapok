# Kapok

Backend-as-a-Service auto-hébergé avec multi-tenancy native et génération
automatique de SDKs.

## Overview

Kapok est une plateforme Backend-as-a-Service (BaaS) auto-hébergée conçue pour
les développeurs frontend qui ont besoin de contrôle total sur leur
infrastructure sans expertise DevOps.

### Features Clés

- **Multi-Tenant Foundation**: Schema-per-tenant isolation dans PostgreSQL
- **Auto-Generated SDKs**: SDKs TypeScript et React hooks générés
  automatiquement
- **Type-Safe API Client**: Client TypeScript avec autocomplétion complète
- **React Query Integration**: Hooks React avec caching intégré
- **CLI Developer-Friendly**: `kapok init`, `kapok dev`, `kapok generate`
- **Kubernetes Deployment**: Déploiement one-command sur AWS EKS, GCP GKE, Azure
  AKS
- **Zero-Config**: Configuration minimale, smart defaults pour 90% des use cases

## 🚀 Quick Start

**Get a working backend with auto-generated SDKs in under 5 minutes!**

### Installation

```bash
go install github.com/kapok/kapok/cmd/kapok@latest
```

### Create Your First Project

```bash
# Initialize project
kapok init my-blog
cd my-blog

# Start development server
kapok dev
```

### Generate SDKs

```bash
# Generate TypeScript SDK
kapok generate sdk --schema public

# Generate React hooks
kapok generate react
```

### Use in Your App

```typescript
import { KapokProvider, useListPosts } from "kapok-react";

function BlogPosts() {
  const { data: posts, isLoading } = useListPosts();

  if (isLoading) return <div>Loading...</div>;

  return (
    <div>
      {posts.map((post) => (
        <article key={post.id}>
          <h2>{post.title}</h2>
        </article>
      ))}
    </div>
  );
}
```

**📖 Complete Guide**: See [docs/quickstart.md](./docs/quickstart.md) for the
full tutorial.

**📦 Examples**: Check [examples/quickstart/](./examples/quickstart/) for
working code samples.

---

## SDK Generation

Kapok automatically generates type-safe SDKs from your PostgreSQL schema.

### TypeScript SDK

```bash
kapok generate sdk --schema public --project-name my-app-sdk
```

**Generates**:

- TypeScript interfaces for all tables
- CRUD functions (`create`, `list`, `getById`, `update`, `delete`)
- Type-safe `KapokClient` class
- Full autocomplete support

**Usage**:

```typescript
import { KapokClient } from "my-app-sdk";

const client = new KapokClient("http://localhost:8080/api");
const posts = await client.posts.list({ limit: 10 });
```

### React Hooks

```bash
kapok generate react --sdk-import ../typescript
```

**Generates**:

- Query hooks: `useListXxx`, `useXxxById`
- Mutation hooks: `useCreateXxx`, `useUpdateXxx`, `useDeleteXxx`
- React Query integration with automatic caching
- `KapokProvider` context component

**Usage**:

```typescript
import { useCreatePosts, useListPosts } from "my-app-react";

function MyComponent() {
  const { data, isLoading } = useListPosts();
  const createPost = useCreatePosts();

  // Your component logic
}
```

---

## CLI Commands

```bash
kapok init <project>    # Initialize new project
kapok dev               # Start development server
kapok generate sdk      # Generate TypeScript SDK
kapok generate react    # Generate React hooks
kapok deploy            # Deploy to Kubernetes (coming soon)
kapok tenant            # Manage tenants (coming soon)
```

---

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

---

## Documentation

- **[Quick Start Guide](./docs/quickstart.md)** - Get started in 5 minutes
- **[Installation Guide](./docs/installation.md)** - Platform-specific
  installation
- **[Architecture](./bmad-output/planning-artifacts/architecture.md)** - System
  design
- **[Product Requirements](./bmad-output/planning-artifacts/prd.md)** - Feature
  specifications
- **[Epics & Stories](./bmad-output/planning-artifacts/epics.md)** - Development
  roadmap

---

## Architecture

Voir la documentation complète d'architecture dans
`_bmad-output/planning-artifacts/`.

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
│   ├── config/             # Configuration structs
│   └── codegen/            # SDK generation
│       ├── typescript/     # TypeScript SDK generator
│       └── react/          # React hooks generator
├── docs/                   # Documentation
│   ├── quickstart.md       # Quick start guide
│   └── installation.md     # Installation guide
├── examples/               # Example projects
│   └── quickstart/         # Quick start example
├── deployments/            # Deployment configurations
│   └── helm/               # Helm charts
├── testdata/               # Test fixtures
└── scripts/                # Build and deployment scripts
```

---

## Development

### Prerequisites

- Go 1.21+
- PostgreSQL 12+
- Node.js 18+ (for SDK generation)
- Docker (optional, for local development)
- Kubernetes cluster (optional, for deployment)

### Building

```bash
go build -o kapok ./cmd/kapok
```

### Testing

```bash
go test ./...
```

### Running Tests with Coverage

```bash
go test ./... -cover
```

---

## Contributing

Contributions are welcome! Please read our contributing guidelines (coming
soon).

---

## License

_To be determined_

---

## Contact

- GitHub: https://github.com/kapok/kapok
- Documentation: [docs/](./docs/)
- Issues: https://github.com/kapok/kapok/issues
- Discussions: https://github.com/kapok/kapok/discussions

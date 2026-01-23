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

---
stepsCompleted: [1, 2, 3]
inputDocuments:
    - "/home/superz/kapok/_bmad-output/planning-artifacts/prd.md"
    - "/home/superz/kapok/_bmad-output/planning-artifacts/architecture.md"
---

# Kapok - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for Kapok,
decomposing the requirements from the PRD and Architecture into implementable
stories.

## Requirements Inventory

### Functional Requirements

**MVP Core Features (3 mois / 10 semaines) :**

**FR01:** Multi-Tenant Foundation - Schema-per-tenant isolation dans PostgreSQL
avec tenant router automatique et database auto-provisioning

**FR02:** Tenant Management API - Create, list, delete tenants via CLI et API
avec provisioning en <30 secondes

**FR03:** GraphQL Engine - Auto-génération de schéma GraphQL depuis
introspection PostgreSQL incluant types, queries, mutations, et relations
auto-détectées

**FR04:** GraphQL Queries & Mutations - Support CRUD complet avec filtering,
sorting, et pagination basiques

**FR05:** CLI Developer-Friendly - Commands `kapok init`, `kapok dev`,
`kapok deploy`, `kapok tenant [create|list|delete]`

**FR06:** Local Development Environment - `kapok dev` lance environnement
PostgreSQL + GraphQL Playground local

**FR07:** Kubernetes Deployment - Génération automatique de Helm charts et
déploiement one-command sur EKS/GKE

**FR08:** Auto-Scaling - Horizontal Pod Autoscaler (HPA) basé sur CPU/memory
activé par défaut

**FR09:** TLS/SSL Automatique - Let's Encrypt integration pour HTTPS automatique

**FR10:** TypeScript SDK - Auto-génération SDK TypeScript avec types
synchronisés automatiquement avec DB schema

**FR11:** React Hooks - Auto-génération React hooks basiques pour
queries/mulations GraphQL

**FR12:** Zero-Config Defaults - Configuration minimale requise, smart defaults
pour 90% des use cases

**FR13:** Documentation Complète - Quick start (5 min), guides, API reference
permettant onboarding autonome

**FR14:** Web Console UI (inspiré PocketBase/Supabase) - Interface web admin
pour gérer tenants, visualiser metrics, logs, et GraphQL Playground intégré

**Growth Features (Post-MVP / 4-6 mois) :**

**FR15:** DB-per-Tenant Option - Support isolation database-per-tenant en plus
de schema-per-tenant

**FR16:** Auto-Migration - Migration automatique schema → DB-per-tenant basée
sur usage thresholds

**FR17:** Configurable Isolation Levels - User peut choisir niveau d'isolation
par tenant

**FR18:** Parent-Child Tenant Hierarchies - Support organization → project →
team hierarchy

**FR19:** Real-Time WebSocket Subscriptions - GraphQL subscriptions sur
WebSocket avec <100ms latency

**FR20:** PostgreSQL LISTEN/NOTIFY - Integration LISTEN/NOTIFY pour real-time
data changes

**FR21:** Row-Level Security (RLS) - PostgreSQL RLS policies pour fine-grained
access control

**FR22:** Role-Based Access Control (RBAC) - User roles avec permissions
hierarchiques

**FR23:** Policy Injection - SQL policy injection automatique dans queries

**FR24:** Audit Trail - Logging automatique de tous data access et modifications

**FR25:** Visual Schema Builder - GUI pour créer/modifier database schemas

**FR26:** GraphQL Playground Intégré - Web-based GraphQL IDE avec autocomplete

**FR27:** Performance Profiler - Query performance analysis et optimization
suggestions

**FR28:** Time-Travel Queries - `kapok time-travel` pour query historical data

**FR29:** Multi-Cloud Optimizations - AWS/GCP/Azure specific optimizations et
features

**FR30:** Advanced Auto-Scaling - Vertical Pod Autoscaler (VPA) + custom metrics
scaling

**FR31:** Cost Optimization - Automated cost tracking, recommendations, tenant
hibernation

**FR32:** Observability Dashboards - Pre-configured Grafana dashboards pour
platform et per-tenant metrics

**Vision Features (6-12 mois) :**

**FR33:** Marketplace Intégrations - Stripe, Twilio, SendGrid, OAuth providers

**FR34:** CLI Plugins - Community-contributed CLI extensions

**FR35:** White-Label Options - Rebranding pour agencies/consultants

**FR36:** AI-Powered Suggestions - `kapok ai-suggest` pour architecture
recommendations

**FR37:** Auto-Query Optimization - IA-driven GraphQL query optimization

**FR38:** Anomaly Detection - Intelligent alerting pour performance/security
anomalies

**FR39:** Compliance Automation - Automated GDPR, HIPAA, SOC2 compliance packs

**FR40:** Penetration Testing-as-Service - Automated security testing

**FR41:** Multi-Region Active-Active - Geographic redundancy avec data
replication

**FR42:** Data Residency Enforcement - Garantir data ne quitte jamais chosen
region

### Non-Functional Requirements

**Performance :**

**NFR01:** GraphQL Query Latency - Simple queries p50 <50ms, p95 <100ms, p99
<200ms

**NFR02:** GraphQL Mutation Latency - Mutations p95 <150ms

**NFR03:** Real-Time Latency - WebSocket message delivery <100ms from DB change
to client

**NFR04:** Concurrent Connections - Support minimum 1000 concurrent WebSocket
connections per instance

**NFR05:** Throughput - Support minimum 10,000 messages/sec per instance

**NFR06:** Cache Hit Rate - Redis query cache hit rate >80% pour repeated
queries

**NFR07:** Database Load Reduction - 40-60% reduction in PostgreSQL queries via
Redis caching

**Security :**

**NFR08:** Encryption at Rest - AES-256 encryption pour all databases et backups

**NFR09:** Encryption in Transit - TLS 1.3+ pour all network communication

**NFR10:** Secrets Management - Integration avec HashiCorp Vault, Kubernetes
secrets encryption

**NFR11:** Key Rotation - Automated encryption key rotation policies

**NFR12:** SQL Injection Prevention - Parameterized queries, no string
concatenation

**NFR13:** GraphQL Input Validation - Input validation et sanitization on all
mutations

**NFR14:** Query Complexity Analysis - Limit GraphQL query depth/cost pour
prevent DDOS

**NFR15:** Authentication - JWT tokens avec short expiration (24h), refresh
token rotation

**NFR16:** MFA Support - TOTP et WebAuthn multi-factor authentication

**NFR17:** Password Security - bcrypt hashing, strong password policies

**NFR18:** Rate Limiting - Redis-based distributed rate limiting per tenant

**Scalability :**

**NFR19:** Horizontal Scaling - Auto-scale from 2 to 50 replicas based on load

**NFR20:** Vertical Scaling - VPA recommendations pour resource adjustments

**NFR21:** Tenant Scalability - Support 100+ tenants in production (MVP), 1000+
tenants (Growth)

**NFR22:** Database Scalability - Support migration de schema-per-tenant à
DB-per-tenant seamlessly

**Reliability :**

**NFR23:** Uptime Target - 99.5% uptime (MVP), 99.9% uptime (Growth)

**NFR24:** Zero-Downtime Upgrades - Rolling deployments, blue-green, canary
releases

**NFR25:** Automated Rollback - Instant rollback sur health check failures

**NFR26:** Health Checks - Health checks à tous niveaux (pod/service/tenant)

**Disaster Recovery :**

**NFR27:** RPO Standard Tier - 6 hours (snapshots every 6h)

**NFR28:** RPO Professional Tier - 1 hour (hourly incremental backups)

**NFR29:** RPO Enterprise Tier - <5 minutes (synchronous replication)

**NFR30:** RTO Standard Tier - 4 hours (restore from snapshot)

**NFR31:** RTO Professional Tier - 1 hour (automated failover to hot standby)

**NFR32:** RTO Enterprise Tier - <15 minutes (automatic failover avec load
balancer)

**NFR33:** Backup Retention - Full daily backups (30 days), incremental hourly
(7 days)

**NFR34:** Point-in-Time Recovery - Restore to any second within retention
period

**NFR35:** Backup Encryption - AES-256 encryption pour all backups

**NFR36:** Immutable Backups - WORM storage pour ransomware protection

**Observability :**

**NFR37:** Metric Collection - Prometheus metrics pour all services avec 15s
scrape interval

**NFR38:** Log Aggregation - Structured logging (JSON) avec tenant_id dans tous
logs

**NFR39:** Distributed Tracing - OpenTelemetry avec 10% sampling rate

**NFR40:** Alerting - Tiered alerts (Critical → PagerDuty, Warning → Slack, Info
→ Metrics only)

**NFR41:** Dashboard Auto-Provisioning - Grafana dashboards auto-deployed et
pre-configured

**Compliance :**

**NFR42:** ISO 27001 Readiness - Security controls, risk management, audit
mechanisms

**NFR43:** FedRAMP Compliance - NIST 800-53 controls, continuous monitoring

**NFR44:** SOC 2 Type II - Audit trails, access control logs, change management
tracking

**NFR45:** GDPR Compliance - Data residency, right to be forgotten, data
portability

**NFR46:** HIPAA Compliance - BAA-capable infrastructure, PHI access logging,
6-year retention

**NFR47:** PCI-DSS Compliance - Network segmentation, vulnerability scanning,
cardholder data encryption

**Developer Experience :**

**NFR48:** Onboarding Time - <5 minutes from install to deployed backend

**NFR49:** CLI Response Time - <500ms pour local commands

**NFR50:** Documentation Quality - Coverage permettant autonomous onboarding
sans support

**NFR51:** Error Messages - Actionable error messages, NOT cryptic Kubernetes
errors

**NFR52:** Type Safety - End-to-end type safety (DB → API → Frontend) avec
auto-sync

### Additional Requirements

**From Architecture Document :**

**AR01:** Monorepo Structure - Single Go monorepo avec `cmd/`, `internal/`,
`pkg/`, `deployments/` organization

**AR02:** Tech Stack - Go (backend), Cobra (CLI), gqlgen (GraphQL), Viper
(config), zerolog (logging)

**AR03:** Context Propagation - `context.Context` pour tenant_id propagation à
tous layers (CRITIQUE)

**AR04:** JWT Strategy - JWT tokens avec tenant_id claims pour stateless auth

**AR05:** Casbin RBAC - Casbin library pour hierarchical RBAC avec PostgreSQL
adapter

**AR06:** Helm Charts - Per-service Helm charts + umbrella chart pour platform
deployment

**AR07:** KEDA Auto-Scaling - Custom metrics scaling (PostgreSQL connections,
query latency) en plus HPA

**AR08:** Redis Strategy - Sentinel (MVP) → Cluster(Growth) pour HA et
horizontal scaling

**AR09:** Real-Time Architecture - PostgreSQL LISTEN/NOTIFY → GraphQL Engine →
Redis Pub/Sub → WebSocket clients

**AR10:** Observability Stack - Auto-deploy Prometheus, Grafana, Loki, Jaeger
avec platform

**AR11:** Naming Conventions - snake_case (SQL), camelCase (Go/JSON), PascalCase
(exported Go)

**AR12:** Error Handling - Error wrapping avec `%w`, custom error types par
domain

**AR13:** Structured Logging - zerolog JSON logging ONLY, jamais fmt.Println()

**AR14:** Testing Strategy - Co-located `_test.go`, dockertest pour integration,
table-driven tests

**AR15:** Config Validation - Fail-fast config validation au startup avec clear
error messages

**AR16:** OpenTelemetry Integration - Context propagation, distributed tracing,
metrics export dès MVP

**AR17:** Query Complexity Limits - GraphQL complexity analysis MANDATORY
(prevent DDOS)

**AR18:** Schema Caching - Redis caching de GraphQL schemas (5 min TTL,
invalidate on DDL)

**AR19:** PostgreSQL Introspection - Dynamic schema generation from PostgreSQL
information_schema

**AR20:** Deployment Target - Kubernetes 1.24+, PostgreSQL 13-16, Redis 6-7, Go
1.21+

**Added Requirement (User Request) :**

**AR21:** Web Console UI - Admin interface inspirée PocketBase/Supabase pour
dashboard, tenant management, metrics visualization, logs, GraphQL Playground
(UX design à créer ultérieurement)

### FR Coverage Map

**Functional Requirements Mapping:**

- FR01 → Epic 2: Multi-Tenant Core Infrastructure (Schema-per-tenant isolation)
- FR02 → Epic 2: Multi-Tenant Core Infrastructure (Tenant Management API)
- FR03 → Epic 3: GraphQL API Auto-Generation (Schema introspection & generation)
- FR04 → Epic 3: GraphQL API Auto-Generation (CRUD queries & mutations)
- FR05 → Epic 1: Project Foundation & Local Development (CLI commands)
- FR06 → Epic 1: Project Foundation & Local Development (Local dev environment)
- FR07 → Epic 4: Kubernetes Deployment & Scaling (Helm charts deployment)
- FR08 → Epic 4: Kubernetes Deployment & Scaling (Auto-scaling HPA)
- FR09 → Epic 4: Kubernetes Deployment & Scaling (TLS/SSL automatic)
- FR10 → Epic 1: Project Foundation & Local Development (TypeScript SDK)
- FR11 → Epic 1: Project Foundation & Local Development (React hooks)
- FR12 → Epic 1: Project Foundation & Local Development (Zero-config defaults)
- FR13 → Epic 1: Project Foundation & Local Development (Documentation)
- FR14 → Epic 6: Web Console UI (Admin interface)
- FR15 → Epic 9: Advanced Multi-Tenancy (DB-per-tenant option)
- FR16 → Epic 9: Advanced Multi-Tenancy (Auto-migration schema → DB)
- FR17 → Epic 9: Advanced Multi-Tenancy (Configurable isolation levels)
- FR18 → Epic 9: Advanced Multi-Tenancy (Parent-child hierarchies)
- FR19 → Epic 10: Real-Time & WebSocket Subscriptions (GraphQL subscriptions)
- FR20 → Epic 10: Real-Time & WebSocket Subscriptions (LISTEN/NOTIFY
  integration)
- FR21 → Epic 11: Advanced RBAC & Row-Level Security (RLS policies)
- FR22 → Epic 11: Advanced RBAC & Row-Level Security (RBAC implementation)
- FR23 → Epic 11: Advanced RBAC & Row-Level Security (Policy injection)
- FR24 → Epic 11: Advanced RBAC & Row-Level Security (Audit trail)
- FR25 → Epic 12: Visual Schema Builder & Advanced DX (Visual schema builder)
- FR26 → Epic 12: Visual Schema Builder & Advanced DX (GraphQL Playground)
- FR27 → Epic 12: Visual Schema Builder & Advanced DX (Performance profiler)
- FR28 → Epic 12: Visual Schema Builder & Advanced DX (Time-travel queries)
- FR29 → Epic 4: Kubernetes Deployment & Scaling (Multi-cloud optimizations)
- FR30 → Epic 4: Kubernetes Deployment & Scaling (VPA advanced scaling)
- FR31 → Epic 5: Observability & Monitoring (Cost optimization)
- FR32 → Epic 5: Observability & Monitoring (Observability dashboards)
- FR33-42 → Vision Phase (Future epics not in current scope)

**Non-Functional Requirements Mapping:**

- NFR01-07 (Performance) → Epic 3: GraphQL API Auto-Generation + Epic 5:
  Observability
- NFR08-18 (Security) → Epic 8: Security & Compliance Foundations + Epic 2:
  Multi-Tenant Core
- NFR19-22 (Scalability) → Epic 4: Kubernetes Deployment & Scaling
- NFR23-26 (Reliability) → Epic 4: Kubernetes Deployment & Scaling
- NFR27-36 (Disaster Recovery) → Epic 7: Backup, Recovery & High Availability
- NFR37-41 (Observability) → Epic 5: Observability & Monitoring
- NFR42-47 (Compliance) → Epic 8: Security & Compliance Foundations
- NFR48-52 (Developer Experience) → Epic 1: Project Foundation & Local
  Development

**Additional Requirements Mapping:**

- AR01-03 (Monorepo, Tech Stack, Context) → Epic 1: Project Foundation
- AR04-05 (JWT, RBAC) → Epic 2: Multi-Tenant Core + Epic 11: Advanced RBAC
- AR06-07 (Helm, KEDA) → Epic 4: Kubernetes Deployment
- AR08-09 (Redis, Real-Time) → Epic 10: Real-Time & WebSocket Subscriptions
- AR10, AR16 (Observability, OpenTelemetry) → Epic 5: Observability & Monitoring
- AR11-15 (Patterns, Testing, Config) → Epic 1: Project Foundation
- AR17-19 (Query Complexity, Caching, Introspection) → Epic 3: GraphQL API
- AR20 (Deployment Targets) → Epic 4: Kubernetes Deployment
- AR21 (Web Console UI) → Epic 6: Web Console UI

## Epic List

### Epic 1: Project Foundation & Local Development

**Goal:** Frontend developers peuvent initialiser un projet Kapok, développer
localement avec un backend GraphQL fonctionnel, et bénéficier d'un SDK
TypeScript auto-généré avec zero configuration.

**FRs Covered:** FR05, FR06, FR10, FR11, FR12, FR13\
**NFRs Covered:** NFR48-NFR52 (Developer Experience)\
**ARs Covered:** AR01-AR03, AR11-AR15, AR20

**User Outcome:** Un développeur frontend peut exécuter `kapok init`, obtenir un
projet configuré automatiquement, lancer `kapok dev` pour un environnement local
avec PostgreSQL + GraphQL Playground, et utiliser un SDK TypeScript type-safe
pour développer son application.

**Value Delivered:**

- Zero-config project initialization
- Local development environment complet
- TypeScript SDK + React hooks auto-générés
- Documentation permettant onboarding autonome en <5 minutes

---

### Epic 2: Multi-Tenant Core Infrastructure

**Goal:** Platform operators peuvent créer, gérer et isoler des tenants avec
database auto-provisioning, garantissant une isolation complète entre tenants
pour sécurité et compliance.

**FRs Covered:** FR01, FR02\
**NFRs Covered:** NFR08-NFR11, NFR21-NFR22 (Security, Scalability)\
**ARs Covered:** AR04-AR05

**User Outcome:** Un operator peut exécuter `kapok tenant create`, obtenir un
tenant isolé en <30 secondes avec son propre schema PostgreSQL, JWT
authentication fonctionnelle, et RBAC Casbin configuré.

**Value Delivered:**

- Schema-per-tenant isolation (MVP)
- Tenant provisioning API complet
- Tenant router automatique
- JWT + Casbin RBAC foundations
- Strong isolation guarantees

---

### Epic 3: GraphQL API Auto-Generation

**Goal:** Les tenants obtiennent automatiquement une API GraphQL complète depuis
leur schema PostgreSQL avec queries, mutations, relations auto-détectées, et
performance optimisée.

**FRs Covered:** FR03, FR04\
**NFRs Covered:** NFR01-NFR02, NFR12-NFR14 (Performance, Security)\
**ARs Covered:** AR17-AR19

**User Outcome:** Dès qu'un tenant crée une table PostgreSQL, l'API GraphQL est
automatiquement générée avec types, queries, mutations, filtering, sorting, et
relations foreign-key détectées, le tout avec <100ms p95 latency.

**Value Delivered:**

- PostgreSQL introspection automatique
- GraphQL schema generation dynamique
- CRUD queries/mutations complètes
- Relations auto-détectées
- Query complexity analysis (anti-DDOS)
- Redis caching (80%+ hit rate)

---

### Epic 4: Kubernetes Deployment & Scaling

**Goal:** Users peuvent déployer Kapok en production sur Kubernetes (AWS EKS,
GCP GKE, Azure AKS) en une commande avec auto-scaling, TLS automatique, et
multi-cloud support.

**FRs Covered:** FR07, FR08, FR09, FR29, FR30\
**NFRs Covered:** NFR19-NFR20, NFR23-NFR26 (Scalability, Reliability)\
**ARs Covered:** AR06-AR07, AR20

**User Outcome:** Un user exécute `kapok deploy`, la commande détecte le cloud
provider, génère les Helm charts optimisés, déploie sur Kubernetes, configure
auto-scaling (HPA + KEDA), et active TLS Let's Encrypt automatiquement.

**Value Delivered:**

- One-command deployment
- Helm charts auto-générés
- Multi-cloud compatibility (AWS/GCP/Azure)
- HPA + KEDA auto-scaling
- TLS/SSL automatique
- Zero-downtime upgrades

---

### Epic 5: Observability & Monitoring

**Goal:** Operators ont visibilité complète sur platform health, per-tenant
metrics, distributed tracing, et peuvent troubleshoot efficacement avec
dashboards pre-configured.

**FRs Covered:** FR31, FR32\
**NFRs Covered:** NFR37-NFR41 (Observability)\
**ARs Covered:** AR10, AR16

**User Outcome:** Après deployment, Prometheus, Grafana, Loki, et Jaeger sont
automatiquement déployés et configurés avec dashboards montrant platform health,
per-tenant CPU/memory/latency, distributed traces, et alerting PagerDuty/Slack.

**Value Delivered:**

- Auto-deployed observability stack
- Pre-configured Grafana dashboards
- Per-tenant metrics isolation
- Distributed tracing (OpenTelemetry)
- Tiered alerting (Critical/Warning/Info)
- Cost tracking per tenant

---

### Epic 6: Web Console UI

**Goal:** Users peuvent gérer tenants, visualiser metrics/logs, explorer GraphQL
API, et configurer platform via interface web intuitive inspirée de PocketBase
et Supabase.

**FRs Covered:** FR14\
**ARs Covered:** AR21

**User Outcome:** Users accèdent à console web Kapok, voient dashboard avec tous
tenants, peuvent create/delete tenants en cliquant, visualiser real-time
metrics, consulter logs, et utiliser GraphQL Playground intégré pour tester
queries.

**Value Delivered:**

- Dashboard tenants overview
- Tenant management UI (create/configure/delete)
- Metrics & logs visualization
- GraphQL Playground intégré
- Settings & configuration UI
- User-friendly alternative au CLI

---

### Epic 7: Backup, Recovery & High Availability

**Goal:** Platform garantit data safety avec backups automatiques multi-tier,
point-in-time recovery, et disaster recovery capabilities respectant RPO/RTO par
tier.

**NFRs Covered:** NFR27-NFR36 (Disaster Recovery)

**User Outcome:** Backups sont automatiquement planifiés selon tier
(6h/1h/<5min), stockés encrypted et immutable, avec point-in-time recovery
disponible via `kapok restore --timestamp`, et automated failover pour
Enterprise tier.

**Value Delivered:**

- Automated backup scheduling (daily full + hourly incremental)
- Point-in-time recovery
- AES-256 encrypted backups
- Immutable storage (WORM)
- Multi-region replication (Enterprise)
- Automated failover (Enterprise)

---

### Epic 8: Security & Compliance Foundations

**Goal:** Platform respecte standards enterprise security (OWASP Top 10) et
établit foundations pour compliance certifications (ISO 27001, SOC 2, GDPR,
HIPAA, FedRAMP, PCI-DSS).

**NFRs Covered:** NFR08-NFR18 (Security), NFR42-NFR47 (Compliance)

**User Outcome:** Platform est sécurisé by-default avec encryption
at-rest/in-transit, parameterized queries, input validation, rate limiting,
audit logging, et documentation pour compliance readiness.

**Value Delivered:**

- OWASP Top 10 mitigations
- Encryption (AES-256 at rest, TLS 1.3+ in transit)
- Authentication (JWT, MFA support)
- Rate limiting (Redis-based)
- Audit trail immutable
- Compliance documentation (ISO/SOC2/GDPR/HIPAA)

---

### Epic 9: Advanced Multi-Tenancy (Growth Phase)

**Goal:** Platform supporte DB-per-tenant isolation avec auto-migration seamless
depuis schema-per-tenant quand usage thresholds sont atteints, et hierarchies
parent-child.

**FRs Covered:** FR15, FR16, FR17, FR18

**User Outcome:** Quand un tenant dépasse thresholds (storage, connections,
QPS), platform migre automatiquement vers dedicated PostgreSQL database avec
zero-downtime, et supporte organization → project → team hierarchies.

**Value Delivered:**

- DB-per-tenant option
- Zero-downtime migration (schema → DB)
- Configurable isolation levels
- Parent-child tenant hierarchies
- Usage-based auto-migration
- Dedicated instance option (Enterprise)

---

### Epic 10: Real-Time & WebSocket Subscriptions (Growth Phase)

**Goal:** Tenants peuvent utiliser GraphQL subscriptions real-time sur WebSocket
avec <100ms latency pour data changes, powered by PostgreSQL LISTEN/NOTIFY +
Redis Pub/Sub.

**FRs Covered:** FR19, FR20\
**NFRs Covered:** NFR03-NFR05 (Real-Time Performance)\
**ARs Covered:** AR08-AR09

**User Outcome:** Developers ajoutent GraphQL subscription dans leur app,
établissent WebSocket connection, et reçoivent real-time updates (<100ms) quand
data change dans PostgreSQL via LISTEN/NOTIFY → Redis Pub/Sub pipeline.

**Value Delivered:**

- GraphQL subscriptions support
- PostgreSQL LISTEN/NOTIFY integration
- Redis Pub/Sub for fan-out
- <100ms end-to-end latency
- Tenant-aware channels
- 1000+ concurrent connections per instance

---

### Epic 11: Advanced RBAC & Row-Level Security (Growth Phase)

**Goal:** Tenants ont fine-grained access control avec Row-Level Security
PostgreSQL, RBAC policies complexes, policy injection automatique, et audit
trail comprehensive.

**FRs Covered:** FR21, FR22, FR23, FR24

**User Outcome:** Administrators définissent RBAC policies (Organization →
Project → Team → User), RLS policies sont automatiquement injectées dans queries
SQL, et tout access/modification est logged dans audit trail immutable.

**Value Delivered:**

- PostgreSQL Row-Level Security
- Hierarchical RBAC (Casbin)
- SQL policy injection automatique
- Audit trail immutable
- Fine-grained permissions
- Compliance-ready access logs

---

### Epic 12: Visual Schema Builder & Advanced DX (Growth Phase)

**Goal:** Users ont outils graphiques pour schema design, GraphQL query testing,
performance profiling, et time-travel debugging pour améliorer developer
experience.

**FRs Covered:** FR25, FR26, FR27, FR28

**User Outcome:** Users accèdent à visual schema builder pour créer tables via
drag-and-drop, testent queries dans GraphQL Playground intégré, profilent
performance avec recommendations, et utilisent `kapok time-travel` pour debug
historical data.

**Value Delivered:**

- Visual schema builder GUI
- Integrated GraphQL Playground
- Query performance profiler
- Time-travel query debugging
- Schema migration assistance
- Auto-optimization suggestions

## Epic 1: Project Foundation & Local Development

**Goal:** Frontend developers peuvent initialiser un projet Kapok, développer localement avec un backend GraphQL fonctionnel, et bénéficier d'un SDK TypeScript auto-généré avec zero configuration.

### Story 1.1: Initialize Monorepo Structure

As a **platform developer**,  
I want to **set up the Go monorepo structure with proper organization**,  
So that **code is organized following best practices and architectural guidelines**.

**Acceptance Criteria:**

**Given** the project repository is empty  
**When** the monorepo structure is created  
**Then** directories `cmd/`, `internal/`, `pkg/`, `deployments/`, `testdata/`, `scripts/` exist  
**And** `go.mod` is initialized with Go 1.21+ and module name `github.com/kapok/kapok`  
**And** `.gitignore` excludes binaries, `.env`, and IDE files  
**And** `README.md` contains project overview and quick start placeholder

### Story 1.2: Implement Cobra CLI Foundation

As a **developer**,  
I want **a Cobra-based CLI with core commands structure**,  
So that **users can execute `kapok init`, `kapok dev`, `kapok deploy`, `kapok tenant`**.

**Acceptance Criteria:**

**Given** the monorepo structure exists  
**When** Cobra CLI is implemented in `cmd/kapok/`  
**Then** `kapok --help` displays available commands  
**And** `kapok --version` displays version information  
**And** All commands (`init`, `dev`, `deploy`, `tenant`) exist with placeholder implementations  
**And** CLI follows naming conventions (snake_case files, PascalCase structs)  
**And** Unit tests cover CLI command routing

**Given** CLI is built  
**When** `kapok init --help` is executed  
**Then** command-specific help text is displayed with flags and examples

### Story 1.3: Implement Viper Configuration Management

As a **developer**,  
I want **Viper-based configuration loading with precedence hierarchy**,  
So that **configuration can be sourced from CLI flags, ENV vars, and YAML files**.

**Acceptance Criteria:**

**Given** Viper library is integrated  
**When** configuration is loaded  
**Then** config hierarchy is: CLI flags > ENV vars > `kapok.yaml` > `~/.kapok/config.yaml` > defaults  
**And** `internal/config/config.go` defines Config struct with validation  
**And** Config validation fails fast at startup with clear error messages  
**And** ENV vars use `KAPOK_` prefix (e.g., `KAPOK_DATABASE_HOST`)  
**And** Secrets are NEVER in config files (ENV vars only)

**Given** invalid configuration  
**When** `kapok` command is executed  
**Then** clear error message explains what's missing/invalid  
**And** process exits with non-zero status

### Story 1.4: Setup Structured Logging with zerolog

As a **developer**,  
I want **structured JSON logging with zerolog**,  
So that **all logs are queryable and include tenant/request context**.

**Acceptance Criteria:**

**Given** zerolog library is integrated  
**When** logging is configured  
**Then** production mode outputs JSON logs  
**And** development mode outputs pretty console logs  
**And** log levels are: DEBUG (dev), INFO (prod), ERROR (always)  
**And** logger is injected, NEVER global variables  
**And** all logs include timestamp, level, message  
**And** NO `fmt.Println()` or `log.Print()` usage anywhere

**Given** a function needs to log  
**When** context contains `tenant_id` and `request_id`  
**Then** logs automatically include those fields  
**And** logs are structured: `{"level":"info","tenant_id":"123","request_id":"abc","msg":"..."}`

### Story 1.5: Implement `kapok init` Command

As a **frontend developer**,  
I want **`kapok init` to create a new project with zero configuration**,  
So that **I can start developing immediately without manual setup**.

**Acceptance Criteria:**

**Given** kapok CLI is installed  
**When** `kapok init my-project` is executed in an empty directory  
**Then** `kapok.yaml` configuration file is created with smart defaults  
**And** `.env.example` template is created  
**And** `README.md` is generated with project-specific quick start  
**And** `docs/` folder contains basic architecture documentation  
**And** command completes in <5 seconds  
**And** success message displays next steps

**Given** `kapok init` in non-empty directory  
**When** command is executed  
**Then** error message warns about overwriting files  
**And** `--force` flag option is suggested  
**And** process exits safely without changes

### Story 1.6: Implement `kapok dev` - Local PostgreSQL

As a **frontend developer**,  
I want **`kapok dev` to launch a local PostgreSQL database**,  
So that **I can develop without external database setup**.

**Acceptance Criteria:**

**Given** Docker is installed and running  
**When** `kapok dev` is executed  
**Then** PostgreSQL container is started using dockertest  
**And** database is accessible on `localhost:5432`  
**And** default credentials are in `.env` file  
**And** database is automatically migrated (when schema exists)  
**And** health check confirms database is ready before proceeding  
**And** logs display connection string (with password masked)

**Given** PostgreSQL container already running  
**When** `kapok dev` is executed again  
**Then** existing container is reused (not recreated)  
**And** connection is verified

**Given** Docker is not installed  
**When** `kapok dev` is executed  
**Then** clear error message explains Docker requirement  
**And** link to Docker installation docs is provided

### Story 1.7: Implement Type-Safe SDK Generator

As a **frontend developer**,  
I want **TypeScript SDK auto-generated from my database schema**,  
So that **I have end-to-end type safety from DB to frontend**.

**Acceptance Criteria:**

**Given** PostgreSQL schema exists  
**When** SDK generator runs  
**Then** TypeScript types are generated for all tables  
**And** CRUD functions are generated for each entity  
**And** Types match database schema exactly (snake_case DB → camelCase TS)  
**And** Generated SDK is written to `sdk/typescript/` directory  
**And** `package.json` is created with proper dependencies  
**And** SDK exports a client class with typed methods

**Given** database schema changes  
**When** SDK generator re-runs  
**Then** types are updated automatically  
**And** breaking changes are flagged in console warnings

### Story 1.8: Implement React Hooks Generator

As a **React developer**,  
I want **React hooks auto-generated for GraphQL queries**,  
So that **I can use data fetching with minimal boilerplate**.

**Acceptance Criteria:**

**Given** GraphQL schema exists  
**When** React hooks generator runs  
**Then** custom hooks are generated for each query (`useUsers`, `useUserById`)  
**And** mutation hooks are generated (`useCreateUser`, `useUpdateUser`)  
**And** hooks use React Query for caching and state management  
**And** TypeScript types are included for all hooks  
**And** generated code is in `sdk/react/hooks/` directory  
**And** proper imports and exports are configured

**Given** a frontend component  
**When** using generated hooks  
**Then** autocomplete works for all fields  
**And** TypeScript errors show for incorrect usage  
**And** data fetching, loading, and error states are handled

### Story 1.9: Create Quick Start Documentation

As a **new user**,  
I want **comprehensive quick start documentation**,  
So that **I can onboard in <5 minutes without external support**.

**Acceptance Criteria:**

**Given** Kapok is installed  
**When** user accesses documentation  
**Then** quick start guide exists at `docs/quickstart.md`  
**And** installation instructions cover all platforms (macOS, Linux, Windows)  
**And** `kapok init` workflow is documented with screenshots  
**And** `kapok dev` usage is explained  
**And** first GraphQL query example is provided  
**And** troubleshooting section covers common issues  
**And** all code examples are tested and working

**Given** a complete beginner  
**When** following quick start step-by-step  
**Then** they reach a working GraphQL API in <5 minutes  
**And** no external documentation is needed

### Story 1.10: Implement CLI Testing Strategy

As a **platform developer**,  
I want **comprehensive CLI tests**,  
So that **all commands are tested and reliable**.

**Acceptance Criteria:**

**Given** CLI codebase exists  
**When** test suite runs  
**Then** unit tests exist for all Cobra commands  
**And** integration tests verify end-to-end workflows  
**And** tests use `testify` for assertions  
**And** CLI output is captured via injected `io.Writer`  
**And** filesystem operations use mock/temp directories  
**And** all tests pass in CI pipeline  
**And** code coverage >60% for `cmd/kapok/`

**Given** a CLI command is modified  
**When** tests run  
**Then** breaking changes are caught by tests  
**And** regression is prevented


## Epic 2: Multi-Tenant Core Infrastructure

**Goal:** Platform operators peuvent créer, gérer et isoler des tenants avec database auto-provisioning, garantissant une isolation complète entre tenants pour sécurité et compliance.

### Story 2.1: Implement PostgreSQL Schema-Per-Tenant Isolation

As a **platform operator**,  
I want **schema-per-tenant isolation in PostgreSQL**,  
So that **each tenant's data is completely isolated from others**.

**Acceptance Criteria:**

**Given** PostgreSQL database is running  
**When** a new tenant is created  
**Then** dedicated schema `tenant_<id>` is created  
**And** schema owner is separate role with limited permissions  
**And** row-level security SET `app.tenant_id` is configured  
**And** all tables in schema follow naming conventions (snake_case, plural)  
**And** foreign key constraints use explicit naming `fk_<table>_<column>`

**Given** tenant `tenant_123` exists  
**When** querying without setting session variable  
**Then** access is denied with clear error message

### Story 2.2: Implement Tenant Provisioning API

As a **platform operator**,  
I want **tenant provisioning via CLI and API**,  
So that **new tenants can be created in <30 seconds**.

**Acceptance Criteria:**

**Given** platform is deployed  
**When** `kapok tenant create --name=acme` is executed  
**Then** tenant is created with unique ID  
**And** PostgreSQL schema is provisioned  
**And** JWT secret is generated  
**And** tenant metadata stored in `tenants` table  
**And** provisioning completes in <30 seconds  
**And** success response includes tenant ID and credentials

**Given** tenant name already exists  
**When** create command is executed  
**Then** error indicates duplicate name  
**And** suggestion to use different name is provided

### Story 2.3: Implement Tenant Router Middleware

As a **platform developer**,  
I want **HTTP middleware that routes requests to correct tenant**,  
So that **multi-tenant isolation is enforced at network layer**.

**Acceptance Criteria:**

**Given** HTTP request with `tenant_id` in JWT  
**When** request is processed by middleware  
**Then** `tenant_id` is extracted from JWT  
**And** injected into `context.Context`  
**And** PostgreSQL session variable `app.tenant_id` is set  
**And** request proceeds to handler with context

**Given** HTTP request without valid JWT  
**When** request is processed  
**Then** 401 Unauthorized is returned  
**And** error response explains missing/invalid token

### Story 2.4: Implement JWT Authentication

As a **platform operator**,  
I want **JW

T-based authentication with tenant_id claims**,  
So that **stateless authentication is supported across services**.

**Acceptance Criteria:**

**Given** user credentials are valid  
**When** JWT is generated  
**Then** token includes claims: `sub` (user_id), `tenant_id`, `email`, `roles`, `exp` (24h)  
**And** token is signed with HMAC SHA256  
**And** signing secret is loaded from ENV variable  
**And** refresh token is generated (7 days expiry)

**Given** expired JWT  
**When** token validation is attempted  
**Then** error indicates token expired  
**And** refresh token flow is suggested

### Story 2.5: Implement Casbin RBAC Foundation

As a **platform developer**,  
I want **Casbin integrated for hierarchical RBAC**,  
So that **fine-grained permissions are supported**.

**Acceptance Criteria:**

**Given** Casbin library is integrated  
**When** RBAC model is defined  
**Then** model supports: subject (user/role), object (resource), action (read/write/delete), tenant  
**And** policies are stored in PostgreSQL  
**And** role hierarchies are supported (admin > developer > viewer)  
**And** enforcer is initialized at startup

**Given** user with role "admin" for tenant_123  
**When** permission check is performed  
**Then** admin has all permissions on tenant_123 resources  
**And** admin has NO permissions on tenant_456 resources

### Story 2.6: Implement Tenant List and Delete Commands

As a **platform operator**,  
I want **list and delete tenant commands**,  
So that **I can manage the full tenant lifecycle**.

**Acceptance Criteria:**

**Given** multiple tenants exist  
**When** `kapok tenant list` is executed  
**Then** all tenants are displayed with ID, name, created_at, status  
**And** output supports `--output=json` flag for scripting  
**And** pagination is supported for large lists

**Given** tenant exists  
**When** `kapok tenant delete <id>` is executed  
**Then** confirmation prompt is shown  
**And** after confirmation, tenant schema is dropped  
**And** tenant metadata is soft-deleted (not hard-deleted)  
**And** audit log records deletion

### Story 2.7: Implement Tenant Isolation Testing

As a **security engineer**,  
I want **automated isolation tests**,  
So that **tenant data leakage is impossible**.

**Acceptance Criteria:**

**Given** two tenants exist (tenant_123, tenant_456)  
**When** isolation test suite runs  
**Then** queries with `tenant_123` context NEVER return `tenant_456` data  
**And** cross-tenant queries are blocked by RLS  
**And** schema-level isolation is verified  
**And** JWT cannot be forged or tampered  
**And** all isolation tests pass in CI/CD pipeline

## Epic 3: GraphQL API Auto-Generation

**Goal:** Les tenants obtiennent automatiquement une API GraphQL complète depuis leur schema PostgreSQL avec queries, mutations, relations auto-détectées, et performance optimisée.

### Story 3.1: Implement PostgreSQL Schema Introspection

As a **platform developer**,  
I want **PostgreSQL schema introspection via information_schema**,  
So that **database structure can be dynamically discovered**.

**Acceptance Criteria:**

**Given** PostgreSQL schema exists with tables  
**When** introspection query runs  
**Then** all tables, columns, types, constraints are discovered  
**And** foreign key relationships are detected  
**And** enums, arrays, JSON types are identified  
**And** introspection result is cached in struct format  
**And** query completes in <100ms

### Story 3.2: Implement GraphQL Schema Generator

As a **platform developer**,  
I want **GraphQL schema auto-generated from PostgreSQL**,  
So that **types, queries, and mutations match database exactly**.

**Acceptance Criteria:**

**Given** PostgreSQL introspection data  
**When** GraphQL schema is generated  
**Then** GraphQL type created for each table (PascalCase)  
**And** fields created for each column (camelCase)  
**And** types map correctly (VARCHAR→String, INTEGER→Int, TIMESTAMP→DateTime)  
**And** foreign keys generate nested object fields  
**And** enums generate GraphQL enum types  
**And** schema is valid per GraphQL spec

### Story 3.3: Implement GraphQL Query Resolvers

As a **frontend developer**,  
I want **CRUD query resolvers auto-generated**,  
So that **I can query data without writing backend code**.

**Acceptance Criteria:**

**Given** table `users` exists  
**When** GraphQL schema is generated  
**Then** queries exist: `users`, `userById(id: ID!)`  
**And** filtering supported: `users(where: UserFilter)`  
**And** sorting supported: `users(orderBy: UserOrderBy)`  
**And** pagination supported: `users(offset: Int, limit: Int)`  
**And** queries return latest data (no stale cache without invalidation)

**Given** users query executed  
**When** response is received  
**Then** p95 latency is <100ms  
**And** nested relations are resolved (N+1 prevented by dataloader)

### Story 3.4: Implement GraphQL Mutation Resolvers

As a **frontend developer**,  
I want **mutation resolvers for create/update/delete**,  
So that **I can modify data via GraphQL**.

**Acceptance Criteria:**

**Given** table `users` exists  
**When** GraphQL schema is generated  
**Then** mutations exist: `createUser`, `updateUser`, `deleteUser`  
**And** input validation is automatic (required fields, types)  
**And** mutations return updated object  
**And** transactions are used (atomic operations)  
**And** tenant_id is automatically injected (user cannot specify)

**Given** createUser mutation  
**When** executed with valid input  
**Then** new user is created in correct tenant schema  
**And** p95 latency is <150ms

### Story 3.5: Implement Query Complexity Analysis

As a **platform operator**,  
I want **GraphQL query complexity limits**,  
So that **DDOS attacks via deeply nested queries are prevented**.

**Acceptance Criteria:**

**Given** GraphQL resolver is configured  
**When** complexity analysis is enabled  
**Then** max query depth is limited (default: 10)  
**And** max query cost is limited (default: 1000)  
**And** cost is calculated based on nesting and field count  
**And** overly complex queries are rejected with error

**Given** malicious nested query (depth > 10)  
**When** query is executed  
**Then** error returned: "Query complexity exceeds limit"  
**And** server resources are protected

### Story 3.6: Implement Redis GraphQL Schema Caching

As a **platform developer**,  
I want **Redis caching of generated GraphQL schemas**,  
So that **introspection overhead is minimized**.

**Acceptance Criteria:**

**Given** GraphQL schema generated for tenant  
**When** schema is cached  
**Then** Redis key is `schema:<tenant_id>`  
**And** TTL is 5 minutes  
**And** cache hit rate >80% in normal operations  
**And** cache is invalidated on DDL operations

**Given** cached schema exists  
**When** same tenant requests schema  
**Then** schema loaded from cache (<10ms)  
**And** NO PostgreSQL introspection query

### Story 3.7: Implement Dataloader for N+1 Prevention

As a **platform developer**,  
I want **dataloader pattern for relation resolution**,  
So that **N+1 query problem is eliminated**.

**Acceptance Criteria:**

**Given** query requests users with posts  
**When** query is resolved  
**Then** users fetched in single query  
**And** posts batched and fetched per user (not individual queries)  
**And** total queries: 2 (users + batched posts) not N+1

**Given** deeply nested relations  
**When** dataloader is used  
**Then** query count is minimized  
**And** latency remains <100ms p95

## Epic 4: Kubernetes Deployment & Scaling

**Goal:** Users peuvent déployer Kapok en production sur Kubernetes en une commande avec auto-scaling, TLS automatique, et multi-cloud support.

### Story 4.1: Implement Helm Chart Generator

As a **platform operator**,  
I want **Helm charts auto-generated for all services**,  
So that **deployment is standardized and repeatable**.

**Acceptance Criteria:**

**Given** Kapok services exist  
**When** Helm chart generator runs  
**Then** charts created for: control-plane, graphql-engine, provisioner  
**And** umbrella chart `kapok-platform` includes all dependencies  
**And** values.yaml supports configuration overrides  
**And** templates include: Deployment, Service, HPA, Ingress, ConfigMap

### Story 4.2: Implement `kapok deploy` Command

As a **platform operator**,  
I want **one-command deployment to Kubernetes**,  
So that **production deployment is simple and fast**.

**Acceptance Criteria:**

**Given** Kubernetes cluster is accessible  
**When** `kapok deploy` is executed  
**Then** cloud provider is auto-detected (AWS/GCP/Azure)  
**And** Helm charts are generated with cloud-specific optimizations  
**And** `helm install` is executed  
**And** deployment status is monitored  
**And** process completes when all pods are ready  
**And** total time <5 minutes

### Story 4.3: Implement HPA Auto-Scaling

As a **platform operator**,  
I want **Horizontal Pod Autoscaler configured**,  
So that **services scale automatically based on load**.

**Acceptance Criteria:**

**Given** services are deployed  
**When** HPA is configured  
**Then** min replicas: 2, max replicas: 50  
**And** scaling triggers: CPU >70%, Memory >80%  
**And** scale-up: add pods when threshold exceeded  
**And** scale-down: remove pods when under-utilized  
**And** metrics scraped every 15s

### Story 4.4: Implement KEDA Custom Metrics Scaling

As a **platform operator**,  
I want **KEDA for custom metrics auto-scaling**,  
So that **scaling is based on tenant-specific metrics**.

**Acceptance Criteria:**

**Given** KEDA is installed  
**When** ScaledObject is created  
**Then** scaling triggers include: PostgreSQL connections, query latency p95, active tenants  
**And** scale-up when connections >80 per instance  
**And** scale-up when latency p95 >200ms  
**And** scale-down gracefully when metrics drop

### Story 4.5: Implement TLS/SSL with Let's Encrypt

As a **platform operator**,  
I want **automatic TLS certificate provisioning**,  
So that **HTTPS is enabled without manual configuration**.

**Acceptance Criteria:**

**Given** Ingress is configured  
**When** deployment completes  
**Then** cert-manager is installed  
**And** Let's Encrypt ClusterIssuer is created  
**And** TLS certificate is requested and obtained  
**And** certificate auto-renewal is configured (30 days before expiry)  
**And** all HTTP traffic redirects to HTTPS

### Story 4.6: Implement Multi-Cloud Support

As a **platform operator**,  
I want **AWS, GCP, and Azure deployment support**,  
So that **users can deploy to their preferred cloud**.

**Acceptance Criteria:**

**Given** `kapok deploy` detects cloud provider  
**When** deploying to AWS EKS  
**Then** EBS storage class is used, ALB ingress controller configured  
**When** deploying to GCP GKE  
**Then** GCE Persistent Disk used, GCE ingress configured  
**When** deploying to Azure AKS  
**Then** Azure Disk used, Azure ingress configured

## Epic 5: Observability & Monitoring

**Goal:** Operators ont visibilité complète sur platform health, per-tenant metrics, distributed tracing, et peuvent troubleshoot efficacement.

### Story 5.1: Auto-Deploy Prometheus Stack

As a **platform operator**,  
I want **Prometheus auto-deployed with platform**,  
So that **metrics collection is ready out-of-box**.

**Acceptance Criteria:**

**Given** `kapok deploy` runs  
**When** observability flag is enabled (default: true)  
**Then** Prometheus is deployed via Helm  
**And** scrape interval is 15s  
**And** retention period is 30 days  
**And** all Kapok services are scraped automatically

### Story 5.2: Auto-Deploy Grafana with Dashboards

As a **platform operator**,  
I want **Grafana with pre-configured dashboards**,  
So that **I can visualize metrics immediately**.

**Acceptance Criteria:**

**Given** Grafana is deployed  
**When** accessing Grafana UI  
**Then** dashboards exist for: Platform Overview, Per-Tenant Metrics, GraphQL Performance, Infrastructure Health  
**And** datasources (Prometheus, Loki) are pre-configured  
**And** default credentials are provided in deployment output

### Story 5.3: Implement OpenTelemetry Integration

As a **platform developer**,  
I want **OpenTelemetry for distributed tracing**,  
So that **requests are traced across all services**.

**Acceptance Criteria:**

**Given** OpenTelemetry SDK is integrated  
**When** request enters system  
**Then** trace ID is generated  
**And** propagated via context across services  
**And** spans created for: HTTP requests, GraphQL queries, DB queries  
**And** traces exported to Jaeger  
**And** sampling rate is 10%

### Story 5.4: Implement Per-Tenant Metrics

As a **platform operator**,  
I want **metrics isolated per tenant**,  
So that **I can monitor individual tenant health**.

**Acceptance Criteria:**

**Given** metrics are collected  
**When** Prometheus scrapes  
**Then** all metrics include `tenant_id` label  
**And** queries can filter by tenant  
**And** per-tenant metrics: CPU, memory, query latency, error rate, storage usage

### Story 5.5: Implement Alert Manager Integration

As a **platform operator**,  
I want **tiered alerting (Critical/Warning/Info)**,  
So that **I'm notified of issues appropriately**.

**Acceptance Criteria:**

**Given** AlertManager is configured  
**When** alert triggers  
**Then** Critical alerts → PagerDuty  
**And** Warning alerts → Slack  
**And** Info alerts → Metrics only (no notification)  
**And** alerts include: tenant_id, severity, description, runbook link

## Epic 6: Web Console UI

**Goal:** Users peuvent gérer tenants, visualiser metrics/logs, explorer GraphQL API, et configurer platform via interface web.

### Story 6.1: Create Web Console Foundation (Next.js)

As a **frontend developer**,  
I want **Next.js web console application**,  
So that **UI can be developed with modern framework**.

**Acceptance Criteria:**

**Given** web console project is initialized  
**When** setup completes  
**Then** Next.js 14+ with App Router  
**And** TypeScript configured  
**And** Tailwind CSS for styling  
**And** Authentication via JWT (same as CLI)  
**And** API client generated from backend GraphQL

### Story 6.2: Implement Dashboard Overview

As a **platform operator**,  
I want **dashboard showing all tenants**,  
So that **I can see platform status at a glance**.

**Acceptance Criteria:**

**Given** user is logged in  
**When** dashboard is accessed  
**Then** displays: total tenants, active tenants, total storage, total queries/day  
**And** shows list of all tenants with: name, created_at, status, storage, last_activity  
**And** real-time updates (auto-refresh every 30s)

### Story 6.3: Implement Tenant Management UI

As a **platform operator**,  
I want **create and manage tenants via UI**,  
So that **I don't need CLI for basic operations**.

**Acceptance Criteria:**

**Given** dashboard is open  
**When** "Create Tenant" button is clicked  
**Then** modal appears with form: name, isolation_level  
**And** after submission, tenant is created  
**And** success notification shown  
**And** tenant appears in list immediately

**Given** tenant in list  
**When** delete button clicked  
**Then** confirmation dialog appears  
**And** after confirmation, tenant is deleted  
**And** audit log is recorded

### Story 6.4: Implement GraphQL Playground Integration

As a **developer**,  
I want **GraphQL Playground embedded in console**,  
So that **I can test queries without external tools**.

**Acceptance Criteria:**

**Given** tenant is selected  
**When** "GraphQL Playground" tab is opened  
**Then** GraphiQL interface is displayed  
**And** schema explorer shows all types  
**And** autocomplete works for queries  
**And** query history is saved  
**And** auth token is automatically included

### Story 6.5: Implement Metrics Visualization

As a **platform operator**,  
I want **metrics visualized in console**,  
So that **I don't need separate Grafana access**.

**Acceptance Criteria:**

**Given** tenant is selected  
**When** "Metrics" tab is opened  
**Then** charts display: query latency (p50/p95/p99), error rate, throughput  
**And** time range selector: 1h, 24h, 7d, 30d  
**And** data refreshes every 15s  
**And** charts use Chart.js or similar library


## Epic 7: Backup, Recovery & High Availability

**Goal:** Platform garantit data safety avec backups automatiques multi-tier, point-in-time recovery, et disaster recovery capabilities.

**Story Outline (Detailed stories to be created when epic is prioritized):**

- Implement automated backup scheduling (daily full + hourly incremental)
- Implement point-in-time recovery mechanism
- Implement AES-256 backup encryption
- Configure immutable storage (WORM)
- Setup multi-region replication (Enterprise tier)
- Implement automated failover (Enterprise tier)
- Create backup restoration testing pipeline

## Epic 8: Security & Compliance Foundations

**Goal:** Platform respecte standards enterprise security et établit foundations pour compliance certifications.

**Story Outline (Detailed stories to be created when epic is prioritized):**

- Implement OWASP Top 10 mitigations
- Configure encryption at-rest and in-transit
- Implement MFA support (TOTP + WebAuthn)
- Setup distributed rate limiting (Redis-based)
- Implement immutable audit trail
- Create compliance documentation (ISO/SOC2/GDPR/HIPAA/FedRAMP/PCI-DSS)
- Implement security scanning CI/CD integration

## Epic 9: Advanced Multi-Tenancy (Growth Phase)

**Goal:** Platform supporte DB-per-tenant isolation avec auto-migration seamless et hierarchies parent-child.

**Story Outline (Detailed stories to be created when epic is prioritized):**

- Implement DB-per-tenant provisioning
- Implement zero-downtime migration (schema → DB)
- Create usage threshold monitoring
- Implement automated migration triggers
- Support parent-child tenant hierarchies (Organization → Project → Team)
- Implement tenant resource quotas per hierarchy level
- Create tenant migration testing framework

## Epic 10: Real-Time & WebSocket Subscriptions (Growth Phase)

**Goal:** Tenants peuvent utiliser GraphQL subscriptions real-time avec <100ms latency.

**Story Outline (Detailed stories to be created when epic is prioritized):**

- Implement PostgreSQL LISTEN/NOTIFY triggers
- Integrate GraphQL subscription resolvers
- Setup Redis Pub/Sub for fan-out
- Implement WebSocket connection management
- Create tenant-aware subscription channels
- Implement subscription authentication and authorization
- Performance test (1000+ concurrent connections, <100ms latency)

## Epic 11: Advanced RBAC & Row-Level Security (Growth Phase)

**Goal:** Tenants ont fine-grained access control avec RLS et RBAC policies complexes.

**Story Outline (Detailed stories to be created when epic is prioritized):**

- Implement PostgreSQL Row-Level Security policies
- Extend Casbin for hierarchical RBAC (Organization → Project → Team → User)
- Implement automatic SQL policy injection
- Create immutable audit trail for all access/modifications
- Implement policy testing framework
- Create compliance-ready access logs
- Performance test RLS impact on query latency

## Epic 12: Visual Schema Builder & Advanced DX (Growth Phase)

**Goal:** Users ont outils graphiques pour schema design, query testing, et performance profiling.

**Story Outline (Detailed stories to be created when epic is prioritized):**

- Create visual schema builder UI (drag-and-drop tables/columns)
- Integrate GraphQL Playground in web console
- Implement query performance profiler with recommendations
- Create time-travel query debugging feature
- Implement schema migration assistance
- Create auto-optimization suggestion engine
- Build schema versioning and rollback capability


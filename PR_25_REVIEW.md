# Code Review — PR #25: feat(api): add control-plane HTTP API server

## Issues Found

### 1. CRITICAL: Hardcoded admin password in seed
**File:** `cmd/control-plane/main.go:141`

The admin password is hardcoded to `"admin"`. This should be read from an environment variable (e.g. `KAPOK_ADMIN_PASSWORD`) with a loud warning if using the default.

### 2. CRITICAL: No RBAC on admin routes
**File:** `internal/api/router.go:52-64`

All authenticated routes share the same middleware. Any authenticated user (including `role=user`) can access `/api/v1/admin/*`, create/delete tenants, etc. Add a `RequireRole("admin")` middleware on the admin group.

### 3. HIGH: Silently swallowed DB errors in Stats handler
**File:** `internal/api/handlers_admin.go:12-16`

Four DB queries with errors discarded via `_`. If `audit_log` table doesn't exist, the handler silently returns 0. At minimum, log the errors.

### 4. MEDIUM: Unused jwt import kept via dummy variable
**File:** `internal/api/router.go:87`

```go
var _ jwt.MapClaims
```

The `jwt` package isn't actually used in this file. The type cast `map[string]interface{}(claims)` works without importing jwt. Remove the import and dummy var.

### 5. MEDIUM: Slug generation is too naive
**File:** `internal/tenant/provisioner.go:50`

```go
slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
```

Doesn't handle special characters or duplicates. `"My Tenant!!!"` produces `"my-tenant!!!"`. Sanitize to `[a-z0-9-]` and add a uniqueness constraint.

### 6. MEDIUM: E2E test asserts fields the API doesn't return
**File:** `tests/e2e/control_plane_test.sh:178-179`

The test checks for `range` and `metrics` fields, but the Metrics handler returns `query_latency_p50`, `throughput`, etc. These two assertions will fail.

### 7. MEDIUM: No request body size limit
**File:** `internal/api/helpers.go:13-15`

`readJSON` doesn't use `http.MaxBytesReader` — any endpoint can receive unbounded request bodies. Add a size limit.

### 8. LOW: CORS origins hardcoded
**File:** `internal/api/router.go:36`

Should be configurable via environment variable for production deployments.

### 9. LOW: Ad-hoc migrations bypass existing system
**File:** `cmd/control-plane/main.go:98-103`

`createUsersTable` and `extendTenantsTable` use raw DDL in main.go, bypassing `database.NewMigrator`. These should be proper migration files to avoid schema drift.

## What looks good

- Clean dependency injection via `Dependencies` struct
- Graceful shutdown handling with signal management
- Proper use of `context.WithValue` with typed key (`contextKeyType`)
- Good E2E test structure covering the full API surface
- Correct use of `COALESCE` in queries for nullable new columns
- Chi router is a solid choice matching the project's Go ecosystem

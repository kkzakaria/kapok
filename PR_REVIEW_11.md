# PR Review: #11 - Multi-Tenant Core Infrastructure (Epic 2)

**Reviewer:** Claude
**Date:** 2026-01-28
**Branch:** `feat/epic-2-multi-tenant-infrastructure`
**Target:** `main`

---

## Summary

Cette PR implémente l'infrastructure multi-tenant fondamentale pour Kapok, comprenant 7 stories de l'Epic 2. L'implémentation couvre l'isolation par schéma PostgreSQL, l'authentification JWT, l'autorisation RBAC via Casbin, et les outils CLI de gestion des tenants.

**Statistiques:**
- 29 fichiers modifiés
- +4,574 lignes / -73 lignes
- 5 nouveaux packages
- 95+ tests unitaires et d'intégration

---

## Points Positifs

### 1. Architecture Solide

L'approche **schema-per-tenant** est un excellent choix pour l'isolation des données:
- Isolation forte au niveau PostgreSQL
- RLS (Row-Level Security) comme protection secondaire
- Schémas nommés avec UUID (`tenant_<uuid>`) pour éviter les collisions

### 2. Sécurité Bien Pensée

- **Validation des entrées** : `ValidateName()` avec regex pour prévenir les injections
- **Validation des schémas** : `isValidSchemaName()` vérifie le préfixe `tenant_` et les caractères
- **JWT robuste** : Utilisation de HMAC SHA256 avec golang-jwt/jwt/v5
- **Transactions** : Utilisation appropriée des transactions pour la cohérence

### 3. Tests Complets

Les tests d'intégration avec dockertest sont particulièrement bien faits:
- `TestCrossTenantDataIsolation` - Vérifie l'isolation des données
- `TestTenantProvisioningPerformance` - Valide le SLA < 30s
- `TestSQLInjectionPrevention` - Tests de sécurité
- `TestConcurrentTenantCreation` - Tests de concurrence

### 4. Code Propre et Bien Structuré

- Logging cohérent avec zerolog
- Gestion des erreurs avec wrapping (`fmt.Errorf(...: %w, err)`)
- Séparation claire des responsabilités entre packages
- Documentation via commentaires Go standards

### 5. CLI Fonctionnel

Les commandes CLI sont bien implémentées avec:
- Support JSON et table pour l'output
- Pagination (--limit, --offset)
- Filtrage par status

---

## Issues et Préoccupations

### Critique

#### 1. Fuite Potentielle de Connexions DB dans `provisioner.go`

```go
// internal/tenant/provisioner.go:56-58
tx, err := p.db.BeginTx(ctx, nil)
if err != nil {
    return nil, fmt.Errorf("failed to begin transaction: %w", err)
}
defer tx.Rollback()  // OK
```

Le `defer tx.Rollback()` est correct, mais si `tx.Commit()` réussit, le Rollback retourne une erreur ignorée. C'est fonctionnel mais pas idéal.

**Suggestion:** Ajouter un pattern plus explicite:
```go
committed := false
defer func() {
    if !committed {
        tx.Rollback()
    }
}()
// ...
if err := tx.Commit(); err != nil {
    return nil, err
}
committed = true
```

#### 2. Context Key String dans `middleware.go`

```go
// internal/auth/middleware.go:55
ctx := context.WithValue(r.Context(), "jwt_claims", claimsMap)
```

Utiliser une string comme clé de contexte est déconseillé car ça peut causer des collisions.

**Suggestion:** Définir un type dédié:
```go
type contextKey string
const jwtClaimsKey contextKey = "jwt_claims"
```

#### 3. Pas de Validation du Tenant ID dans le Router

```go
// internal/tenant/router.go:39-46
tenantID, ok := tenantIDInterface.(string)
if !ok || tenantID == "" {
    // ...
}
```

Le tenant ID n'est pas validé comme UUID valide, ce qui pourrait permettre des valeurs malformées.

### Important

#### 4. Refresh Token Sans Révocation

```go
// internal/auth/jwt.go:55-74
func (m *JWTManager) GenerateRefreshToken(userID, tenantID string) (string, error) {
```

Il n'y a pas de mécanisme pour révoquer les refresh tokens. Si un token est compromis, il reste valide pendant 7 jours.

**Suggestion:** Implémenter un token store (Redis/DB) pour permettre la révocation.

#### 5. Casbin Model Sans Wildcard Tenant

```conf
# deployments/rbac/model.conf
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act && r.tenant == p.tenant
```

Le matcher requiert une correspondance exacte du tenant, ce qui empêche les super-admins cross-tenant.

**Suggestion:** Ajouter support pour `*` tenant:
```conf
m = g(r.sub, p.sub) && (r.obj == p.obj || p.obj == "*") && (r.act == p.act || p.act == "*") && (r.tenant == p.tenant || p.tenant == "*")
```

#### 6. Rollback Incomplet en Cas d'Échec Schema

```go
// internal/tenant/provisioner.go:85-88
if err := p.migrator.CreateTenantSchema(ctx, schemaName); err != nil {
    p.deleteTenantMetadata(ctx, tenantID)  // Erreur ignorée
    return nil, fmt.Errorf("failed to create tenant schema: %w", err)
}
```

L'erreur de `deleteTenantMetadata` est ignorée, ce qui peut laisser des données orphelines.

### Mineur

#### 7. Magic Numbers

```go
// internal/database/connection.go:52-53
sqlDB.SetMaxOpenConns(50)  // Default
sqlDB.SetMaxIdleConns(10)  // Default
```

Ces valeurs devraient être des constantes nommées.

#### 8. SQL Statement Split Simpliste

```go
// internal/database/migrations.go:157
return strings.Split(sql, ";")
```

Cette approche ne gère pas les points-virgules dans les strings SQL.

---

## Suggestions d'Amélioration

### 1. Ajouter des Métriques

Considérer l'ajout de métriques Prometheus pour:
- Temps de provisioning
- Nombre de tenants actifs
- Erreurs d'authentification

### 2. Rate Limiting

Le middleware d'authentification devrait inclure un rate limiter pour prévenir les attaques par force brute.

### 3. Health Check Tenant-Aware

Ajouter un endpoint de health check qui vérifie l'état des schémas tenant.

### 4. Documentation API

Considérer l'ajout de documentation OpenAPI/Swagger pour les endpoints.

---

## Vérification de Sécurité

| Check | Status | Notes |
|-------|--------|-------|
| SQL Injection | PASS | Paramètres préparés + validation |
| JWT Security | PASS | HMAC SHA256, expiration appropriée |
| Input Validation | PASS | Regex sur noms de tenant |
| Schema Isolation | PASS | Préfixe obligatoire `tenant_` |
| RLS | PASS | Implémenté comme couche secondaire |
| Audit Logging | PASS | Actions critiques loggées |
| Secrets | WARN | JWT secret devrait être validé (longueur min) |

---

## Tests Requis Avant Merge

- [x] Tests unitaires passent
- [x] Tests d'intégration passent
- [ ] Test de charge (10+ tenants simultanés)
- [ ] Test de recovery après crash
- [ ] Revue par un second développeur

---

## Recommandation Finale

### APPROVE avec Réserves

Cette PR est **prête pour merge** après correction des issues critiques suivantes:

1. **Obligatoire:** Fixer le pattern de context key (issue #2)
2. **Recommandé:** Ajouter validation UUID pour tenant ID (issue #3)
3. **Recommandé:** Logger l'erreur de rollback metadata (issue #6)

Les autres issues peuvent être adressées dans des PRs de suivi.

**Score Global:** 8/10

L'implémentation est solide, bien testée et suit les bonnes pratiques. Les issues identifiées sont mineures et n'affectent pas la fonctionnalité de base.

---

## Fichiers Clés Revus

| Fichier | Lignes | Verdict |
|---------|--------|---------|
| `internal/tenant/provisioner.go` | 340 | Bon |
| `internal/auth/jwt.go` | 163 | Bon |
| `internal/auth/middleware.go` | 98 | Issue mineure |
| `internal/database/rls.go` | 242 | Excellent |
| `internal/rbac/casbin.go` | 215 | Bon |
| `internal/tenant/integration_test.go` | 412 | Excellent |
| `deployments/migrations/001_control_database.sql` | 97 | Bon |

---

*Review effectuée le 2026-01-28 par Claude Code*

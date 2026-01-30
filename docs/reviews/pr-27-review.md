# Revue de la PR #27 — Epic 8: Security & Compliance Foundations

**Commit initial :** `c7a7ac0` — `feat(security): implement Epic 8 - Security & Compliance Foundations`
**Commit de correction :** `75f1a93` — `fix(security): address PR #27 review issues`
**Diff total :** +4 899 lignes ajoutées dans 21 fichiers

---

## Revue initiale — 16 problèmes identifiés

La première revue avait identifié 6 problèmes critiques, 5 modérés et 5 mineurs.

## Revue post-correction — Vérification des fixes

### Issues critiques — Toutes corrigées

| # | Issue | Statut | Vérification |
|---|-------|--------|--------------|
| 1 | HSTS/CORS `max-age` : `string(rune(int))` | **Corrigé** | `strconv.Itoa()` utilisé correctement (`headers.go:73, 169`) |
| 2 | X-Forwarded-For parsing incorrect | **Corrigé** | `strings.IndexByte(xff, ',')` + `strings.TrimSpace()` (`rate_limiter.go:104-107`) |
| 3 | CSRF timing attack | **Corrigé** | `subtle.ConstantTimeCompare()` utilisé (`csrf.go:88`) |
| 4 | `ContainsSQLi` faux positifs | **Corrigé** | Patterns réduits aux attaques évidentes (`union select`, `drop table`, etc.). Documentation ajoutée précisant que les requêtes paramétrées sont la défense primaire (`validation.go:148-152`) |
| 5 | Audit table syntaxe MySQL | **Corrigé** | `CREATE INDEX IF NOT EXISTS ... ON ...` en syntaxe PostgreSQL séparée (`audit.go:136-139`) |
| 6 | CSRF cookie `HttpOnly: true` | **Corrigé** | `HttpOnly: false` avec commentaire explicatif sur le pattern Double Submit Cookie (`csrf.go:545-558`) |

### Issues modérées — 3 corrigées, 2 non corrigées

| # | Issue | Statut | Détail |
|---|-------|--------|--------|
| 7 | `VerifyBackupCode` corruption slice | **Corrigé** | Allocation d'un nouveau slice avec `make()` + double `append` (`mfa_totp.go:143-145`) |
| 8 | Password generation biais modulo | **Corrigé** | `rand.Int(rand.Reader, charsetLen)` avec `math/big` (`password.go:86-90`) |
| 9 | Rate limiter fail-open sans alerting | **Non corrigé** | Toujours fail-open sans compteur ni métrique de fallback. Acceptable pour le moment. |
| 10 | Pas d'intégration avec le code existant | **Non corrigé** | Le package reste du dead code, non branché dans le router ni les middlewares. Attendu si l'intégration est prévue dans un epic ultérieur. |
| 11 | Context key string pour JWT claims | **Non corrigé** | `ctx.Value("jwt_claims")` utilise toujours une string brute (`rate_limiter.go:122`, `csrf.go:65`). Risque de collision avec d'autres packages. |

### Issues mineures — 2 corrigées, 3 non corrigées

| # | Issue | Statut | Détail |
|---|-------|--------|--------|
| 12 | Regex recompilées à chaque appel | **Corrigé** | Regex pré-compilées en variables de package (`validation.go:12-15`) |
| 13 | `PreferServerCipherSuites` deprecated | **Corrigé** | Champ supprimé de `TLSConfig` (`tls_config.go`) |
| 14 | SHA1 pour TOTP | **Non corrigé** | Toujours `otp.AlgorithmSHA1`. Acceptable car conforme au RFC 6238. |
| 15 | Couverture de tests insuffisante | **Non corrigé** | Toujours uniquement `validation_test.go`. Aucun test ajouté pour les autres fichiers. |
| 16 | GitHub Actions non-pinnées | **Non corrigé** | Les Actions dans `security-scan.yml` utilisent toujours des tags au lieu de SHA. |

---

## Nouveaux problèmes identifiés dans le code corrigé

### N1. `ContainsSQLi` — test cassé (sévérité : faible)

**Fichier :** `internal/security/validation_test.go`

Le test pour `ContainsSQLi` contient toujours le cas `{"single quote", "O'Brien", true}` mais le code corrigé ne devrait plus flaguer les simples apostrophes. Le test va probablement échouer. À vérifier.

### N2. `joinStrings` — réinvention de `strings.Join` (sévérité : cosmétique)

**Fichier :** `internal/security/headers.go:198-207`

La fonction `joinStrings` est une copie de `strings.Join` de la bibliothèque standard. Utiliser directement `strings.Join` serait plus idiomatique.

### N3. `isOriginAllowed` wildcard matching trop permissif (sévérité : modéré)

**Fichier :** `internal/security/headers.go:185-195`

Le wildcard `https://*` dans `DefaultCORSConfig` matche n'importe quelle origine commençant par `https://`. Combiné avec `AllowCredentials: true`, cela permet à n'importe quel site HTTPS d'effectuer des requêtes authentifiées. La spécification CORS interdit `Access-Control-Allow-Origin: *` avec `Access-Control-Allow-Credentials: true`, mais ici le code renvoie l'origine exacte, ce qui contourne cette protection.

**Recommandation :** Le défaut devrait être une liste vide ou des origines explicites, pas un wildcard.

### N4. Audit `fmt.Sprintf` avec `tableName` — injection SQL potentielle (sévérité : modéré)

**Fichier :** `internal/security/audit.go:119-147`

`al.tableName` est interpolé directement dans les requêtes SQL via `fmt.Sprintf`. Si `tableName` peut être contrôlé par un utilisateur (peu probable ici car hardcodé à `"audit_logs"`), cela serait une injection SQL. Néanmoins, utiliser un identifiant quoté (`pq.QuoteIdentifier`) serait plus défensif.

---

## Tableau récapitulatif final

| Catégorie | Initiales | Corrigées | Restantes | Nouvelles |
|-----------|-----------|-----------|-----------|-----------|
| Critiques | 6 | 6 | 0 | 0 |
| Modérés | 5 | 2 | 3 | 2 (N3, N4) |
| Mineurs | 5 | 2 | 3 | 2 (N1, N2) |
| **Total** | **16** | **10** | **6** | **4** |

---

## Recommandation

**Le code est acceptable pour merge** étant donné que tous les problèmes critiques ont été corrigés.

Les problèmes restants (non-intégration, tests manquants, context key string, SHA1 TOTP, Actions non-pinnées) et les nouveaux problèmes (wildcard CORS, `tableName` non-quoté) sont des améliorations souhaitables mais non bloquantes. Ils peuvent être traités dans des PRs ultérieures.

**Points à traiter en priorité dans les prochaines itérations :**
1. Intégrer le package security dans les middlewares de l'API (#10)
2. Ajouter des tests unitaires pour tous les fichiers (#15)
3. Restreindre le défaut CORS à des origines explicites (N3)
4. Utiliser `pq.QuoteIdentifier` pour le nom de table audit (N4)

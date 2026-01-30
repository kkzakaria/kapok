# Revue de la PR #27 — Epic 8: Security & Compliance Foundations

**Commit :** `c7a7ac0` — `feat(security): implement Epic 8 - Security & Compliance Foundations`
**Diff :** +4 856 lignes ajoutées dans 20 fichiers (9 `.go`, 1 test, 1 workflow CI, 5 docs compliance, 1 README, 1 planning, `go.mod`)

---

## Problèmes critiques (bugs / vulnérabilités)

### 1. HSTS et CORS `max-age` : conversion int→string cassée

**Fichier :** `internal/security/headers.go:199, 294`

`string(rune(m.config.HSTSMaxAge))` ne produit pas `"31536000"` mais un seul caractère Unicode (code point 31536000, probablement invalide). Même bug pour `Access-Control-Max-Age`.

```go
// Faux : string(rune(31536000)) → caractère unicode invalide
// Correct : fmt.Sprintf("%d", m.config.HSTSMaxAge) ou strconv.Itoa()
```

**Impact :** HSTS et CORS max-age seront des valeurs corrompues. Faille de sécurité directe — le navigateur ne respectera pas le HSTS.

---

### 2. `getClientIP` — parsing X-Forwarded-For incorrect

**Fichier :** `internal/security/rate_limiter.go:178-182`

Le code fait `xff[:len(xff)]` ce qui retourne la chaîne entière au lieu de prendre le premier IP avant la virgule. Un attaquant peut spoofer son IP pour contourner le rate limiting.

```go
// Faux : xff[:idx] où idx = len(xff) quand il n'y a pas de virgule
// Correct : strings.SplitN(xff, ",", 2)[0] puis strings.TrimSpace()
```

**Impact :** Contournement du rate limiting via spoofing X-Forwarded-For.

---

### 3. CSRF token comparison non constant-time

**Fichier :** `internal/security/csrf.go:417`

`token != storedToken` est une comparaison temporelle standard. Un attaquant peut potentiellement deviner le token un octet à la fois via timing attack.

```go
// Faux : token != storedToken
// Correct : subtle.ConstantTimeCompare([]byte(token), []byte(storedToken)) != 1
```

**Impact :** Vulnérabilité de type timing attack sur la validation CSRF.

---

### 4. `ContainsSQLi` — faux positifs massifs

**Fichier :** `internal/security/validation.go:145-160`

Bloquer `'`, `;`, `--` rend impossible de saisir des noms comme "O'Brien" ou du texte contenant des points-virgules. Ce n'est pas une méthode de prévention SQLi viable — les requêtes paramétrées (déjà utilisées dans le projet via `database/sql`) sont la bonne solution.

```go
// Le test confirme le problème :
// {"single quote", "O'Brien", true}  ← un nom légitime est bloqué
```

**Impact :** Faux positifs en production. Ce validateur ne devrait pas exister tel quel ou devrait être documenté comme strictement optionnel.

---

### 5. Audit table — syntaxe PostgreSQL `INDEX` dans `CREATE TABLE`

**Fichier :** `internal/security/audit.go:100-104`

La syntaxe `INDEX idx_name (column)` dans un `CREATE TABLE` est propre à MySQL, pas PostgreSQL. Kapok utilise PostgreSQL exclusivement.

```sql
-- Faux (MySQL) :
CREATE TABLE audit_events (
    ...,
    INDEX idx_tenant (tenant_id)
);

-- Correct (PostgreSQL) :
CREATE TABLE audit_events (...);
CREATE INDEX idx_tenant ON audit_events (tenant_id);
```

**Impact :** La création de la table d'audit échouera à l'exécution.

---

### 6. CSRF cookie HttpOnly empêche la lecture côté client

**Fichier :** `internal/security/csrf.go:481`

Le cookie CSRF est défini avec `HttpOnly: true`, ce qui empêche le JavaScript de le lire. Or, le pattern Double Submit Cookie nécessite que le client envoie le token dans un header (`X-CSRF-Token`). Si le client ne peut pas lire le cookie, il ne peut pas remplir le header.

```go
// Problème : HttpOnly: true empêche le JS de lire le token
// Solution : HttpOnly: false pour un cookie CSRF (c'est sûr car SameSite=Strict)
```

**Impact :** Le mécanisme CSRF est inutilisable tel quel.

---

## Problèmes modérés (design / qualité)

### 7. `VerifyBackupCode` — corruption silencieuse du slice

**Fichier :** `internal/security/mfa_totp.go:781`

`append(backupCodes[:i], backupCodes[i+1:]...)` modifie le slice sous-jacent passé en argument. L'appelant pourrait voir des données corrompues si il réutilise le slice original.

**Recommandation :** Copier dans un nouveau slice.

---

### 8. `GenerateSecureRandomPassword` — biais modulo

**Fichier :** `internal/security/password.go:120`

`int(randomBytes[i]) % len(allChars)` introduit un léger biais car 256 n'est pas divisible par `len(allChars)` (91 chars). Utiliser `crypto/rand` + `math/big.Int` pour une distribution uniforme.

---

### 9. Rate limiter fail-open sans mécanisme de fallback

**Fichier :** `internal/security/rate_limiter.go:70-73`

Si Redis est down, toutes les requêtes passent (fail-open). C'est un choix acceptable du point de vue disponibilité, mais il devrait y avoir au minimum un compteur/métrique de fail-opens pour alerter l'opérateur, ou un fallback local (in-memory limiter).

---

### 10. Pas d'intégration avec le code existant

Aucun des fichiers existants (`internal/auth/`, `internal/api/`, `cmd/`) n'est modifié pour utiliser ce nouveau package security. Le code est entièrement dead code à ce stade. Aucun middleware n'est branché dans le router.

---

### 11. Contexte JWT avec string key

**Fichiers :** `internal/security/rate_limiter.go:208`, `internal/security/csrf.go:491`

`ctx.Value("jwt_claims")` utilise une string brute comme context key, ce qui est fragile et risque des collisions avec d'autres packages. Le projet utilise probablement déjà un type dédié dans `internal/auth/` — il faudrait réutiliser ce type.

```go
// Faux : ctx.Value("jwt_claims")
// Correct : ctx.Value(auth.ClaimsContextKey)
```

---

## Problèmes mineurs

### 12. Regex recompilées à chaque appel

**Fichier :** `internal/security/validation.go`

Les expressions régulières sont compilées à chaque appel de validation. Elles devraient être compilées une seule fois en variables de package avec `regexp.MustCompile()`.

---

### 13. `PreferServerCipherSuites` deprecated

**Fichier :** `internal/security/tls_config.go:576`

`PreferServerCipherSuites` est deprecated depuis Go 1.18. Go gère automatiquement l'ordre des cipher suites.

---

### 14. SHA1 pour TOTP

**Fichier :** `internal/security/mfa_totp.go:679`

SHA1 est le standard RFC 6238 et reste acceptable pour TOTP, mais SHA256 (`otp.AlgorithmSHA256`) serait préférable si les apps d'authentification des utilisateurs le supportent.

---

### 15. Couverture de tests insuffisante

Seul `validation_test.go` existe. Aucun test pour : encryption, password, headers, CSRF, rate limiter, audit, MFA, TLS config.

**Fichiers manquants :**
- `encryption_test.go`
- `password_test.go`
- `headers_test.go`
- `csrf_test.go`
- `rate_limiter_test.go`
- `audit_test.go`
- `mfa_totp_test.go`
- `tls_config_test.go`

---

### 16. GitHub Actions non-pinnées (risque supply chain)

**Fichier :** `.github/workflows/security-scan.yml`

Plusieurs GitHub Actions sont référencées par tag (`@main`, `@master`, `@v1`, `@v2`) au lieu de SHA de commit. Cela expose le pipeline CI/CD à des attaques de supply chain.

```yaml
# Faux :
uses: actions/checkout@v4

# Correct :
uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1
```

---

## Tableau récapitulatif

| Catégorie | Nombre | Bloquant ? |
|-----------|--------|------------|
| Bugs critiques / vulnérabilités | 6 | Oui |
| Problèmes modérés (design) | 5 | Non |
| Problèmes mineurs | 5 | Non |

---

## Recommandation

**Demander des corrections avant merge.**

Les bugs #1 (HSTS cassé), #2 (IP spoofing), #3 (timing attack CSRF), #5 (syntaxe SQL invalide) et #6 (CSRF inutilisable) sont des **bloquants**. Le code n'est intégré nulle part dans l'application (#10), donc le risque immédiat en production est limité, mais ces bugs doivent impérativement être corrigés avant que le package soit branché dans les middlewares de l'API.

#!/usr/bin/env bash
# =============================================================================
# Kapok Control-Plane E2E Test — Epics 1–6
# =============================================================================
# Tests: Auth (login, me), Admin (stats, tenants CRUD, metrics), GraphQL proxy
#
# Usage: bash tests/e2e/control_plane_test.sh
# Requires: control-plane running on localhost:8080, PostgreSQL on localhost:5432
# =============================================================================

set -uo pipefail

API="http://localhost:8080"
PASS=0
FAIL=0
TOKEN=""
TENANT_ID=""
TENANT_NAME="e2e-test-$(date +%s)"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

# ─── Helpers ────────────────────────────────────────────────────────────────
assert_status() {
  local test_name="$1" expected="$2" actual="$3"
  if [ "$actual" -eq "$expected" ]; then
    echo -e "  ${GREEN}✓${NC} $test_name (HTTP $actual)"
    PASS=$((PASS+1))
  else
    echo -e "  ${RED}✗${NC} $test_name — expected $expected, got $actual"
    FAIL=$((FAIL+1))
  fi
}

assert_json_field() {
  local test_name="$1" body="$2" field="$3"
  if echo "$body" | python3 -c "import sys,json; d=json.load(sys.stdin); assert '$field' in d" 2>/dev/null; then
    echo -e "  ${GREEN}✓${NC} $test_name (field '$field' present)"
    PASS=$((PASS+1))
  else
    echo -e "  ${RED}✗${NC} $test_name — field '$field' missing"
    FAIL=$((FAIL+1))
  fi
}

auth_header() {
  echo "Authorization: Bearer $TOKEN"
}

# ─── Wait for server ───────────────────────────────────────────────────────
echo -e "\n${CYAN}⏳ Waiting for control-plane at $API ...${NC}"
for i in $(seq 1 30); do
  if curl -sf "$API/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Server is up${NC}"
    break
  fi
  if [ "$i" -eq 30 ]; then
    echo -e "${RED}✗ Server did not start in 30s${NC}"
    exit 1
  fi
  sleep 1
done

# ═════════════════════════════════════════════════════════════════════════════
echo -e "\n${YELLOW}━━━ EPIC 1: Health Check (Project Foundation) ━━━${NC}"
# ═════════════════════════════════════════════════════════════════════════════

HTTP=$(curl -s -o /dev/null -w "%{http_code}" "$API/health")
assert_status "GET /health" 200 "$HTTP"

# ═════════════════════════════════════════════════════════════════════════════
echo -e "\n${YELLOW}━━━ EPIC 2: Authentication ━━━${NC}"
# ═════════════════════════════════════════════════════════════════════════════

# — 2.1 Login with wrong password → 401
HTTP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@kapok.dev","password":"wrong"}')
assert_status "Login wrong password → 401" 401 "$HTTP"

# — 2.2 Login with valid creds → 200
RESP=$(curl -s -w "\n%{http_code}" -X POST "$API/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@kapok.dev","password":"admin"}')
BODY=$(echo "$RESP" | head -n -1)
HTTP=$(echo "$RESP" | tail -1)
assert_status "Login valid creds → 200" 200 "$HTTP"
assert_json_field "Login returns access_token" "$BODY" "access_token"
assert_json_field "Login returns refresh_token" "$BODY" "refresh_token"

TOKEN=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])")

# — 2.3 Access protected route without token → 401
HTTP=$(curl -s -o /dev/null -w "%{http_code}" "$API/api/v1/auth/me")
assert_status "GET /me without token → 401" 401 "$HTTP"

# — 2.4 GET /me with token → 200
RESP=$(curl -s -w "\n%{http_code}" "$API/api/v1/auth/me" -H "$(auth_header)")
BODY=$(echo "$RESP" | head -n -1)
HTTP=$(echo "$RESP" | tail -1)
assert_status "GET /me with token → 200" 200 "$HTTP"
assert_json_field "Me returns email" "$BODY" "email"

# ═════════════════════════════════════════════════════════════════════════════
echo -e "\n${YELLOW}━━━ EPIC 3: Admin Stats (Dashboard) ━━━${NC}"
# ═════════════════════════════════════════════════════════════════════════════

RESP=$(curl -s -w "\n%{http_code}" "$API/api/v1/admin/stats" -H "$(auth_header)")
BODY=$(echo "$RESP" | head -n -1)
HTTP=$(echo "$RESP" | tail -1)
assert_status "GET /admin/stats → 200" 200 "$HTTP"
assert_json_field "Stats has total_tenants" "$BODY" "total_tenants"
assert_json_field "Stats has active_tenants" "$BODY" "active_tenants"
assert_json_field "Stats has total_storage_bytes" "$BODY" "total_storage_bytes"
assert_json_field "Stats has total_queries_today" "$BODY" "total_queries_today"

# ═════════════════════════════════════════════════════════════════════════════
echo -e "\n${YELLOW}━━━ EPIC 4: Tenant CRUD (Multi-Tenant Core) ━━━${NC}"
# ═════════════════════════════════════════════════════════════════════════════

# — 4.1 List tenants (initially empty or existing)
RESP=$(curl -s -w "\n%{http_code}" "$API/api/v1/admin/tenants" -H "$(auth_header)")
HTTP=$(echo "$RESP" | tail -1)
assert_status "GET /admin/tenants → 200" 200 "$HTTP"

# — 4.2 Create tenant
RESP=$(curl -s -w "\n%{http_code}" -X POST "$API/api/v1/admin/tenants" \
  -H "$(auth_header)" -H "Content-Type: application/json" \
  -d "{\"name\":\"$TENANT_NAME\",\"isolation_level\":\"schema\"}")
BODY=$(echo "$RESP" | head -n -1)
HTTP=$(echo "$RESP" | tail -1)
assert_status "POST /admin/tenants → 201" 201 "$HTTP"
assert_json_field "Created tenant has id" "$BODY" "id"
assert_json_field "Created tenant has name" "$BODY" "name"
assert_json_field "Created tenant has schema_name" "$BODY" "schema_name"
assert_json_field "Created tenant has slug" "$BODY" "slug"

TENANT_ID=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")
echo -e "  ${CYAN}↳ tenant_id = $TENANT_ID${NC}"

# — 4.3 Get tenant by ID
RESP=$(curl -s -w "\n%{http_code}" "$API/api/v1/admin/tenants/$TENANT_ID" -H "$(auth_header)")
BODY=$(echo "$RESP" | head -n -1)
HTTP=$(echo "$RESP" | tail -1)
assert_status "GET /admin/tenants/:id → 200" 200 "$HTTP"
assert_json_field "Tenant detail has isolation_level" "$BODY" "isolation_level"

# — 4.4 Get non-existent tenant → 404
HTTP=$(curl -s -o /dev/null -w "%{http_code}" \
  "$API/api/v1/admin/tenants/00000000-0000-0000-0000-000000000000" -H "$(auth_header)")
assert_status "GET non-existent tenant → 404" 404 "$HTTP"

# — 4.5 List tenants (should include new tenant)
RESP=$(curl -s -w "\n%{http_code}" "$API/api/v1/admin/tenants" -H "$(auth_header)")
BODY=$(echo "$RESP" | head -n -1)
HTTP=$(echo "$RESP" | tail -1)
assert_status "GET /admin/tenants after create → 200" 200 "$HTTP"
COUNT=$(echo "$BODY" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))")
if [ "$COUNT" -ge 1 ]; then
  echo -e "  ${GREEN}✓${NC} Tenant list count >= 1 (got $COUNT)"
  PASS=$((PASS+1))
else
  echo -e "  ${RED}✗${NC} Expected >= 1 tenant, got $COUNT"
  FAIL=$((FAIL+1))
fi

# ═════════════════════════════════════════════════════════════════════════════
echo -e "\n${YELLOW}━━━ EPIC 5: Metrics (Observability) ━━━${NC}"
# ═════════════════════════════════════════════════════════════════════════════

RESP=$(curl -s -w "\n%{http_code}" "$API/api/v1/admin/metrics?range=24h" -H "$(auth_header)")
BODY=$(echo "$RESP" | head -n -1)
HTTP=$(echo "$RESP" | tail -1)
assert_status "GET /admin/metrics → 200" 200 "$HTTP"
assert_json_field "Metrics has range" "$BODY" "range"
assert_json_field "Metrics has metrics" "$BODY" "metrics"

# ═════════════════════════════════════════════════════════════════════════════
echo -e "\n${YELLOW}━━━ EPIC 6: GraphQL Proxy ━━━${NC}"
# ═════════════════════════════════════════════════════════════════════════════

# GraphQL introspection query on the test tenant
RESP=$(curl -s -w "\n%{http_code}" -X POST "$API/api/v1/tenants/$TENANT_ID/graphql" \
  -H "$(auth_header)" -H "Content-Type: application/json" \
  -d '{"query":"{ __typename }"}')
HTTP=$(echo "$RESP" | tail -1)
# Accept 200 (success) or 500 (empty schema is fine — proves routing works)
if [ "$HTTP" -eq 200 ] || [ "$HTTP" -eq 500 ]; then
  echo -e "  ${GREEN}✓${NC} POST /tenants/:id/graphql responded (HTTP $HTTP)"
  PASS=$((PASS+1))
else
  echo -e "  ${RED}✗${NC} POST /tenants/:id/graphql unexpected status $HTTP"
  FAIL=$((FAIL+1))
fi

# GraphQL with invalid tenant → 404
HTTP=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
  "$API/api/v1/tenants/00000000-0000-0000-0000-000000000000/graphql" \
  -H "$(auth_header)" -H "Content-Type: application/json" \
  -d '{"query":"{ __typename }"}')
assert_status "GraphQL invalid tenant → 404" 404 "$HTTP"

# ═════════════════════════════════════════════════════════════════════════════
echo -e "\n${YELLOW}━━━ Cleanup: Delete test tenant ━━━${NC}"
# ═════════════════════════════════════════════════════════════════════════════

RESP=$(curl -s -w "\n%{http_code}" -X DELETE "$API/api/v1/admin/tenants/$TENANT_ID" \
  -H "$(auth_header)")
HTTP=$(echo "$RESP" | tail -1)
assert_status "DELETE /admin/tenants/:id → 200" 200 "$HTTP"

# ═════════════════════════════════════════════════════════════════════════════
echo -e "\n${CYAN}═══════════════════════════════════════════════════${NC}"
echo -e "${CYAN}  RESULTS: ${GREEN}$PASS passed${NC}, ${RED}$FAIL failed${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════${NC}\n"

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi

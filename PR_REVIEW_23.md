# PR #23 Review: Web Console UI Dashboard (Epic 6)

## Summary
Adds a Next.js 15 / React 19 web console with login, dashboard, tenant CRUD, GraphQL playground, and metrics charts. Well-structured code with good separation of concerns. Several issues identified below.

## Issues

### 1. [High] Validate tenant ID before URL interpolation
**File:** `web/src/lib/api.ts:113`
`getTenant(id)` and `deleteTenant(id)` interpolate user-controlled `id` (from `useParams()`) directly into the URL path without validation. Should validate UUID format before interpolation.

### 2. [Medium] No Next.js middleware for auth
**File:** `web/src/components/ui/Shell.tsx:14`
Auth check is purely client-side via `localStorage`. Pages are accessible server-side. Add Next.js middleware to redirect unauthenticated requests before page load.

### 3. [Medium] Missing error boundaries
No `error.tsx` files exist. If any page component throws during render, the entire app crashes with a white screen. Add Next.js error boundaries.

### 4. [Medium] `useAsync` stale reference race condition
**File:** `web/src/lib/hooks.ts:35-37`
`fnRef` is updated in an effect without deps (runs after render), but `execute` may call the old reference if deps and fn change in the same render.

### 5. [Low] No loading state on dashboard initial load
**File:** `web/src/app/dashboard/page.tsx`
When `stats` is null and no error, the page shows the header but stats/tenants section is empty. Add a skeleton or spinner.

### 6. [Low] Chart.js `TimeScale` registered but unused
**File:** `web/src/app/metrics/page.tsx:278`
`TimeScale` is registered but `CategoryScale` is used for the x-axis. Remove unused registration.

### 7. [Low] Query history uses array index as React key
**File:** `web/src/app/playground/page.tsx:651`
`key={i}` on history items can cause reconciliation issues with duplicate queries.

### 8. [Low] Root page should be middleware redirect
**File:** `web/src/app/page.tsx`
The root redirect page renders a client component that returns null. Use Next.js middleware instead.

### 9. [Nit] `next.config.js` uses CJS syntax
**File:** `web/next.config.js`
Uses `module.exports` but Next.js 15 supports `next.config.ts`. Use `.ts` for consistency with the rest of the project.

### 10. [Nit] Kapok color palette is just green renamed
**File:** `web/tailwind.config.ts`
The `kapok` colors are identical to Tailwind's default green palette. Consider a distinct brand color.

## Positives
- Clean component structure (Shell, Sidebar, StatCard, Modal)
- Delete confirmation requiring typed tenant name
- GraphQL playground with query history
- Proper `Suspense` usage for `useSearchParams`
- Standalone output mode for containerized deployment
- Good Makefile integration with web-* targets

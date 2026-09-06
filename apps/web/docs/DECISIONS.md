# Decisions

Why each runtime dependency and each notable architectural choice is here, plus
the alternative that was rejected. Keep this current when dependencies change.

## Runtime dependencies

| Dependency | Why it's here | Rejected alternative |
|---|---|---|
| `swr` | Server-cache + revalidation + mutation primitives; the app's single home for server state | React Query — heavier; SWR is the fixed choice for this boilerplate |
| `zustand` | Minimal UI/client-state store with granular selectors and `persist` | Redux / bare Context — more boilerplate and verbosity for the same job |
| `react-hook-form` | Performant uncontrolled forms, good TS ergonomics | Formik — heavier, weaker TS story |
| `zod` | Schema + inferred type from one source of truth | yup — weaker type inference |
| `@hookform/resolvers` | Bridges zod ↔ react-hook-form; major must match installed `zod` | hand-rolled resolver — reinvents validation glue |
| `@base-ui/react` + shadcn primitives | Accessible, unstyled primitives generated into the repo (owned, not vendored at runtime) | MUI / Chakra — runtime CSS-in-JS and opinionated theming |
| `class-variance-authority`, `tailwind-merge`, `clsx` / `cn` | shadcn `cn` helper + variant styling with correct class precedence | ad-hoc class strings — merge and precedence bugs |
| `lucide-react` | Tree-shakeable icon set shadcn expects | react-icons — larger, inconsistent |
| `sonner` | Toast notifications for success/error surfaces | hand-rolled toaster — accessibility and stacking work |

### Dev/test dependencies

| Dependency | Why it's here | Rejected alternative |
|---|---|---|
| MSW | Realistic network mocking for hook/integration tests | manual `fetch` stubs — brittle, diverge from real requests |
| Playwright | E2E happy path | Cypress — heavier; Playwright chosen by preference |
| Vitest + happy-dom + Testing Library | Fast unit/component tests | Jest — slower, heavier config in an ESM/Vite project |

**Version alignment (R-8):** `@hookform/resolvers` must support the installed
`zod` major (currently zod 4, using `z.email()`). If the latest versions
disagree, pin the compatible pair here.

## Cookie & BFF strategy (R-1, R-4)

**Decision:** BFF route handlers under `app/api/*` proxy to the Go backend. The
backend returns `accessToken` + `refreshToken` in the JSON **response body**, so
the BFF is the only place that can convert them into httpOnly cookies
(`arsen_access`, `arsen_refresh`). The browser never sees a token and never talks
to the Go backend directly.

- Cookies: `httpOnly; SameSite=Lax; Path=/`; `Secure` in prod (omitted on
  `http://localhost`).
- `arsen_access` `Max-Age` = `expiresIn` from the login/refresh response.
- `arsen_refresh` `Max-Age` = `SESSION_REFRESH_MAX_AGE` env (default 30 days),
  because the backend doesn't expose refresh expiry.
- Route protection is **presence-based**: `proxy.ts` checks for the presence of
  either cookie, not their validity (validity is the backend's job on each call).

**Rejected alternative:** a single encrypted session cookie (iron-session style)
— viable and slightly more opaque, but adds a dependency and encryption-key
management. Left as a documented upgrade path.

**Rejected alternative:** browser-direct calls with `credentials: "include"` —
impossible here because the backend returns tokens in the body, not via
`Set-Cookie`; the browser would have to hold tokens in JS, violating the
token-secrecy requirement.

## Silent refresh (R-5)

**Decision:** The server-only fetch helper (`lib/api/server.ts`) attaches
`Authorization: Bearer <access cookie>`. On a 401 it calls `POST /api/tokens`
with the refresh cookie **once**, rewrites both cookies, and retries the original
request **once**. A second failure clears cookies and surfaces 401. Refresh is a
server-side BFF concern and is **never** triggered from `useEffect`.

## In-memory `resource` store is a stand-in (R-7)

**Decision:** The `resource` slice's BFF handlers back onto an **in-memory,
per-server-process `Map`** (`src/lib/resource-store.ts`), seeded with a few rows.
It is clearly labeled in code as a rename-me placeholder.

**Why:** The Go backend has no generic resources endpoint. `resource` is an
explicit "replace with your real backend" demonstration of the full
read/write/optimistic chain, without inventing backend scope. It resets on server
restart and is not multi-instance safe — that's intentional for a boilerplate.

**Rejected alternatives:** (a) add a `resources` feature to the Go backend — out
of scope for a frontend feature; (b) point at an existing backend resource — none
exists.

## Spec deviations (R-1, R-2)

Two spec assumptions did not survive contact with the as-built backend and were
deliberately changed:

- **R-1 — endpoint shape & token delivery.** The spec assumed `/auth/*` paths and
  `Set-Cookie` from the backend. Reality: REST-resource paths (`POST /api/sessions`,
  `POST /api/tokens`, `DELETE /api/sessions/current`, `POST /api/users`,
  `GET /api/users/me`) and tokens in the response body. The BFF adapts to the real
  contract.
- **R-2 — register does not log the user in.** `POST /api/users` creates an
  *unverified* user with no tokens; `POST /api/sessions` returns **403** until the
  email is verified. So register success redirects to `/login?registered=1` with a
  "verify your email" notice rather than dropping the user into the app. The
  authenticated E2E path therefore uses a **pre-verified account** (seeded/verified
  out-of-band, since dev email is a no-op). FR-001, US1 scenario 1, FR-027, and
  SC-008 should be read as "register → verify → login → app".

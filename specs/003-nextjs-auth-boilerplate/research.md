# Phase 0 Research: Next.js Feature-Sliced Frontend Boilerplate

All items below resolve unknowns from the Technical Context or record decisions that shape Phase 1. Findings against the real backend (`apps/api`) take precedence over assumptions made in the spec.

## R-1. Backend auth contract (as-built) — token delivery is response-body, endpoints are REST-resource style

**Decision**: The BFF reads `accessToken` + `refreshToken` from the JSON response body and sets them as httpOnly cookies itself. Endpoint paths are the real ones below (not `/auth/login`).

**Findings** (from `apps/api/features/auth/handler.go`, `features/user/handler.go`):

| Purpose | Method + path | Auth | Request body | Success | Errors |
|---|---|---|---|---|---|
| Login | `POST /api/sessions` | none | `{email, password}` | 200 `sessionResponse` | 400, 401, 403 (unverified), 429 |
| Refresh | `POST /api/tokens` | none | `{refreshToken}` | 200 `sessionResponse` | 401, 429 |
| Logout | `DELETE /api/sessions/current` | Bearer | — | 204 | 401 |
| Register | `POST /api/users` | none | `{email, password}` | 201 `userResponse` (no tokens) | 400, 429 |
| Profile | `GET /api/users/me` | Bearer | — | 200 `userResponse` | 401 |

`sessionResponse` = `{self, kind, accessToken, refreshToken, tokenType:"Bearer", expiresIn}` (`expiresIn` in seconds).
`userResponse` = `{self, kind, id, email, emailVerified, createdAt, message?}`.

**Rationale**: Because tokens arrive in the body, the BFF is the only place that can convert them into httpOnly cookies — exactly the boundary the spec requires. `expiresIn` sets the access-cookie `Max-Age`.

**Alternatives considered**: Browser-direct calls with `credentials: 'include'` — rejected: the backend returns tokens in the body (not `Set-Cookie`), so the browser would have to hold them in JS, violating FR-007/FR-008.

## R-2. Email verification gate — register does NOT log the user in

**Decision**: Register (`POST /api/users`) creates an **unverified** user and returns no tokens; login (`POST /api/sessions`) returns **403** until the email is verified (`login_command.go` checks `u.EmailVerified`). The register page therefore redirects to a "check your email to verify" screen, not into the app.

**Impact on spec (must be reconciled)**:
- US1 Acceptance Scenario 1 ("register → session established → redirected into protected area") and **FR-001** are not achievable as written.
- The E2E happy path (**FR-027 / SC-008**: "register → land in app → create resource …") cannot proceed straight from register.

**Resolution for this plan**:
- Register success → redirect to `/login?registered=1` (or a verify-notice page), with copy explaining verification is required.
- The authenticated portion of E2E uses a **pre-verified account**: seed one via the API + a test-only verification step, or via a DB seed in the e2e setup. Dev email is a no-op (root CLAUDE.md), so the verification token is not deliverable by email in dev — the e2e/setup path must verify out-of-band (DB update or a seeded verified user).
- **Recommendation**: update spec FR-001, US1 scenario 1, FR-027, and SC-008 to reflect the verification step (e.g., "register → verify → login → app"). Flagged in the completion report for confirmation.

**Alternatives considered**: Adding a dev-only auto-verify or a token-retrieval endpoint to the Go backend — rejected for now: it changes backend scope (this feature is frontend-only). Left as a follow-up option if a fully register-driven E2E is desired.

## R-3. Error contract → RHF field errors

**Decision**: `ApiError` wraps the RFC 9457 body `{type, title, status, detail, instance, errors?: [{field, detail}]}` (content-type `application/problem+json`). Forms map `errors[]` onto `setError(field, { message: detail })`; absent field errors fall back to a generic message from `title`/`detail`.

**Findings**: `apps/api/pkg/response/problem.go` — `ProblemDetail.Errors` is `[]FieldError{field, detail}`; validation failures return 400 "Validation Failed" with populated `errors`. Registration conflicts are normalized to a generic 400 (anti-enumeration) — so the register form must not assume a field-specific email-taken error.

## R-4. Cookie strategy

**Decision**: Two httpOnly cookies set by the BFF:
- `arsen_access` — the JWT; `Max-Age` = `expiresIn` from the login/refresh response.
- `arsen_refresh` — the opaque refresh token; `Max-Age` = a configured default (env `SESSION_REFRESH_MAX_AGE`, default 30 days) since the backend does not expose refresh expiry in the body.

Both: `httpOnly; Secure (prod); SameSite=Lax; Path=/`. `Secure` is omitted on `http://localhost` in dev. The **presence** of `arsen_access` (or `arsen_refresh`) is what `middleware.ts` checks.

**Rationale**: SameSite=Lax + presence-check were fixed in clarification. Separate cookies keep refresh usable server-side even after the access cookie expires, enabling silent refresh.

**Alternatives considered**: Single encrypted session cookie (iron-session style) — viable and slightly more opaque, but adds a dependency and encryption key management; rejected for boilerplate simplicity. Documented as an upgrade path in DECISIONS.md.

## R-5. Silent refresh flow

**Decision**: The server-only fetch helper (`lib/api/server.ts`) attaches `Authorization: Bearer <access cookie>`. On a 401 from the backend, it calls `POST /api/tokens` with the refresh cookie once, rewrites both cookies, and retries the original request once. A second failure clears cookies and surfaces 401 (BFF returns 401 → client redirects to login). Refresh is never triggered from `useEffect`; it is a server-side concern inside the BFF.

## R-6. RSC → SWR boundary (per route)

**Decision**:
- `(app)/resources/page.tsx` (RSC) fetches the initial list server-side and passes it as `fallback` into a client `<SWRConfig fallback={...}>` boundary; the client list hook (`useResources`) owns revalidation thereafter.
- `(app)/resources/[id]/page.tsx` (RSC) fetches the detail server-side as `fallback` for `useResource(id)`.
- `(app)/dashboard/page.tsx` is mostly static/server; no SWR needed unless it shows live data.
- Auth pages are client components (forms). `useSession` hydrates from the persisted profile and reconciles against `/api/me`.

**Rationale**: Matches the spec's default resolution; avoids double-fetching and gives instant first paint.

## R-7. Demo `resource` backing store

**Decision**: The `resource` BFF handlers (`app/api/resources/*`) use an in-memory, per-server-process store (a module-level `Map`) seeded with a few rows. RSC pages read via a shared server function so first paint and the BFF agree.

**Rationale**: The Go backend has no generic resources endpoint; the spec's `resource` is an explicit rename-me stand-in. An in-memory store keeps the boilerplate runnable and demonstrates the full read/write/optimistic chain without inventing backend scope. Clearly labeled in code + DECISIONS.md as "replace with your real backend."

**Alternatives considered**: (a) Add a `resources` feature to the Go backend — out of scope (frontend feature). (b) Point the resource slice at an existing backend resource — none exists. In-memory chosen; documented as temporary.

## R-8. Dependency list (justifications for DECISIONS.md)

| Dependency | Why | Rejected alternative |
|---|---|---|
| `swr` | Server-cache + revalidation + mutation primitives | React Query — heavier; SWR fixed by spec |
| `zustand` | Minimal UI-state store with granular selectors + persist | Redux/Context — boilerplate/verbosity |
| `react-hook-form` | Performant uncontrolled forms | Formik — heavier, less TS-friendly |
| `zod` | Schema + inferred types, one source of truth | yup — weaker inference |
| `@hookform/resolvers` | Bridges zod ↔ RHF; must match installed zod major | hand-rolled resolver — reinvents the wheel |
| Radix UI (via shadcn) | Accessible unstyled primitives generated into repo | MUI/Chakra — runtime CSS-in-JS, opinionated theme |
| `class-variance-authority`, `tailwind-merge`, `clsx` | shadcn `cn` + variant styling | ad-hoc class strings — merge/precedence bugs |
| `lucide-react` | Tree-shakeable icon set shadcn expects | react-icons — larger, inconsistent |
| MSW (dev) | Realistic network mocking for hook tests | manual fetch stubs — forbidden by spec |
| Playwright (dev) | E2E happy path | Cypress — heavier, chosen by preference |

**Version alignment note**: after install, verify `@hookform/resolvers` supports the installed `zod` major; pin the compatible pair if the latest disagree (record in DECISIONS.md). Confirm whether shadcn init emits Tailwind v4 CSS-first `@theme` (expected, since `apps/web` already uses Tailwind v4) vs a JS config, and follow whatever it generates.

## R-9. Tooling deltas from current `apps/web`

**Findings**: `apps/web` already has Next 16.3.4, React 19.2.8, Tailwind v4, ESLint 9 flat config, Vitest 5, Prettier, `@/*` path alias, `moduleResolution: bundler`, `strict: true`.

**Decision / deltas to apply**:
- Extend `tsconfig.json` with `noUncheckedIndexedAccess`, `noImplicitOverride`, `verbatimModuleSyntax`.
- Extend `eslint.config.mjs`: `typescript-eslint` recommended-type-checked, `no-restricted-imports` (block cross-feature internal imports + global `mutate` misuse), ban `any`, require named exports outside Next reserved files, `eslint-config-prettier` last.
- Add `prettier-plugin-tailwindcss`.
- Add `vitest` MSW setup + coverage config; add Playwright config + `e2e` script.
- Add scripts: `check` (typecheck + lint + test), `e2e`.
- Read `node_modules/next/dist/docs/` before writing Next 16 code (AGENTS.md warns of breaking changes vs training data).

## R-10. Access-token expiry / refresh timing

**Decision**: Reactive refresh (on 401) is the baseline, satisfying FR-006's "or upon an unauthorized proxied response." Proactive pre-expiry refresh is optional; if added, compute from the access cookie's `Max-Age`. The backend's exact access/refresh durations live in Go config and are not exposed to the web app — the web app treats `expiresIn` (from the response) as authoritative for the access cookie only.

# Quickstart & Validation Guide

How to run the boilerplate and validate that the feature works end to end. Details of shapes live in [data-model.md](./data-model.md) and [contracts/](./contracts/); do not duplicate them here.

## Prerequisites

- Node.js 20+, `pnpm` 10+.
- The Go backend running with Postgres + Redis: from repo root `task docker-up` (or `task api:run` with a local `.env`). Backend listens on `http://localhost:8080`.
- A **verified** user account for the authenticated flows (register alone cannot log in — see [research R-2](./research.md)). Create one by registering, then verifying the email out-of-band (dev email is a no-op, so verify via DB or a seed script).

## Setup

```bash
cp apps/web/.env.example apps/web/.env   # ensure NEXT_PUBLIC_API_URL=http://localhost:8080
cd apps/web && pnpm install
```

Env keys used by the web app:
- `NEXT_PUBLIC_API_URL` — Go backend base URL.
- `SESSION_REFRESH_MAX_AGE` (optional) — refresh cookie lifetime in seconds (default 30d).

## Run

```bash
pnpm dev            # Next dev server (App Router)
```

Open `http://localhost:3000` → unauthenticated visit to `/dashboard` should redirect to `/login`.

## Quality gates (must all pass)

```bash
pnpm typecheck      # tsc --noEmit — zero errors
pnpm lint           # eslint — zero warnings
pnpm test           # vitest run — unit/component/hook, coverage on src/features/*
pnpm e2e            # playwright — happy path
pnpm check          # typecheck + lint + test
```

## Validation scenarios

Map to spec acceptance criteria. Each should be demonstrable manually and (where noted) covered by an automated test.

1. **Route protection (FR-005, SC-003)** — visit `/dashboard` or `/resources` while logged out → redirected to `/login?next=…`. Covered by middleware test + e2e final step.
2. **Register (FR-001, adjusted per R-2)** — submit valid new credentials on `/register` → redirected to `/login?registered=1` with a "verify your email" notice. No cookies set. Covered by register-hook test (asserts no token anywhere).
3. **Login (FR-002)** — with a verified account, submit correct credentials → cookies set server-side, redirected into `(app)`. Covered by login-hook test (MSW) + e2e.
4. **Token secrecy (FR-007/FR-008, SC-002)** — after login, assert in an automated test that no token appears in `localStorage`, `sessionStorage`, `authStore` state, or serialized client output; only the profile is present. This is the highest-priority test.
5. **Session persistence (FR-003, SC-004)** — reload the page → still authenticated, profile shown, no re-login. `useSession` reconciles against `/api/me`.
6. **Silent refresh (FR-006)** — with an expired access cookie, perform an action → BFF refreshes via `/api/tokens` and the action succeeds without user interaction. Covered by a server-fetch test simulating a 401-then-refresh.
7. **Resource read + RSC→SWR (FR-018, US3 scenario 1)** — `/resources` first paint is server-rendered; client revalidation does not double-fetch.
8. **Optimistic create (FR-019, SC-006)** — create a resource → it appears in the list immediately; on forced error the optimistic row rolls back. Covered by a create-hook test (bound `mutate`).
9. **Draft persistence (FR-021)** — start the create form, close and reopen → draft restored; on successful submit → draft cleared, modal closed (order: `reset` → `clearDraft` → `closeModal`).
10. **Logout (FR-004)** — logout → upstream session revoked, cookies cleared, profile cleared, redirect to `/login`; protected routes then inaccessible.
11. **Error surfacing (FR-010, R-3)** — invalid login shows a generic message; backend validation `errors[]` map to field-level messages; no raw backend text/tokens/PII leak.

## E2E happy path (Playwright)

`register (→ verify notice) → login with a pre-verified account → land in app → create a resource → see it listed → logout → confirm /resources redirects to /login`.

> Note: the E2E authenticated leg uses a pre-verified account (R-2). If you want a fully register-driven E2E, add a dev/test-only verification step to the backend first.

# Implementation Plan: Next.js Feature-Sliced Frontend Boilerplate

**Branch**: `003-nextjs-auth-boilerplate` | **Date**: 2026-09-05 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/003-nextjs-auth-boilerplate/spec.md`

## Summary

Build a production-grade, feature-sliced Next.js App Router boilerplate inside the existing `apps/web` workspace. It delivers secure end-to-end authentication and one generic CRUD `resource` slice, following fixed house conventions: `zod → zodResolver → react-hook-form → useSWRMutation` for writes and `fetcher → SWR hook → component` for reads; server data in SWR, UI state in Zustand, never mixed.

The access and refresh tokens returned by the Go backend (in the JSON response body) are never exposed to the browser: a thin **BFF layer** of Next Route Handlers under `app/api/*` proxies every call to the Go backend, stores the tokens in httpOnly/Secure/SameSite=Lax cookies, forwards the bearer server-side, and performs silent refresh with one retry on 401. `middleware.ts` guards the protected route group by **presence** of the session cookie (the server proxy is the authority for validity).

**Two deviations from the spec were discovered against the real backend and must be acknowledged (see research.md R-1, R-2):**

1. **Email verification is required before login.** The Go backend (`POST /api/sessions`) returns 403 until the user's email is verified, and register (`POST /api/users`) returns **no tokens**. Therefore "register → land straight in the app" (spec US1 scenario 1, FR-001, and the E2E happy path FR-027/SC-008) is not achievable as written. Plan resolution: register redirects to a "verify your email" screen; the authenticated happy path (and E2E) uses a pre-verified account. This should be reflected back into the spec's acceptance criteria.
2. **No backend `resources` endpoint exists.** The demo `resource` slice's BFF handlers back onto an in-memory, per-server-process store inside the Next server (clearly labeled as a stand-in to be repointed at a real backend). This keeps the boilerplate self-contained and still exercises the full read/write/optimistic chain.

## Technical Context

**Language/Version**: TypeScript 5.x (strict), Node.js 20+

**Primary Dependencies**: Next.js 16.3.4 (App Router), React 19.2.8, `swr`, `zustand`, `react-hook-form`, `zod`, `@hookform/resolvers`, shadcn/ui primitives (Radix UI + `class-variance-authority` + `tailwind-merge` + `clsx`), `lucide-react`, Tailwind CSS v4. Dev/test: Vitest, React Testing Library, `@testing-library/jest-dom`, MSW, `@playwright/test`, ESLint 9 (flat config) + `typescript-eslint`, Prettier + `prettier-plugin-tailwindcss`.

**Storage**: No client-side persistence of secrets. Server session held only in httpOnly cookies. Demo `resource` data held in an in-memory server-process store (stand-in). Non-sensitive user profile cached client-side via Zustand `persist` (localStorage) for instant UI.

**Testing**: Vitest + RTL + MSW for unit/component/hook tests; Playwright for the E2E happy path. TDD for logic-dense units (schemas, hooks, fetchers) per project memory.

**Target Platform**: Modern evergreen browsers (client); Node.js runtime for the Next server + BFF route handlers.

**Project Type**: Web application — frontend workspace (`apps/web`) in the existing monorepo, paired with the Go backend (`apps/api`).

**Performance Goals**: Instant first paint via RSC-fetched data handed to SWR `fallback`; optimistic create reflected in <100 ms perceived; silent refresh transparent to the user.

**Constraints**: Access + refresh tokens MUST never be readable by client JS (httpOnly cookies only). Strict TypeScript (`noUncheckedIndexedAccess`, `noImplicitOverride`, `verbatimModuleSyntax`). Lint must pass with zero warnings. No cross-feature internal imports. No object-typed SWR keys. Bound `mutate` only for optimistic updates.

**Scale/Scope**: Boilerplate scope — two slices (`auth`, `resource`), ~6 BFF auth/me endpoints + resource CRUD endpoints, login/register pages, protected app shell + dashboard + resources list/detail, and the supporting `lib/` infrastructure. Coverage target ≥80% on hooks/fetchers/schemas.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The project constitution (`.specify/memory/constitution.md`) is an unpopulated template — it defines no ratified principles, so there are no formal gates to evaluate. Applying the spirit of the referenced example principles plus this repo's established conventions:

- **Test-First**: Honored — schemas, fetchers, and hooks are written test-first (project memory: TDD required); MSW mocks the network. **PASS**
- **Simplicity / YAGNI**: Honored — exactly two slices, minimal justified dependencies, no speculative abstraction. **PASS**
- **Integration testing**: Honored — Playwright covers the cross-cutting auth + CRUD happy path; MSW covers hook/fetcher contracts. **PASS**
- **Observability / boundaries**: Honored — errors surface as a typed `ApiError`; no tokens/PII logged. **PASS**

No violations → Complexity Tracking left empty.

**Post-Phase 1 re-check**: Design introduces no new violations. The BFF layer and in-memory resource store are the minimum needed to satisfy the token-secrecy requirement and provide a self-contained CRUD demo; both are justified in research.md. **PASS**

## Project Structure

### Documentation (this feature)

```text
specs/003-nextjs-auth-boilerplate/
├── plan.md              # This file
├── research.md          # Phase 0 output — decisions & backend-contract findings
├── data-model.md        # Phase 1 output — entities, schemas, stores, cookies
├── quickstart.md        # Phase 1 output — run & validate guide
├── contracts/
│   ├── backend-endpoints.md   # Upstream Go API contract (as-built)
│   └── bff-endpoints.md       # Next BFF route-handler contract
└── checklists/
    └── requirements.md  # Spec quality checklist (from /speckit-specify)
```

### Source Code (repository root)

```text
apps/web/
├── src/
│   ├── app/
│   │   ├── (auth)/
│   │   │   ├── login/page.tsx
│   │   │   └── register/page.tsx
│   │   ├── (app)/                     # protected group
│   │   │   ├── layout.tsx             # requires session; renders AppShell
│   │   │   ├── dashboard/page.tsx
│   │   │   └── resources/
│   │   │       ├── page.tsx           # RSC fetch → SWR fallback
│   │   │       └── [id]/page.tsx
│   │   ├── api/                       # BFF route handlers → Go backend / in-memory store
│   │   │   ├── auth/
│   │   │   │   ├── login/route.ts     # POST → /api/sessions ; sets cookies
│   │   │   │   ├── register/route.ts  # POST → /api/users
│   │   │   │   ├── refresh/route.ts   # POST → /api/tokens ; rotates cookies
│   │   │   │   └── logout/route.ts    # DELETE → /api/sessions/current ; clears cookies
│   │   │   ├── me/route.ts            # GET → /api/users/me
│   │   │   └── resources/
│   │   │       ├── route.ts           # GET list, POST create
│   │   │       └── [id]/route.ts      # GET, PATCH, DELETE
│   │   ├── layout.tsx                 # root: providers, fonts, <html>
│   │   ├── providers.tsx             # SWRConfig + toaster
│   │   ├── error.tsx
│   │   ├── not-found.tsx
│   │   └── globals.css
│   ├── features/
│   │   ├── auth/
│   │   │   ├── components/           # LoginForm, RegisterForm, UserMenu
│   │   │   ├── hooks/                # useLogin, useRegister, useSession, useLogout
│   │   │   ├── fetchers/             # login, register, logout, me
│   │   │   ├── schemas/              # loginSchema, registerSchema (+ inferred types)
│   │   │   ├── store.ts             # authStore (profile only)
│   │   │   └── types.ts
│   │   └── resource/
│   │       ├── components/           # ResourceList, ResourceCard, ResourceForm, ResourceFilters
│   │       ├── hooks/                # useResources, useResource, useCreate/Update/DeleteResource
│   │       ├── fetchers/
│   │       ├── schemas/
│   │       ├── store.ts             # resourceStore (filters, draft, wizard)
│   │       └── types.ts
│   ├── components/
│   │   ├── ui/                        # shadcn primitives (generated)
│   │   └── shared/                    # AppShell, DataTable, etc.
│   ├── lib/
│   │   ├── api/                       # client.ts, error.ts, server.ts
│   │   ├── auth/                      # cookie names + session cookie helpers (server-only)
│   │   ├── stores/                    # ui-store.ts
│   │   ├── swr/                       # config.ts, keys.ts
│   │   └── utils/                     # cn.ts
│   └── test/
│       ├── setup.ts                   # RTL + jest-dom + MSW server
│       └── msw/                       # handlers / fixtures
├── e2e/                               # Playwright specs
├── docs/                             # ARCHITECTURE.md, DECISIONS.md
├── middleware.ts                     # presence-based route protection for (app)
├── components.json                   # shadcn config
├── eslint.config.mjs / prettier.config.mjs / tsconfig.json / vitest.config.ts
└── package.json
```

**Structure Decision**: Feature-sliced layout inside the existing `apps/web` workspace (already scaffolded with Next 16.3.4 / React 19.2.8 / Tailwind v4 / Vitest). `lib/*` never imports `features/*`; feature slices never import each other except the documented `auth` → app-shell `useSession` exception. Route files stay thin. All backend access flows through the BFF under `app/api/*`; the browser never calls the Go backend directly.

## Complexity Tracking

> No constitution violations to justify — section intentionally empty.

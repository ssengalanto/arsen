@AGENTS.md

# Web App (apps/web/)

Next.js 16 (App Router) + React 19 + TypeScript + Tailwind CSS v4 frontend,
organized as **feature slices**. It talks to the Go backend (`apps/api`)
exclusively through its own BFF route handlers, converting backend tokens into
httpOnly cookies so the browser never holds a token.

For monorepo-wide commands and architecture overview, see the root
[CLAUDE.md](../../CLAUDE.md).

## Read these first

- [`README.md`](./README.md) — clone → working login in ~5 minutes
- [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md) — slice anatomy, read/write chains, and the step-by-step "add a feature slice" guide
- [`docs/DECISIONS.md`](./docs/DECISIONS.md) — why each dependency/decision (referenced as R-1, R-2, … throughout the code)

## Next.js 16 Warning

This project uses Next.js 16, which has breaking changes from earlier versions.
Before writing any Next.js code, read the relevant guide in
`node_modules/next/dist/docs/` — your training data may not reflect the current
API surface. (This is also why `AGENTS.md` is imported above.)

## Commands

All commands run from this directory or via `task web:<cmd>` from the repo root.

```bash
pnpm dev             # Dev server (or: task web:dev)
pnpm build           # Production build (or: task web:build)
pnpm test            # Vitest, unit + integration, MSW-mocked (or: task web:test)
pnpm test:watch      # Vitest watch mode (or: task web:test-watch)
pnpm e2e             # Playwright end-to-end
pnpm lint            # ESLint incl. the feature import-boundary rule (or: task web:lint)
pnpm typecheck       # tsc --noEmit, strict (or: task web:typecheck)
pnpm format          # Prettier (or: task web:format)
pnpm check           # typecheck + lint + test — run before pushing
pnpm storybook       # component catalog on http://localhost:6006
pnpm build-storybook # static catalog build → storybook-static/ (CI gate)
```

## Project Structure

```
src/
  features/<name>/     # self-contained slices — the app's real surface area
    index.ts           #   PUBLIC ENTRY — the only thing other code may import
    types.ts           #   domain types (never token fields)
    schemas/           #   zod schemas + inferred input types (one source of truth)
    store.ts           #   Zustand UI/client state
    fetchers/          #   thin functions hitting relative BFF URLs via lib/api/client
    hooks/             #   read (useSWR) / write (useSWRMutation) chains
    components/         #   the slice's UI
  app/                 # Next.js routes: pages (RSC) + BFF route handlers under app/api/*
  lib/                 # shared, feature-agnostic plumbing (api, swr, auth cookies, stores)
  components/ui/       # shadcn / base-ui primitives (generated, shared)
  components/shared/   # cross-feature presentational components (if any)
  test/                # Vitest setup + MSW handlers/fixtures
```

Path alias: `@/*` maps to `./src/*`. Reference slices: `features/auth/` (auth +
session) and `features/resource/` (full CRUD + optimistic updates).

## Rules that lint enforces

- **No cross-slice reach-ins.** A feature imports only another feature's public
  entry `@/features/<name>` — never `@/features/*/*`. Within a slice use relative
  paths. Enforced by `no-restricted-imports` and `features/__tests__/import-boundary.test.ts`.
- **Named exports only**, except Next.js special files (`page`/`layout`/`error`/…),
  config files, and Storybook (`*.stories.tsx`, `.storybook/**`).
- **Never use the global `mutate` from `swr`** — use the bound `mutate` from
  `useSWRConfig()` or the mutation hook return.

## Data flow

- **Read:** `component → hook (useSWR) → fetcher → /api/... (BFF) → Go backend`.
  Components never call fetch directly and never store server data in Zustand.
- **Write:** `zod schema → zodResolver → react-hook-form → mutation hook → cache update`.
  Map RFC 9457 problem+json field errors onto `form.setError`.
- **Server state → SWR; UI/client state → Zustand.** The only exception is the
  auth store persisting the non-sensitive profile for instant hydration, with
  `/api/me` (SWR) authoritative.

## BFF & tokens

The browser never calls the Go API directly. Route handlers under `app/api/*`
proxy to the backend (`UPSTREAM_URL`, derived from `NEXT_PUBLIC_API_URL`) and set
tokens as httpOnly cookies. Tokens must never appear in `localStorage`,
`sessionStorage`, or any store — asserted by `features/auth/token-secrecy.test.ts`.

Inside Docker Compose, `docker-compose.yml` overrides `NEXT_PUBLIC_API_URL` to the
API's service name (`http://api-dev:8080`) so the server-side BFF resolves it
in-network; on the host it defaults to `http://localhost:8080`.

## Component catalog (Storybook)

Reusable primitives in `components/ui` (and any future `components/shared`) are
cataloged in Storybook. **Before building a new reusable component, run
`pnpm storybook` and search the catalog — reuse an existing primitive instead of
duplicating UI.** Every shared component ships a co-located `*.stories.tsx`; a
broken or missing story fails `pnpm build-storybook` (the CI gate). Prop tables
are auto-generated via `autodocs` + `react-docgen-typescript`. Feature-internal
components stay private to their slice and are not cataloged.

## Testing

- **Vitest** + React Testing Library + happy-dom for unit/integration; network is
  mocked with **MSW** (`src/test/msw/`).
- **Playwright** for end-to-end (`e2e/`, `playwright.config.ts`).
- Follow TDD: write the failing test first, then implement.

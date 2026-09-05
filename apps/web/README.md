# arsen web

Next.js 16 feature-sliced frontend for the `arsen` stack. It talks to the Go
backend (`apps/api`) exclusively through its own BFF route handlers, converting
backend tokens into httpOnly cookies so the browser never holds a token.

- **Architecture & how to add a feature:** [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md)
- **Why each dependency / decision:** [`docs/DECISIONS.md`](./docs/DECISIONS.md)

## Quick start (clone → working login in ~5 minutes)

### Prerequisites

- Node 20+ and **pnpm 10** (`corepack enable`)
- The Go backend running with Postgres + Redis. From the repo root:
  `task docker-up` starts the full dev stack, or run `task api:run` with a local
  `.env` + Postgres + Redis.

### 1. Install & configure

```bash
pnpm install
cp .env.example .env
```

`.env` values:

| Var | Default | Meaning |
|---|---|---|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | Base URL of the Go backend the BFF proxies to |
| `SESSION_REFRESH_MAX_AGE` | `2592000` (30 days) | Refresh-cookie lifetime in seconds |

### 2. Have a **verified** account ready

Registration creates an **unverified** user and does **not** log you in — the
backend returns **403** on login until the email is verified (see R-2 in
[`docs/DECISIONS.md`](./docs/DECISIONS.md)). Dev email is a no-op, so verify
out-of-band:

- Register at `/register` (or `POST /api/users`), then flip `email_verified` to
  `true` for that user directly in Postgres, **or**
- seed a pre-verified user in your DB setup.

Then sign in at `/login`.

### 3. Run

```bash
pnpm dev
```

Open <http://localhost:3000>. Sign in with the verified account → you land on
`/dashboard`. Protected routes (`/dashboard`, `/resources`) redirect to `/login`
when no session cookie is present.

## Scripts

```bash
pnpm dev          # Next.js dev server
pnpm build        # production build
pnpm start        # serve the production build
pnpm test         # Vitest (unit + integration, MSW-mocked)
pnpm test:watch   # Vitest watch mode
pnpm e2e          # Playwright end-to-end
pnpm lint         # ESLint (incl. the feature import-boundary rule)
pnpm typecheck    # tsc --noEmit (strict)
pnpm format       # Prettier
pnpm check        # typecheck + lint + test (run before pushing)
```

All of the above are also available from the repo root via `task web:<cmd>`.

## Project layout

```
src/
  features/<name>/   # self-contained slices (schemas, store, fetchers, hooks, components)
  app/               # Next.js routes: pages (RSC) + BFF route handlers under app/api/*
  lib/               # shared plumbing: api client, swr config, auth cookies, ui store
  components/ui/      # shadcn / base-ui primitives
```

See [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md) for the read/write chains,
the server-state-vs-UI-state rules, and the step-by-step guide to adding a slice.

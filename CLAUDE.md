# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

This is a Go + Next.js monorepo orchestrated by [Task](https://taskfile.dev). All commands use `task` from the repo root unless noted.

### Root-level (runs across both workspaces)
```bash
task test          # Run all tests (Go + Vitest)
task lint          # Run all linters (golangci-lint + ESLint)
task build         # Build both workspaces
task docker-up     # Start dev stack via Docker Compose (hot reload)
task docker-down   # Stop containers and remove volumes
```

### API (`apps/api/`)
```bash
task api:test                        # go test ./...
task api:test -- -run TestName       # Run a single test
task api:test -- -v ./features/auth/ # Verbose tests in one package
task api:test-coverage               # HTML coverage report
task api:lint                        # golangci-lint
task api:lint-fix                    # Auto-fix lint issues
task api:build                       # Compile binary to apps/api/bin/arsen
task api:run                         # Run locally (needs .env + Postgres + Redis)
task api:swagger                     # Regenerate Swagger docs
task api:migrate-up                  # Apply all pending migrations
task api:migrate-down                # Rollback last migration
task api:migrate-create -- name      # Create new migration pair
task api:tidy                        # go mod tidy
```

### Web (`apps/web/`)
```bash
task web:dev          # Next.js dev server
task web:build        # Production build
task web:test         # Vitest
task web:test-watch   # Vitest watch mode
task web:lint         # ESLint
task web:typecheck    # TypeScript strict check
task web:format       # Prettier
```

## Architecture

### Monorepo Layout

- `go.work` — Go workspace linking `apps/api`
- `apps/api/` — Go backend (chi, uber/fx, pgx/sqlx, Redis)
- `apps/web/` — Next.js 16 frontend (React 19, TypeScript, Tailwind v4, App Router)
- `specs/` — Feature specifications and plans (managed by Spec Kit)

### API: Vertical Slice + CQRS

Each feature is a self-contained package under `features/` with no cross-feature repository imports:

```
features/auth/
  module.go              # fx.Module: wires repository, handlers, command buses, routes
  handler.go             # HTTP handlers (decode request → dispatch command/query → encode response)
  login_command.go       # Command struct + CommandHandler implementation
  repository.go          # Repository interface + SQL implementation
  *_test.go              # Co-located unit tests
```

**Adding a new feature:**
1. Create `features/<name>/` with handler, commands/queries, repository, and `module.go`
2. Each command/query gets its own `CommandBus`/`QueryBus` with middleware (recovery, logging, validation) wired in `module.go`
3. Register the `fx.Module` in `cmd/api/main.go`

**CQRS buses** (`pkg/cqrs/`) use Go generics: `CommandBus[C, R]` dispatches commands, `QueryBus[Q, R]` dispatches queries. `cqrs.Unit` is the return type for commands with no result. Middleware wraps the handler chain in reverse order.

**Cross-feature data access** — Features never import another feature's repository. Instead, the owning feature exposes a query via its `QueryBus`, and consumers inject the bus. For example, auth needs user data for login, so user exposes `GetUserByEmailQuery` / `GetUserByEmailResult` through `*cqrs.QueryBus[user.GetUserByEmailQuery, *user.GetUserByEmailResult]`. Auth injects this bus — it never touches `user.Repository`. This keeps each slice's data access private and routes all cross-feature reads through the CQRS bus.

### Dependency Injection (uber/fx)

All wiring happens through `fx.Module` declarations. The app is composed in `cmd/api/main.go` — each package provides its dependencies and features register their own routes via `fx.Invoke`.

### Shared Infrastructure (`apps/api/pkg/`)

- `config/` — Viper-based 12-factor config from env vars
- `server/` — Chi router setup with global middleware (RequestID → Logging → Recovery → CORS)
- `middleware/` — HTTP middleware: auth (JWT), CORS, logging, rate limiting, request ID
- `jwt/` — HS256 JWT service for access/refresh tokens
- `database/` — PostgreSQL connection (pgx + sqlx)
- `redis/` — Redis client wrapper
- `email/` — Resend integration with no-op fallback for dev
- `response/` — RFC 9457 problem details for all error responses
- `cqrs/` — Generic command/query buses with middleware
- `token/` — 256-bit cryptographic token generation
- `cleanup/` — Background worker that purges expired tokens on a configurable interval

### Key Design Decisions

- **RFC 9457 problem details** — All errors return `application/problem+json`, including 404/405 from the router.
- **Anti-enumeration** — Registration, forgot-password, and resend-verification return identical responses regardless of whether the email exists.
- **Token rotation with family model** — Refresh tokens are grouped by `family_id`. On refresh, the old family is revoked and a new token issued. Reuse of a revoked token invalidates the entire family (theft detection).
- **Migrations** — SQL files in `apps/api/migrations/` managed by golang-migrate. Four tables: `users`, `refresh_tokens`, `verification_tokens`, `password_reset_tokens`.

### Web: Feature Slices + BFF

Next.js 16 (App Router) + React 19 + TypeScript + Tailwind v4, organized as **feature slices**. Each slice under `src/features/<name>/` owns everything it needs (schemas, store, fetchers, hooks, components) behind a single public entry (`index.ts`). Cross-slice reach-ins are lint-blocked — a feature imports only another feature's public entry `@/features/<name>`, never `@/features/*/*`. Shared, feature-agnostic plumbing lives in `lib/*`; shared UI primitives in `components/ui`.

**BFF pattern** — The browser never calls the Go API directly. Route handlers under `app/api/*` proxy to the backend and convert tokens into httpOnly cookies, so tokens never reach client JS (asserted by `features/auth/token-secrecy.test.ts`). The upstream base URL comes from `NEXT_PUBLIC_API_URL`; `docker-compose.yml` overrides it to the API's compose service name so the server-side BFF resolves it in-network.

**State split** — Server state flows `component → hook (SWR) → fetcher → /api/* (BFF)`; UI/client state lives in Zustand. Never store server data in Zustand as the source of truth.

**Storybook** — Reusable primitives in `components/ui` are cataloged in Storybook with co-located `*.stories.tsx`. `pnpm build-storybook` is a CI gate (`.github/workflows/storybook.yml`). Check the catalog before building a new shared component.

The `AGENTS.md` file (auto-generated by `next dev`) warns that this Next.js version has breaking changes from training data — read `node_modules/next/dist/docs/` before writing Next.js code. See `apps/web/docs/ARCHITECTURE.md` for slice anatomy and the "add a feature slice" guide, and `apps/web/docs/DECISIONS.md` for the rationale behind each dependency.

## Environment Setup

```bash
cp apps/api/.env.example apps/api/.env
cp apps/web/.env.example apps/web/.env
```

Docker Compose provides Postgres 16 and Redis 7. `task docker-up` starts everything with hot reload (Go uses Air, Next.js uses its dev server).

## Testing

- **Go**: `testify` for assertions. Unit tests are co-located (`*_test.go`). Integration tests live in `apps/api/tests/integration/`. Table-driven tests are the standard pattern.
- **Web**: Vitest + React Testing Library with happy-dom for unit/integration (network mocked via MSW in `src/test/msw/`); Playwright for end-to-end (`apps/web/e2e/`). The feature import-boundary is verified by a lint test (`features/__tests__/import-boundary.test.ts`).

# Data Model: Monorepo Restructure

This feature is a structural/infrastructure change — no new database entities, tables, or data models are introduced.

## Workspace Entities (Filesystem)

### API Workspace (`apps/api/`)

Contains all existing entities from the JWT auth boilerplate (users, refresh_tokens, verification_tokens, password_reset_tokens). No changes to the data model — everything moves as-is into the `apps/api/` directory.

### Web Workspace (`apps/web/`)

No server-side data model. The web frontend consumes the API via HTTP. Client-side state management (if any) will be defined in future feature specs.

## File Layout Model

```
/                          # Repository root
├── go.work                # Go workspace file (points to apps/api)
├── Taskfile.yml           # Root orchestrator (includes api: and web:)
├── docker-compose.yml     # Multi-service dev/prod compose
├── .env.example           # Reference for all env vars across workspaces
├── .gitignore             # Root-level shared ignore patterns
├── README.md              # Monorepo overview and getting started
│
├── apps/
│   ├── api/               # Go backend (moved from repo root)
│   │   ├── go.mod         # Module: arsen (unchanged)
│   │   ├── go.sum
│   │   ├── Taskfile.yml   # API-specific tasks (build, test, lint, etc.)
│   │   ├── Dockerfile     # Multi-stage Go build
│   │   ├── .air.toml      # Hot reload config
│   │   ├── .golangci.yml  # Linter config
│   │   ├── .env.example   # API env vars
│   │   ├── .dockerignore
│   │   ├── cmd/
│   │   ├── features/
│   │   ├── pkg/
│   │   ├── migrations/
│   │   ├── docs/
│   │   └── tests/
│   │
│   └── web/               # Next.js frontend (new)
│       ├── package.json
│       ├── pnpm-lock.yaml
│       ├── Taskfile.yml   # Web-specific tasks (dev, build, test, lint)
│       ├── Dockerfile     # Multi-stage Node build
│       ├── .env.example   # Web env vars (NEXT_PUBLIC_API_URL, etc.)
│       ├── next.config.ts
│       ├── tsconfig.json
│       ├── vitest.config.ts
│       ├── src/
│       │   └── app/       # Next.js App Router
│       └── __tests__/
│
├── specs/                 # Spec Kit feature specs (stays at root)
└── .specify/              # Spec Kit config (stays at root)
```

## Files That Move (API)

All existing files at the repo root that belong to the Go API move into `apps/api/`:

| Current Location | New Location |
|-----------------|-------------|
| `go.mod`, `go.sum` | `apps/api/go.mod`, `apps/api/go.sum` |
| `cmd/` | `apps/api/cmd/` |
| `features/` | `apps/api/features/` |
| `pkg/` | `apps/api/pkg/` |
| `migrations/` | `apps/api/migrations/` |
| `docs/` | `apps/api/docs/` |
| `tests/` | `apps/api/tests/` |
| `Dockerfile` | `apps/api/Dockerfile` |
| `.air.toml` | `apps/api/.air.toml` |
| `.golangci.yml` | `apps/api/.golangci.yml` |
| `.dockerignore` | `apps/api/.dockerignore` |
| `.env.example` | `apps/api/.env.example` |
| `README.md` | `apps/api/README.md` (API-specific readme) |

## Files That Stay at Root

| File | Reason |
|------|--------|
| `.git/` | Git history |
| `.gitignore` | Shared ignore rules |
| `specs/` | Spec Kit feature specs |
| `.specify/` | Spec Kit configuration |
| `.claude/` | Claude Code configuration |
| `.github/` | GitHub configuration |

## Files Created or Rewritten at Root

| File | Purpose |
|------|---------|
| `go.work` | Go workspace file pointing to `apps/api` |
| `Taskfile.yml` | Root orchestrator delegating to `api:` and `web:` |
| `docker-compose.yml` | Updated with `apps/api` and `apps/web` build contexts |
| `.env.example` | Reference listing all env vars across both workspaces |
| `README.md` | New monorepo overview |

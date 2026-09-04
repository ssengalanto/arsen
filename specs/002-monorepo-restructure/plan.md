# Implementation Plan: Monorepo Restructure

**Branch**: `002-monorepo-restructure` | **Date**: 2026-09-04 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/002-monorepo-restructure/spec.md`

## Summary

Restructure the repository from a single Go project at the root into a monorepo with two workspaces: `apps/api/` (existing Go backend, moved) and `apps/web/` (new Next.js + TypeScript frontend). The Go module path `arsen` is preserved. A root-level `go.work` file, Taskfile, and Docker Compose orchestrate both workspaces. No Turborepo — Taskfile v3 with `includes:` is the sole orchestrator.

## Technical Context

**Language/Version**: Go 1.26.2 (API), TypeScript 5.x + Next.js 16 (Web)

**Primary Dependencies**:
- API: chi, sqlx, pgx, fx, golang-jwt, resend-go, viper, swaggo (unchanged)
- Web: Next.js 16, React 19, Tailwind CSS v4, Vitest, React Testing Library, ESLint, Prettier

**Storage**: PostgreSQL 16 (unchanged), Redis 7 (unchanged)

**Testing**: Go `testing` + testify (API), Vitest + React Testing Library + happy-dom (Web)

**Target Platform**: Linux server (Docker), macOS/Linux dev machines

**Project Type**: Monorepo — web service (Go API) + web application (Next.js frontend)

**Performance Goals**: N/A — structural change, no new runtime behavior

**Constraints**: Go module path `arsen` must not change; all existing API tests must pass without modification

**Scale/Scope**: 2 workspaces, ~50 Go source files moving, 1 new Next.js scaffold

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Constitution template is unpopulated (placeholders only) — no project-specific governance constraints defined. Gate passes by default.

**Post-Phase 1 re-check**: Still passes. No violations introduced.

## Project Structure

### Documentation (this feature)

```text
specs/002-monorepo-restructure/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output (file layout model)
├── quickstart.md        # Phase 1 output (7 validation scenarios)
├── contracts/
│   ├── taskfile-interface.md      # CLI command contract
│   └── docker-compose-interface.md # Service topology contract
└── tasks.md             # Phase 2 output (created by /speckit-tasks)
```

### Source Code (repository root)

```text
/                              # Repository root
├── go.work                    # Go workspace: use ./apps/api
├── Taskfile.yml               # Root orchestrator (includes api:, web:)
├── docker-compose.yml         # Multi-service: api, web, postgres, redis
├── .env.example               # Reference for all env vars
├── .gitignore                 # Shared ignore patterns (Go + Node + OS)
├── README.md                  # Monorepo overview
│
├── apps/
│   ├── api/                   # Go backend (moved from repo root)
│   │   ├── go.mod             # Module: arsen (unchanged)
│   │   ├── go.sum
│   │   ├── Taskfile.yml       # API tasks: build, test, lint, migrate, swagger
│   │   ├── Dockerfile         # Multi-stage: dev (Air) + build + prod (distroless)
│   │   ├── .air.toml
│   │   ├── .golangci.yml
│   │   ├── .env.example
│   │   ├── .dockerignore
│   │   ├── cmd/api/
│   │   ├── features/          # auth/, user/, health/
│   │   ├── pkg/               # cqrs/, config/, database/, email/, jwt/, etc.
│   │   ├── migrations/
│   │   ├── docs/swagger/
│   │   └── tests/integration/
│   │
│   └── web/                   # Next.js frontend (new)
│       ├── package.json
│       ├── pnpm-lock.yaml
│       ├── Taskfile.yml       # Web tasks: dev, build, test, lint, format, typecheck
│       ├── Dockerfile         # Multi-stage: dev (pnpm dev) + build + prod (standalone)
│       ├── .env.example       # NEXT_PUBLIC_API_URL, etc.
│       ├── next.config.ts
│       ├── tsconfig.json
│       ├── vitest.config.ts
│       ├── eslint.config.mjs
│       ├── .prettierrc
│       ├── src/
│       │   └── app/           # App Router: layout.tsx, page.tsx
│       └── __tests__/         # Vitest test files
│
├── specs/                     # Spec Kit (stays at root)
├── .specify/                  # Spec Kit config (stays at root)
├── .claude/                   # Claude Code config (stays at root)
└── .github/                   # GitHub config (stays at root)
```

**Structure Decision**: Conventional `apps/` monorepo layout with `apps/api/` and `apps/web/`. All Go files move into `apps/api/`; `go.work` at root enables IDE and CLI ergonomics. Root Taskfile delegates via `includes:` to workspace-specific Taskfiles. Docker Compose at root references workspace-specific Dockerfiles.

## Key Technical Decisions

| Decision | Choice | Reference |
|----------|--------|-----------|
| Directory layout | `apps/api/` + `apps/web/` | [research.md #1](./research.md) |
| Go module path | Keep `arsen`, add `go.work` | [research.md #1](./research.md) |
| Next.js version | 16 (latest stable), App Router, `src/` dir | [research.md #2](./research.md) |
| Web test framework | Vitest + React Testing Library + happy-dom | [research.md #3](./research.md) |
| Docker strategy | Per-workspace Dockerfiles and build contexts | [research.md #4](./research.md) |
| Build orchestration | Taskfile v3 `includes:` (no Turborepo) | [research.md #5](./research.md) |
| .gitignore strategy | Root shared + workspace-specific | [research.md #6](./research.md) |
| Web package manager | pnpm | Spec assumption |
| CSS framework | Tailwind CSS v4 (ships with Next.js 16) | [research.md #2](./research.md) |

## Complexity Tracking

No constitution violations. No complexity justification needed.

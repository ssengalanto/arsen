# Tasks: Monorepo Restructure

**Input**: Design documents from `specs/002-monorepo-restructure/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Create the `apps/` directory structure and prepare for the file move.

- [ ] T001 Create `apps/api/` and `apps/web/` directories at the repository root
- [ ] T002 Create `go.work` at the repository root with `go 1.26.2` and `use ./apps/api` — this file enables root-level Go tooling and IDE support after the move

**Checkpoint**: Directory structure exists. No files have moved yet.

---

## Phase 2: Foundational (File Migration)

**Purpose**: Move all Go source code into `apps/api/` and verify the build is intact. This is the critical migration step — everything else depends on it.

**CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T003 Move all Go source directories and files into `apps/api/`: `cmd/`, `features/`, `pkg/`, `migrations/`, `docs/`, `tests/`, `go.mod`, `go.sum`. Use `git mv` to preserve history.
- [ ] T004 Move API configuration files into `apps/api/`: `.air.toml`, `.golangci.yml`, `.dockerignore`, `.env.example`, `Dockerfile`. Use `git mv` to preserve history.
- [ ] T005 Update `apps/api/.air.toml` — set `root = "."` and verify `cmd` path is `go build -o ./tmp/arsen ./cmd/api` (paths are relative to `apps/api/` now)
- [ ] T006 Update `apps/api/.golangci.yml` — verify `goimports.local-prefixes` is still `arsen` and all paths are correct relative to `apps/api/`
- [ ] T007 Verify Go build passes: run `go build ./...` from `apps/api/` and `go build ./apps/api/...` from the repo root (via `go.work`)
- [ ] T008 Verify all Go tests pass: run `go test ./... -count=1` from `apps/api/` — all existing tests must pass with zero modifications

**Checkpoint**: All Go code lives in `apps/api/`. Build and tests pass. `go.work` enables root-level tooling.

---

## Phase 3: User Story 1 — Monorepo Project Structure (Priority: P1) MVP

**Goal**: The API workspace is fully self-contained at `apps/api/` with its own Taskfile, Dockerfile, and README.

**Independent Test**: `cd apps/api && task build && task test` succeeds. All existing API behavior is unchanged.

- [ ] T009 [US1] Create `apps/api/Taskfile.yml` with all API-specific tasks (build, run, test, test-coverage, lint, lint-fix, swagger, migrate-up, migrate-down, migrate-create, tidy) — adapt from the current root `Taskfile.yml` with paths relative to `apps/api/`
- [ ] T010 [US1] Update `apps/api/Dockerfile` — verify build context paths are correct for the `apps/api/` directory (COPY go.mod, go.sum, COPY . ., build path `./cmd/api`)
- [ ] T011 [US1] Move the current root `README.md` to `apps/api/README.md` and update it to describe the API workspace specifically (remove monorepo-level content, adjust paths)
- [ ] T012 [US1] Update the root `.gitignore` — keep shared patterns (`.env`, `.DS_Store`, IDE files), add Node.js/Next.js patterns (`node_modules/`, `.next/`, `.turbo/`), remove Go-specific patterns that now belong in `apps/api/.gitignore`
- [ ] T013 [US1] Create `apps/api/.gitignore` with Go-specific patterns (`bin/`, `*.exe`, `*.test`, `*.out`, `tmp/`, `coverage.out`, `coverage.html`, `build-errors.log`, `vendor/`)
- [ ] T014 [US1] Remove the old root-level `Taskfile.yml`, `Dockerfile`, `.air.toml`, `.golangci.yml`, `.dockerignore`, `.env.example` (they now live in `apps/api/`; root Taskfile will be recreated in US3)
- [ ] T015 [US1] Verify `task build` and `task test` work from within `apps/api/` using the new `apps/api/Taskfile.yml`
- [ ] T016 [US1] Verify `task lint` passes from within `apps/api/` using the workspace-local `.golangci.yml`

**Checkpoint**: `apps/api/` is a self-contained Go workspace. Build, test, and lint all work from within the directory.

---

## Phase 4: User Story 2 — Web Frontend Foundation (Priority: P2)

**Goal**: A fully configured Next.js + TypeScript project at `apps/web/` with Vitest, ESLint, Prettier, and Tailwind CSS.

**Independent Test**: `cd apps/web && pnpm install && pnpm dev` starts the Next.js dev server. `pnpm test` runs Vitest.

- [ ] T017 [US2] Scaffold Next.js project in `apps/web/` using `pnpm create next-app@latest apps/web --typescript --tailwind --eslint --app --src-dir --use-pnpm` (run from repo root; if directory exists, scaffold into it)
- [ ] T018 [US2] Verify TypeScript strict mode is enabled in `apps/web/tsconfig.json` — set `"strict": true` if not already set by the scaffold
- [ ] T019 [US2] Configure Prettier in `apps/web/` — create `.prettierrc` with project conventions, add `prettier` dev dependency, add `format` script to `package.json`
- [ ] T020 [US2] Install and configure Vitest in `apps/web/` — add `vitest`, `@vitejs/plugin-react`, `happy-dom`, `@testing-library/react`, `@testing-library/jest-dom`, `@testing-library/user-event` as dev dependencies. Create `apps/web/vitest.config.ts` with React plugin and happy-dom environment. Add `test` and `test:watch` scripts to `package.json`.
- [ ] T021 [US2] Create a sample test file at `apps/web/__tests__/page.test.tsx` that imports and renders the home page component, asserting it mounts without errors — verifies the Vitest + React Testing Library setup works end-to-end
- [ ] T022 [US2] Create `apps/web/.env.example` with `NEXT_PUBLIC_API_URL=http://localhost:8080` and any other web-specific environment variables
- [ ] T023 [US2] Create `apps/web/Taskfile.yml` with web-specific tasks: dev, build, start, test, test-watch, lint, lint-fix, format, typecheck — each delegating to the corresponding `pnpm` script
- [ ] T024 [US2] Verify the full web workflow: `cd apps/web && pnpm install && task dev` starts the dev server, `task test` runs Vitest, `task lint` runs ESLint, `task typecheck` runs `tsc --noEmit`

**Checkpoint**: `apps/web/` is a fully configured Next.js workspace. Dev server, tests, linting, and type checking all work.

---

## Phase 5: User Story 3 — Shared Development Workflow (Priority: P3)

**Goal**: A root-level Taskfile that orchestrates both workspaces with aggregate commands.

**Independent Test**: `task test` from the repo root runs both Go and Vitest tests. `task lint` runs both linters.

- [ ] T025 [US3] Create the root `Taskfile.yml` with `includes:` for `api:` (pointing to `apps/api/Taskfile.yml` with `dir: ./apps/api`) and `web:` (pointing to `apps/web/Taskfile.yml` with `dir: ./apps/web`)
- [ ] T026 [US3] Add aggregate tasks to the root `Taskfile.yml`: `test` (runs `api:test` then `web:test`), `lint` (runs `api:lint` then `web:lint`), `build` (runs `api:build` then `web:build`)
- [ ] T027 [US3] Create the root `README.md` — monorepo overview explaining the `apps/api/` and `apps/web/` layout, prerequisites (Go, Node.js, pnpm, Docker, Taskfile), getting started instructions for each workspace, and available root-level commands
- [ ] T028 [US3] Create the root `.env.example` — reference file documenting all environment variables across both workspaces with section headers (API, Web, Database, Redis)
- [ ] T029 [US3] Verify aggregate commands work: `task test`, `task lint`, `task build` from the repo root execute both workspaces and report results

**Checkpoint**: Developers can operate both workspaces from the repo root with a single command.

---

## Phase 6: User Story 4 — Docker Compose Multi-Service (Priority: P4)

**Goal**: Docker Compose starts the full stack (API + web + Postgres + Redis) with hot reload in dev mode.

**Independent Test**: `docker compose --profile dev up --build` starts all 4 services. `curl localhost:8080/healthz` and `curl localhost:3000` both succeed.

- [ ] T030 [US4] Create `apps/web/Dockerfile` — multi-stage: `dev` target (pnpm install + pnpm dev with volume mount), `build` target (pnpm build), `prod` target (standalone Next.js output on minimal base image)
- [ ] T031 [US4] Create `apps/web/.dockerignore` with `node_modules/`, `.next/`, `.env`, `*.md`, `.git`
- [ ] T032 [US4] Rewrite the root `docker-compose.yml` — update `api-dev` and `api-prod` services to use `build.context: ./apps/api`, add `web-dev` and `web-prod` services using `build.context: ./apps/web`, update volume mounts to reference `apps/api/` and `apps/web/` respectively, add `node-modules` volume
- [ ] T033 [US4] Add `dev` and `docker-up` and `docker-down` tasks to the root `Taskfile.yml` — `dev` runs `docker compose --profile dev up --build`, `docker-up` runs `docker compose --profile prod up --build -d`, `docker-down` runs `docker compose down -v`
- [ ] T034 [US4] Verify dev mode: `docker compose --profile dev up --build` starts all services — API health check passes (`curl localhost:8080/healthz`), web serves pages (`curl localhost:3000`), hot reload works for both API and web
- [ ] T035 [US4] Verify service communication: the web container can reach the API container via the Docker network service name (e.g., `http://api-dev:8080`)

**Checkpoint**: Full stack runs in Docker with hot reload. API and web communicate.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final cleanup, validation, and documentation polish.

- [ ] T036 [P] Verify `go.work` enables root-level Go tooling: `go build ./apps/api/...` and `go test ./apps/api/...` work from the repo root
- [ ] T037 [P] Verify workspace independence: delete `apps/web/node_modules` and confirm `cd apps/api && task test` still passes; delete `apps/api/bin/` and confirm `cd apps/web && task test` still passes
- [ ] T038 Run the full quickstart.md validation — execute all 7 scenarios and verify they pass
- [ ] T039 Final review: verify no stale files remain at the repo root (no orphaned Go files, no old Dockerfile/Taskfile), confirm `.gitignore` covers all generated artifacts in both workspaces

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2 — makes the API workspace self-contained
- **US2 (Phase 4)**: Depends on Phase 2 — can run in parallel with US1
- **US3 (Phase 5)**: Depends on US1 and US2 — needs both workspace Taskfiles to exist
- **US4 (Phase 6)**: Depends on US1 and US2 — needs both workspace Dockerfiles to exist
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: Depends on Foundational only — no dependency on other stories
- **US2 (P2)**: Depends on Foundational only — can run in parallel with US1
- **US3 (P3)**: Depends on US1 + US2 — needs both workspace Taskfiles
- **US4 (P4)**: Depends on US1 + US2 — needs both workspace Dockerfiles

### Within Each User Story

- Tasks within a story generally run sequentially (each builds on the prior)
- Tasks marked [P] can run in parallel with other [P] tasks in the same phase

### Parallel Opportunities

- **US1 and US2 can run in parallel** after Phase 2 completes
- Phase 7 tasks marked [P] (T036, T037) can run in parallel

---

## Parallel Example: After Phase 2 Completes

```bash
# Agent A: User Story 1 (API workspace self-containment)
Agent: "Create apps/api/Taskfile.yml"       # T009
Agent: "Update apps/api/Dockerfile"          # T010
Agent: "Move README to apps/api/README.md"   # T011

# Agent B: User Story 2 (Web frontend scaffold) — runs in parallel
Agent: "Scaffold Next.js in apps/web/"       # T017
Agent: "Configure Vitest"                    # T020
Agent: "Create apps/web/Taskfile.yml"        # T023
```

---

## Implementation Strategy

### MVP First (Phase 1 → 2 → 3)

1. Complete Phase 1: Setup — directories created
2. Complete Phase 2: Foundational — Go code moved, build passes
3. Complete Phase 3: US1 — API workspace is self-contained
4. **STOP and VALIDATE**: `cd apps/api && task build && task test && task lint` all pass
5. The repository is now a monorepo with the API workspace functional

### Incremental Delivery

1. Setup + Foundational → File migration complete
2. Add US1 → API workspace self-contained (MVP)
3. Add US2 → Web frontend scaffold (development can begin on both)
4. Add US3 → Root orchestration (single commands for everything)
5. Add US4 → Docker full stack (integration environment)
6. Polish → Validation and cleanup

### Parallel Team Strategy

With multiple developers after Foundational is complete:

- **Developer A**: US1 (API workspace) → US3 (Root orchestration)
- **Developer B**: US2 (Web frontend) → US4 (Docker Compose)

---

## Notes

- [P] tasks = different files, no dependencies on incomplete tasks
- [Story] label maps task to specific user story for traceability
- Use `git mv` for all file moves to preserve Git history
- Verify Go tests pass after every structural change
- The Go module path `arsen` does NOT change — only the directory containing `go.mod` moves
- Commit after each phase or logical group
- Stop at any checkpoint to validate independently

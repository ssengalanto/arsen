# Quickstart Validation: Monorepo Restructure

## Prerequisites

- Go 1.26+ installed
- Node.js 20+ installed
- pnpm installed (`npm install -g pnpm`)
- Docker and Docker Compose installed
- Taskfile v3 installed (`go install github.com/go-task/task/v3/cmd/task@latest`)

## Scenario 1: API Workspace Builds and Tests Pass

**Goal**: Verify the Go API is fully functional after the move to `apps/api/`.

```bash
cd apps/api
task build
task test
```

**Expected outcome**:
- Binary builds successfully at `apps/api/bin/arsen`
- All existing Go tests pass (same count as before restructure)
- No import errors or missing dependencies

## Scenario 2: Web Workspace Starts and Serves Pages

**Goal**: Verify the Next.js scaffold is properly configured.

```bash
cd apps/web
pnpm install
task dev
```

**Expected outcome**:
- Dependencies install without errors
- Next.js dev server starts on port 3000
- Visiting `http://localhost:3000` shows a page
- No TypeScript compilation errors

## Scenario 3: Web Tests Run with Vitest

**Goal**: Verify the testing framework is configured.

```bash
cd apps/web
task test
```

**Expected outcome**:
- Vitest runs and exits cleanly
- At least one sample test passes

## Scenario 4: Root Taskfile Orchestration

**Goal**: Verify top-level commands work across both workspaces.

```bash
# From repo root
task test   # Runs both Go and Vitest tests
task lint   # Runs both golangci-lint and ESLint
task build  # Builds both API binary and Next.js bundle
```

**Expected outcome**:
- All three commands succeed
- Output from both workspaces is visible
- If one fails, the failure is attributed to the correct workspace

## Scenario 5: Docker Compose Full Stack

**Goal**: Verify all services start and communicate.

```bash
# From repo root
# Create .env files from examples first:
cp apps/api/.env.example apps/api/.env
cp apps/web/.env.example apps/web/.env

task dev
```

**Expected outcome**:
- PostgreSQL, Redis, API, and Web containers all start
- API healthcheck passes: `curl http://localhost:8080/healthz` returns 200
- Web server is reachable: `curl http://localhost:3000` returns HTML
- `docker compose down -v` cleans up all containers and volumes

## Scenario 6: Workspace Independence

**Goal**: Verify workspaces don't interfere with each other.

```bash
# Delete web workspace node_modules and verify API still works
rm -rf apps/web/node_modules
cd apps/api && task test   # Should pass

# Delete api binary and verify web still works
rm -rf apps/api/bin
cd apps/web && pnpm install && task dev   # Should start
```

**Expected outcome**:
- Each workspace builds and runs independently
- Missing files in one workspace don't affect the other

## Scenario 7: Go Workspace File

**Goal**: Verify `go.work` enables root-level Go tooling.

```bash
# From repo root
go build ./apps/api/...
go test ./apps/api/...
```

**Expected outcome**:
- Go commands work from the repo root via `go.work`
- IDE (VS Code, GoLand) resolves imports correctly when opening the repo root

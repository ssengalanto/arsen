# Contract: Taskfile Interface

Defines the CLI commands available at the repository root and within each workspace. These are the public interface that developers and CI use.

## Root Taskfile Commands

| Command | Description | Delegates To |
|---------|-------------|-------------|
| `task test` | Run all tests across both workspaces | `api:test` + `web:test` |
| `task lint` | Run all linters across both workspaces | `api:lint` + `web:lint` |
| `task build` | Build both workspaces | `api:build` + `web:build` |
| `task dev` | Start full dev stack via Docker Compose | `docker compose up --build` |
| `task docker-up` | Start production stack | `docker compose --profile prod up --build -d` |
| `task docker-down` | Stop all containers | `docker compose down -v` |
| `task api:*` | Any API workspace task | `apps/api/Taskfile.yml` |
| `task web:*` | Any web workspace task | `apps/web/Taskfile.yml` |

## API Workspace Tasks (`apps/api/Taskfile.yml`)

| Command | Description |
|---------|-------------|
| `task build` | Build the Go binary |
| `task run` | Run the API locally |
| `task test` | Run all Go tests |
| `task test-coverage` | Run tests with coverage report |
| `task lint` | Run golangci-lint |
| `task lint-fix` | Run golangci-lint with auto-fix |
| `task swagger` | Generate Swagger docs |
| `task migrate-up` | Apply pending migrations |
| `task migrate-down` | Rollback last migration |
| `task migrate-create` | Create a new migration pair |
| `task tidy` | Tidy Go modules |

## Web Workspace Tasks (`apps/web/Taskfile.yml`)

| Command | Description |
|---------|-------------|
| `task dev` | Start Next.js dev server |
| `task build` | Build Next.js production bundle |
| `task start` | Start production server |
| `task test` | Run Vitest |
| `task test-watch` | Run Vitest in watch mode |
| `task lint` | Run ESLint |
| `task lint-fix` | Run ESLint with auto-fix |
| `task format` | Run Prettier |
| `task typecheck` | Run TypeScript type checker |

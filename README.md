# Arsen

Monorepo for the Arsen platform: a Go API backend and a Next.js web frontend.

## Structure

```
apps/
  api/    Go backend — JWT auth, vertical slice architecture, CQRS
  web/    Next.js frontend — React, TypeScript, Tailwind CSS
```

## Prerequisites

- Go 1.26+
- Node.js 20+
- [pnpm](https://pnpm.io) (`npm install -g pnpm`)
- Docker and Docker Compose
- [Task](https://taskfile.dev) (go-task/task)

## Getting Started

### API

```bash
cd apps/api
cp .env.example .env
task build
task test
```

### Web

```bash
cd apps/web
cp .env.example .env
pnpm install
task dev
```

### Full Stack (Docker)

```bash
cp apps/api/.env.example apps/api/.env
cp apps/web/.env.example apps/web/.env
task docker-up
```

This starts PostgreSQL, Redis, the API (with hot reload), and the web frontend.

## Root Commands

| Command            | Description                              |
|--------------------|------------------------------------------|
| `task test`        | Run all tests (Go + Vitest)              |
| `task lint`        | Run all linters (golangci-lint + ESLint) |
| `task build`       | Build both workspaces                    |
| `task docker-up`   | Start dev stack via Docker Compose       |
| `task docker-down` | Stop all containers and remove volumes   |

Workspace-specific commands are available via `task api:<command>` and `task web:<command>`. See each workspace's `Taskfile.yml` for details.

## Go Workspace

The root `go.work` file enables Go tooling from the repo root:

```bash
go build ./apps/api/...
go test ./apps/api/...
```

## License

See [LICENSE](LICENSE) for details.

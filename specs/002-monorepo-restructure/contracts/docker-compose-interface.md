# Contract: Docker Compose Services

Defines the services, their build contexts, ports, and dependencies for the multi-service Docker Compose setup.

## Services

### `api-dev` (development)

| Property | Value |
|----------|-------|
| Build context | `./apps/api` |
| Dockerfile | `apps/api/Dockerfile` (target: `dev`) |
| Port | `8080:8080` |
| Volumes | `./apps/api:/app`, `go-modules:/go/pkg/mod` |
| Env file | `./apps/api/.env` |
| Depends on | `postgres` (healthy), `redis` (healthy) |
| Profile | `dev` |

### `web-dev` (development)

| Property | Value |
|----------|-------|
| Build context | `./apps/web` |
| Dockerfile | `apps/web/Dockerfile` (target: `dev`) |
| Port | `3000:3000` |
| Volumes | `./apps/web:/app`, `node-modules:/app/node_modules` |
| Env file | `./apps/web/.env` |
| Depends on | `api-dev` |
| Profile | `dev` |

### `api-prod` (production)

| Property | Value |
|----------|-------|
| Build context | `./apps/api` |
| Dockerfile | `apps/api/Dockerfile` (target: `prod`) |
| Port | `8080:8080` |
| Env file | `./apps/api/.env` |
| Depends on | `postgres` (healthy), `redis` (healthy) |
| Profile | `prod` |

### `web-prod` (production)

| Property | Value |
|----------|-------|
| Build context | `./apps/web` |
| Dockerfile | `apps/web/Dockerfile` (target: `prod`) |
| Port | `3000:3000` |
| Env file | `./apps/web/.env` |
| Depends on | `api-prod` |
| Profile | `prod` |

### `postgres`

| Property | Value |
|----------|-------|
| Image | `postgres:16-alpine` |
| Port | `5432:5432` |
| Volume | `postgres-data:/var/lib/postgresql/data` |
| Healthcheck | `pg_isready -U arsen -d arsen` |

### `redis`

| Property | Value |
|----------|-------|
| Image | `redis:7-alpine` |
| Port | `6379:6379` |
| Volume | `redis-data:/data` |
| Healthcheck | `redis-cli ping` |

## Networking

All services share the default Docker Compose network. The web service reaches the API via the service name `api-dev` (or `api-prod`), e.g., `http://api-dev:8080`.

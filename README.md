# Arsen

Go backend boilerplate with JWT authentication, vertical slice architecture, and CQRS.

## Features

- JWT authentication (HS256) with access/refresh token rotation
- Email verification and password reset flows
- Vertical slice architecture -- each feature is self-contained
- CQRS with Go generics (`pkg/cqrs/`)
- Dependency injection via uber/fx
- PostgreSQL (sqlx + pgx), Redis (go-redis)
- Email delivery via Resend with automatic no-op fallback in dev
- RFC 9457 problem details for all error responses
- Anti-enumeration: register, forgot-password, and resend-verification never leak whether an email exists
- Rate limiting, CORS, request ID, and structured logging middleware
- 12-factor configuration via environment variables
- Docker and Docker Compose for dev (hot reload) and production
- Token cleanup background worker for expired refresh tokens

## Prerequisites

- Go 1.23+
- Docker and Docker Compose
- [Task](https://taskfile.dev) (go-task/task)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI (optional, for manual migrations)

## Quick Start

```bash
git clone <repo-url>
cd golang-boilerplate
cp .env.example .env
task dev
```

This starts PostgreSQL, Redis, and the API with hot reload. Migrations run automatically on startup.

The API listens on `http://localhost:8080` by default.

## API Endpoints

### Health

| Method | Path       | Description              |
|--------|------------|--------------------------|
| GET    | `/healthz` | Liveness probe (always 200) |
| GET    | `/readyz`  | Readiness probe (pings DB)  |

### Auth

| Method | Path                    | Auth     | Description                        |
|--------|-------------------------|----------|------------------------------------|
| POST   | `/api/sessions`         | None     | Login (returns access + refresh tokens) |
| POST   | `/api/tokens`           | None     | Refresh token (rotation)           |
| DELETE  | `/api/sessions/current` | Bearer   | Logout (revokes all refresh tokens) |

### Users

| Method | Path                              | Auth     | Description                |
|--------|-----------------------------------|----------|----------------------------|
| POST   | `/api/users`                      | None     | Register a new account     |
| POST   | `/api/users/verify`               | None     | Verify email with token    |
| POST   | `/api/users/resend-verification`  | None     | Resend verification email  |
| GET    | `/api/users/me`                   | Bearer   | Get current user profile   |
| POST   | `/api/users/forgot-password`      | None     | Request password reset     |
| POST   | `/api/users/reset-password`       | None     | Reset password with token  |

## Curl Examples

Full authentication flow against `localhost:8080`.

**1. Register**

```bash
curl -s -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "password": "s3cureP@ssw0rd"}' | jq
```

```json
{
  "self": "/api/users/550e8400-e29b-41d4-a716-446655440000",
  "kind": "User",
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com",
  "emailVerified": false,
  "createdAt": "2025-01-15T10:30:00Z",
  "message": "Account created. Please check your email to verify your address."
}
```

**2. Verify email**

In dev mode (`ENV=dev`), the verification token is printed to the application logs. Copy the token from the log output.

```bash
curl -s -X POST http://localhost:8080/api/users/verify \
  -H "Content-Type: application/json" \
  -d '{"token": "<token-from-logs>"}' | jq
```

```json
{
  "self": "/api/users/550e8400-e29b-41d4-a716-446655440000",
  "kind": "User",
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com",
  "emailVerified": true,
  "createdAt": "2025-01-15T10:30:00Z",
  "message": "Email verified successfully."
}
```

**3. Login**

```bash
curl -s -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "password": "s3cureP@ssw0rd"}' | jq
```

```json
{
  "self": "/api/sessions/current",
  "kind": "Session",
  "accessToken": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4...",
  "tokenType": "Bearer",
  "expiresIn": 900
}
```

**4. Get profile** (authenticated)

```bash
curl -s http://localhost:8080/api/users/me \
  -H "Authorization: Bearer <accessToken>" | jq
```

```json
{
  "self": "/api/users/me",
  "kind": "User",
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com",
  "emailVerified": true,
  "createdAt": "2025-01-15T10:30:00Z"
}
```

**5. Refresh token**

```bash
curl -s -X POST http://localhost:8080/api/tokens \
  -H "Content-Type: application/json" \
  -d '{"refreshToken": "<refreshToken>"}' | jq
```

```json
{
  "self": "/api/tokens",
  "kind": "TokenPair",
  "accessToken": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "bmV3IHJlZnJlc2ggdG9rZW4...",
  "tokenType": "Bearer",
  "expiresIn": 900
}
```

The old refresh token is revoked. Use the new one for subsequent refreshes.

**6. Logout** (authenticated)

```bash
curl -s -X DELETE http://localhost:8080/api/sessions/current \
  -H "Authorization: Bearer <accessToken>"
```

Returns `204 No Content`. All refresh tokens for the user are revoked.

**7. Forgot password**

```bash
curl -s -X POST http://localhost:8080/api/users/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com"}' | jq
```

```json
{
  "kind": "Acknowledgment",
  "message": "If an account exists with this email, a password reset link has been sent."
}
```

Always returns 200 regardless of whether the email exists (anti-enumeration).

**8. Reset password**

In dev mode, the reset token is printed to the application logs.

```bash
curl -s -X POST http://localhost:8080/api/users/reset-password \
  -H "Content-Type: application/json" \
  -d '{"token": "<reset-token-from-logs>", "password": "n3wS3cure!Pass"}' | jq
```

```json
{
  "kind": "Acknowledgment",
  "message": "Password has been reset successfully. Please log in with your new password."
}
```

## Project Structure

```
cmd/api/              Application entrypoint (uber/fx wiring)
features/
  auth/               Login, token refresh, logout
  user/               Registration, verification, password reset, profile
  health/             Liveness and readiness probes
pkg/
  config/             12-factor config via Viper
  cqrs/               Command/Query/Event buses with Go generics
    middleware/        CQRS middleware (logging, recovery, validation)
  database/           PostgreSQL connection (sqlx + pgx)
  email/              Email sender (Resend + no-op dev fallback)
  jwt/                JWT service (HS256, golang-jwt)
  middleware/         HTTP middleware (auth, CORS, logging, rate limit, request ID)
  redis/              Redis client wrapper
  response/           RFC 9457 problem details + JSON helpers
  server/             Chi router setup with global middleware
  token/              Cryptographic token generation (256-bit)
migrations/           PostgreSQL migrations (golang-migrate)
```

## Task Targets

| Command              | Description                              |
|----------------------|------------------------------------------|
| `task dev`           | Start full dev stack with hot reload     |
| `task build`         | Build the `arsen` binary                 |
| `task run`           | Run the application locally              |
| `task test`          | Run all tests                            |
| `task test-coverage` | Run tests with HTML coverage report      |
| `task lint`          | Run golangci-lint                        |
| `task lint-fix`      | Run golangci-lint with auto-fix          |
| `task swagger`       | Generate Swagger docs                    |
| `task migrate-up`    | Apply all pending migrations             |
| `task migrate-down`  | Rollback the last migration              |
| `task migrate-create`| Create a new migration pair              |
| `task docker-up`     | Start production stack (detached)        |
| `task docker-down`   | Stop all containers and remove volumes   |
| `task tidy`          | Tidy Go modules                          |

## Environment Variables

Copy `.env.example` to `.env` and adjust as needed.

| Variable                 | Default                          | Description                              |
|--------------------------|----------------------------------|------------------------------------------|
| `ENV`                    | `dev`                            | Environment: `dev`, `staging`, `prod`    |
| `SERVER_ADDRESS`         | `:8080`                          | HTTP listen address                      |
| `DATABASE_URL`           | *(required)*                     | PostgreSQL connection string             |
| `REDIS_URL`              | *(required)*                     | Redis connection string                  |
| `JWT_SECRET`             | *(required)*                     | HMAC secret for signing JWTs             |
| `JWT_ISSUER`             | `arsen`                          | JWT `iss` claim                          |
| `JWT_AUDIENCE`           | `arsen-api`                      | JWT `aud` claim                          |
| `ACCESS_TOKEN_DURATION`  | `15m`                            | Access token lifetime                    |
| `REFRESH_TOKEN_DURATION` | `168h`                           | Refresh token lifetime (7 days)          |
| `RESEND_API_KEY`         | *(required in prod)*             | Resend API key for email delivery        |
| `EMAIL_FROM_ADDRESS`     | `noreply@example.com`            | Sender email address                     |
| `EMAIL_FROM_NAME`        | `Arsen`                          | Sender display name                      |
| `APP_BASE_URL`           | `http://localhost:8080`          | Base URL for links in emails             |
| `RATE_LIMIT_REQUESTS`    | `10`                             | Requests per window                      |
| `RATE_LIMIT_BURST`       | `20`                             | Burst size for rate limiter              |
| `LOG_LEVEL`              | `debug`                          | Log level (debug, info, warn, error)     |
| `CLEANUP_INTERVAL`       | `1h`                             | Interval for expired token cleanup       |

In `dev` mode, if `RESEND_API_KEY` is not set, emails are logged to stdout instead of being sent.

## Testing

```bash
task test                # Run all tests
task test-coverage       # Generate HTML coverage report (coverage.html)
go test ./... -run TestLogin -v   # Run a specific test
```

## Architecture Notes

**Vertical Slice Architecture** -- Each feature (auth, user, health) is a self-contained package with its own handler, commands/queries, repository, and fx module. No shared domain layer forces cross-feature coupling.

**CQRS with Go Generics** -- `pkg/cqrs/` provides type-safe `CommandBus[C, R]` and `QueryBus[Q, R]` with middleware support (logging, recovery, validation). Commands mutate state, queries read it. Event publishing is available for cross-feature communication (e.g., sending a welcome email after registration).

**Dependency Injection** -- uber/fx wires everything in `cmd/api/main.go`. Each package exposes an `fx.Module` that provides its constructors. The application starts by composing all modules.

**Token Rotation with Family Model** -- Refresh tokens use a family ID. When a token is refreshed, the entire family is revoked and a new token is issued under the same family. If a revoked token is reused, the entire family is invalidated, detecting token theft.

**Anti-Enumeration** -- Registration normalizes conflict errors to generic bad-request responses. Forgot-password and resend-verification always return 200 with an identical message regardless of whether the email exists.

**RFC 9457 Problem Details** -- All error responses use `application/problem+json` with structured `type`, `title`, `status`, `detail`, `instance`, and optional `errors` array for field-level validation failures.

**Docker** -- Multi-stage Dockerfile: `dev` target uses Air for hot reload with mounted source; `prod` target produces a minimal distroless image running as non-root. Docker Compose manages PostgreSQL 16, Redis 7, and the API with health checks.

## License

See [LICENSE](LICENSE) for details.

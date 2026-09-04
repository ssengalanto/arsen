# Implementation Plan: Go Backend Boilerplate with JWT Authentication

**Branch**: `001-go-jwt-auth-boilerplate` | **Date**: 2026-09-04 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/001-go-jwt-auth-boilerplate/spec.md`

## Summary

Production-grade Go backend boilerplate implementing JWT authentication with email verification, password reset, and welcome emails via Resend. Uses Vertical Slice Architecture with CQRS (in-house, generics-based), SOLID principles, and uber/fx dependency injection. PostgreSQL for persistence, Redis interface seam for future caching, chi for routing, and full 12-factor compliance. Registration gates login behind email verification; password reset revokes all refresh token families.

## Technical Context

**Language/Version**: Go (latest stable, 1.23+)

**Primary Dependencies**:
- HTTP routing: go-chi/chi/v5
- JWT: golang-jwt/jwt/v5
- Validation: go-playground/validator/v10
- Rate limiting: golang.org/x/time/rate
- Config: spf13/viper
- DI: uber-go/fx
- DB: jmoiron/sqlx + jackc/pgx/v5/stdlib
- Cache: go-redis/redis/v9 (interface seam only)
- Email: resend/resend-go/v2
- Migrations: golang-migrate/migrate/v4
- Swagger: swaggo/swag + swaggo/http-swagger
- Testing: stretchr/testify
- Utilities: samber/lo
- Dev reload: cosmtrek/air
- Linting: golangci-lint

**Storage**: PostgreSQL (primary, via sqlx + pgx/v5 stdlib driver), Redis (connection + interface seam, no cache logic)

**Testing**: `go test` + stretchr/testify (assert, require, mock). Table-driven tests. Integration tests against real PostgreSQL.

**Target Platform**: Linux server (Docker containers), macOS for development

**Project Type**: Web service (REST API)

**Performance Goals**: Sub-second startup (FR-042). No explicit throughput target — boilerplate is optimized for correctness, security, and developer ergonomics.

**Constraints**: Stateless process (12-factor), single-instance rate limiter (interface seam for Redis upgrade), HS256 JWT signing, bcrypt cost >= 12

**Scale/Scope**: Auth boilerplate with ~13 endpoints across 3-4 vertical slices. Designed for team reuse and extension.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Constitution file is an unfilled template — no project-specific principles or gates are defined. No violations possible. Proceeding to Phase 0.

**Post-Phase 1 Re-check**: Same — no gates defined.

## Project Structure

### Documentation (this feature)

```text
specs/001-go-jwt-auth-boilerplate/
├── plan.md              # This file
├── research.md          # Phase 0 output - technology decisions and rationale
├── data-model.md        # Phase 1 output - entity definitions and relationships
├── quickstart.md        # Phase 1 output - validation scenarios
├── contracts/           # Phase 1 output - API endpoint contracts
│   ├── auth.md          # Login, logout, token refresh
│   ├── users.md         # Registration, profile, verification, password reset
│   └── health.md        # Healthz, readyz
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (/speckit-tasks command)
```

### Source Code (repository root)

```text
cmd/
└── api/
    └── main.go                    # Application entry point, fx.New()

features/
├── auth/                          # Auth vertical slice
│   ├── module.go                  # fx.Module registration
│   ├── handler.go                 # HTTP handlers (login, logout, refresh)
│   ├── login_command.go           # LoginCommand + handler
│   ├── logout_command.go          # LogoutCommand + handler
│   ├── refresh_command.go         # RefreshTokenCommand + handler
│   └── repository.go             # Session/token repository interface + impl
├── user/                          # User vertical slice
│   ├── module.go                  # fx.Module registration
│   ├── handler.go                 # HTTP handlers (register, profile, verify, reset)
│   ├── register_command.go        # RegisterCommand + handler
│   ├── verify_email_command.go    # VerifyEmailCommand + handler
│   ├── resend_verification_command.go
│   ├── forgot_password_command.go
│   ├── reset_password_command.go
│   ├── get_profile_query.go       # GetProfileQuery + handler
│   └── repository.go             # User repository interface + impl
└── health/                        # Health vertical slice
    ├── module.go                  # fx.Module registration
    ├── handler.go                 # HTTP handlers (healthz, readyz)
    └── readiness_query.go         # ReadinessQuery + handler

pkg/
├── cqrs/                          # In-house CQRS infrastructure
│   ├── command.go                 # CommandBus, CommandHandler, middleware
│   ├── query.go                   # QueryBus, QueryHandler, middleware
│   ├── event.go                   # EventBus, EventHandler (sync/async)
│   ├── middleware/
│   │   ├── logging.go
│   │   ├── validation.go
│   │   └── recovery.go
│   └── errors.go                  # Domain error types (NotFound, Conflict, etc.)
├── config/                        # Viper-based configuration
│   ├── config.go                  # Typed config structs
│   └── module.go                  # fx.Module
├── database/                      # PostgreSQL connection
│   ├── database.go                # sqlx connection setup
│   └── module.go                  # fx.Module
├── redis/                         # Redis connection + cache interface seam
│   ├── redis.go
│   └── module.go                  # fx.Module
├── email/                         # Resend email service
│   ├── email.go                   # EmailSender interface + Resend impl
│   ├── noop.go                    # No-op sender for dev mode
│   └── module.go                  # fx.Module
├── token/                         # Token generation utilities
│   └── token.go                   # Secure random token generation, hashing
├── server/                        # HTTP server setup
│   ├── server.go                  # chi router, middleware composition
│   └── module.go                  # fx.Module
├── middleware/                     # Shared HTTP middleware
│   ├── auth.go                    # JWT auth middleware
│   ├── ratelimit.go               # Per-IP rate limiter
│   ├── requestid.go               # Request/correlation ID
│   ├── logging.go                 # Structured request logging
│   └── cors.go                    # CORS configuration
├── response/                      # HTTP response helpers
│   ├── json.go                    # JSON response writer with self/kind
│   └── problem.go                 # RFC 9457 problem details
└── cleanup/                       # Background token cleanup
    ├── cleanup.go                 # Goroutine for purging expired tokens
    └── module.go                  # fx.Module

migrations/
├── 000001_create_users.up.sql
├── 000001_create_users.down.sql
├── 000002_create_refresh_tokens.up.sql
├── 000002_create_refresh_tokens.down.sql
├── 000003_create_verification_tokens.up.sql
├── 000003_create_verification_tokens.down.sql
├── 000004_create_password_reset_tokens.up.sql
└── 000004_create_password_reset_tokens.down.sql

.air.toml                          # Air live reload config
.golangci.yml                      # golangci-lint config
.env.example                       # Example env vars
docker-compose.yml                 # Full stack: API + Postgres + Redis
Dockerfile                         # Multi-target: dev + prod
Taskfile.yml                       # Task runner targets
```

**Structure Decision**: Go web service using Vertical Slice Architecture. Features are organized under `features/` with each slice owning its handlers, commands/queries, and repository. Shared infrastructure lives in `pkg/`. The `cmd/api/` entry point composes the fx dependency graph. Migrations are top-level SQL files managed by golang-migrate.

## Complexity Tracking

> No constitution violations to justify — constitution is unfilled template.

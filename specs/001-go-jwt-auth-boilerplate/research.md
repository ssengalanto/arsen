# Research: Go Backend Boilerplate with JWT Authentication

**Phase 0 Output** | **Date**: 2026-09-04

All technology choices were explicitly specified in the feature spec clarifications. This document records each decision with rationale and alternatives considered.

## HTTP Routing

- **Decision**: go-chi/chi/v5
- **Rationale**: Lightweight, idiomatic Go, stdlib net/http compatible. Supports middleware chaining, route grouping, and URL parameters natively. No reflection or code generation required.
- **Alternatives considered**: gorilla/mux (archived/maintenance mode), gin (non-stdlib interface, heavier), stdlib net/http ServeMux (insufficient for middleware composition and route grouping at scale)

## JWT Library

- **Decision**: golang-jwt/jwt/v5
- **Rationale**: Community-maintained successor to dgrijalva/jwt-go. Supports HS256, RS256, custom claims, and claim validation. Well-tested and widely adopted.
- **Alternatives considered**: lestrrat-go/jwx (more comprehensive but heavier for HS256-only use case), go-jose (JOSE-focused, overkill for simple JWT)

## Validation

- **Decision**: go-playground/validator/v10
- **Rationale**: Struct tag-based validation, extensible with custom validators, widely adopted. Integrates naturally with Go structs used in request parsing.
- **Alternatives considered**: ozzo-validation (code-based rather than tag-based, more verbose), manual validation (error-prone, inconsistent)

## Rate Limiting

- **Decision**: golang.org/x/time/rate (token bucket)
- **Rationale**: Stdlib-adjacent, no external dependencies. Token bucket algorithm is well-suited for per-IP rate limiting. Simple API.
- **Alternatives considered**: ulule/limiter (more features but external dependency), go-redis/redis_rate (Redis-backed, premature for single-instance)

## Configuration

- **Decision**: spf13/viper
- **Rationale**: Mature, supports env vars, .env files, typed structs. 12-factor compliant when configured for env var binding. Widely used in Go ecosystem.
- **Alternatives considered**: kelseyhightower/envconfig (lighter but no .env file support), caarlos0/env (simpler but less featured), manual os.Getenv (tedious, no validation)

## Dependency Injection

- **Decision**: uber-go/fx
- **Rationale**: Runtime DI with module system. Supports lifecycle hooks (OnStart/OnStop) for server, DB, cleanup goroutines. Forces explicit dependency declaration. Ideal for composing vertical slices as independent modules.
- **Alternatives considered**: google/wire (compile-time, no lifecycle hooks), manual construction (doesn't scale with module count), samber/do (less mature)

## Database Access

- **Decision**: jmoiron/sqlx + jackc/pgx/v5/stdlib
- **Rationale**: sqlx provides struct scanning, named parameters, and Get/Select convenience over raw database/sql. pgx/v5 stdlib driver is the most performant and well-maintained PostgreSQL driver for Go. sqlx bridges the ergonomic gap without an ORM.
- **Alternatives considered**: GORM (too opinionated, hides SQL), pgx native (no struct scanning), ent (codegen ORM, overkill for this scope), raw database/sql (too verbose)

## Caching Infrastructure

- **Decision**: go-redis/redis/v9 — interface seam only, no cache logic
- **Rationale**: Establishing the Redis connection and cache interface upfront means caching can be added to any layer without structural changes. No premature optimization — actual cache usage deferred until profiling identifies need.
- **Alternatives considered**: In-memory cache (e.g., patrickmn/go-cache — not distributed), skip Redis entirely (loses the interface seam for future scaling)

## Email Sending

- **Decision**: resend/resend-go/v2
- **Rationale**: Resend provides a modern, developer-friendly transactional email API. The Go SDK is thin and idiomatic. Wrapping behind an interface enables provider swaps and test mocking.
- **Alternatives considered**: AWS SES (heavier SDK, more config), SendGrid (older API design), SMTP direct (infrastructure burden, deliverability challenges), mailgun (similar to Resend but less developer-focused)

## Email Service Architecture

- **Decision**: Interface-based email sender with Resend implementation and no-op dev fallback
- **Rationale**: `EmailSender` interface allows tests to mock email sending, dev mode to skip it, and production to use Resend. Follows Dependency Inversion (FR-063) and Interface Segregation (FR-062). The no-op sender logs email details (without tokens) so developers can extract verification/reset tokens from logs during development.
- **Alternatives considered**: Direct Resend client usage (untestable, couples business logic to provider), event-driven async-only (adds complexity, verification email needs to be somewhat synchronous with registration)

## Token Storage Pattern (Verification + Password Reset)

- **Decision**: 256-bit random tokens, stored as SHA-256 hashes, same pattern as refresh tokens
- **Rationale**: Consistent security model across all token types. SHA-256 hashing means database compromise doesn't expose usable tokens. 256-bit entropy makes brute force infeasible. Single-use + expiry limits attack window.
- **Alternatives considered**: JWT-based tokens (stateless but not revocable without blocklist, defeats purpose), UUID v4 (less entropy, 122 bits), storing raw tokens (database compromise = full access)

## Password Hashing

- **Decision**: bcrypt with cost >= 12
- **Rationale**: Mature, well-understood, adaptive cost factor. Cost 12 provides ~250ms hash time on modern hardware, balancing security with UX. Go stdlib support via golang.org/x/crypto/bcrypt.
- **Alternatives considered**: argon2id (stronger memory-hardness but more complex tuning, less ecosystem tooling), scrypt (less widely adopted in Go ecosystem)

## Migrations

- **Decision**: golang-migrate/migrate/v4
- **Rationale**: Versioned up/down SQL file pairs. Database-agnostic driver model. CLI and library modes. Integrates cleanly with Taskfile targets. No code generation.
- **Alternatives considered**: goose (similar but less active), atlas (schema-based rather than migration-based, different paradigm), manual SQL (no version tracking)

## Swagger / API Docs

- **Decision**: swaggo/swag + swaggo/http-swagger
- **Rationale**: Generates OpenAPI spec from Go source annotations. Serves Swagger UI at configurable path. Taskfile target regenerates docs. Annotations live next to handlers.
- **Alternatives considered**: go-swagger (heavier, code generation approach), manual OpenAPI YAML (drifts from code), grpc-gateway (wrong paradigm for REST)

## Testing Strategy

- **Decision**: go test + stretchr/testify, table-driven tests, real PostgreSQL for integration
- **Rationale**: testify provides fluent assertions (assert for non-fatal, require for preconditions) and mock package. Table-driven tests are idiomatic Go. Real PostgreSQL avoids mock/prod divergence.
- **Alternatives considered**: gomock (more ceremony), dockertest for test containers (good complement, may add for integration tests), pure stdlib testing (verbose assertions)

## Utilities

- **Decision**: samber/lo
- **Rationale**: Generic utility functions (Map, Filter, Contains, etc.) reduce boilerplate loops. Cleaner than hand-rolled generics for common operations.
- **Alternatives considered**: Hand-rolled utility functions (reinventing the wheel), no utilities (verbose loops everywhere)

## CQRS Package Design

- **Decision**: In-house, generics-based, zero external dependencies
- **Rationale**: Improved version of github.com/ssengalanto/axiaos CQRS package. Key improvements: (1) CommandHandler[C, R] returns results via two type parameters, (2) no marker interfaces — `any` constraint, (3) Validatable constraint checked at compile time, (4) structured domain errors mapping to RFC 9457. EventBus supports sync and async dispatch modes.
- **Alternatives considered**: mediator pattern libraries (less Go-idiomatic), no CQRS (business logic bleeds into handlers), external CQRS packages (none match the exact API design needed)

## Project Layout

- **Decision**: Vertical Slice Architecture under `features/`, shared infra in `pkg/`
- **Rationale**: Each feature (auth, user, health) is self-contained with its own handlers, commands/queries, and repository. Shared infrastructure (database, config, email, CQRS, middleware) lives in `pkg/`. New features require only a new directory + fx.Module registration.
- **Alternatives considered**: Clean Architecture layers (too many abstraction layers for a boilerplate), flat package structure (doesn't scale), domain-driven design (overkill for auth-only scope)

## Dev Tooling

- **Decision**: Taskfile (go-task.dev) + Air (cosmtrek/air) + Docker Compose + golangci-lint
- **Rationale**: Taskfile is simpler and more readable than Makefile, with cross-platform support. Air provides Go-specific live reload watching .go files. Docker Compose runs the full stack. golangci-lint aggregates multiple linters with a single config.
- **Alternatives considered**: Makefile (less readable, platform quirks), nodemon/watchexec (not Go-specific), manual docker run (tedious)

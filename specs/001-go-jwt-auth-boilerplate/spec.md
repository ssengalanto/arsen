# Feature Specification: Go Backend Boilerplate with JWT Authentication

**Feature Directory**: `specs/001-go-jwt-auth-boilerplate`

**Created**: 2026-09-04

**Status**: Draft

**Input**: User description: "Production-grade Go backend boilerplate using Vertical Slice Architecture, SOLID principles, and complete JWT authentication with TDD"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - New User Registration (Priority: P1)

A new user visits the service and creates an account by providing their email and password. The system validates the input, ensures the email is not already taken, securely hashes the password, creates the account with an unverified email status, and sends a verification email via Resend. The user must verify their email before they can log in.

**Why this priority**: Registration is the entry point for all users. Without it, no other auth flows are possible. It also establishes the core data model (users table) that every other story depends on.

**Independent Test**: Can be fully tested by sending a POST request with email/password and verifying a user record is created with `emailVerified = false`, a verification email is dispatched, and the response confirms the account was created with instructions to check email.

**Acceptance Scenarios**:

1. **Given** no account exists for "alice@example.com", **When** a user registers with a valid email and a strong password, **Then** the system creates the account with `emailVerified = false`, sends a verification email, and returns a 201 response indicating the user should check their email.
2. **Given** an account already exists for "alice@example.com", **When** a user tries to register with the same email, **Then** the system returns an error that is indistinguishable from a generic registration failure (no user enumeration).
3. **Given** a registration request with an invalid email or a missing password, **When** submitted, **Then** the system returns field-level validation errors in RFC 9457 problem details format.
4. **Given** a burst of registration requests from the same source, **When** the rate limit is exceeded, **Then** the system returns 429 Too Many Requests.

---

### User Story 2 - User Login (Priority: P1)

An existing user logs in with their email and password. The system verifies credentials and issues a short-lived access token and a long-lived opaque refresh token.

**Why this priority**: Login is the primary authentication flow and is required before any protected resource can be accessed. It is co-equal with registration as the foundation of the auth feature.

**Independent Test**: Can be tested by first creating a user (via registration or direct DB insert), then sending a login request and verifying that valid tokens are returned.

**Acceptance Scenarios**:

1. **Given** a registered user with a verified email "alice@example.com", **When** they log in with correct credentials, **Then** the system returns an access token (short-lived, ~15 min) and a refresh token (long-lived, ~7 days).
2. **Given** no account exists for "bob@example.com", **When** someone attempts to log in with that email, **Then** the system returns the same error response and takes comparable time as a wrong-password attempt (no user enumeration).
3. **Given** a registered user, **When** they log in with the wrong password, **Then** the system returns a generic authentication failure identical to the "user not found" case.
4. **Given** a burst of login attempts from the same source, **When** the rate limit is exceeded, **Then** the system returns 429 Too Many Requests.
5. **Given** a registered user whose email is not yet verified, **When** they attempt to log in with correct credentials, **Then** the system returns 403 Forbidden with a message indicating the email must be verified before login is allowed.

---

### User Story 3 - Access Protected Resources (Priority: P1)

An authenticated user accesses protected endpoints (e.g., GET /api/users/me) by presenting a valid access token in the Authorization header. The system validates the token and injects the authenticated user's identity into the request context.

**Why this priority**: This proves the auth middleware works end-to-end and is the core value proposition of the entire auth feature: gating access to resources.

**Independent Test**: Can be tested by obtaining an access token via login, then calling /users/me and verifying the correct user profile is returned.

**Acceptance Scenarios**:

1. **Given** a valid access token, **When** the user calls GET /api/users/me, **Then** the system returns the authenticated user's profile (including self, kind, id, email, emailVerified, createdAt).
2. **Given** no Authorization header, **When** a request is made to a protected endpoint, **Then** the system returns 401 Unauthorized.
3. **Given** an expired access token, **When** a request is made, **Then** the system returns 401 Unauthorized.
4. **Given** a tampered or malformed token, **When** a request is made, **Then** the system returns 401 Unauthorized.

---

### User Story 4 - Token Refresh with Rotation (Priority: P2)

A user whose access token has expired uses their refresh token to obtain a new access token and a new refresh token. The old refresh token is invalidated (rotated). If a previously-rotated refresh token is reused, the system detects the reuse and revokes the entire token family.

**Why this priority**: Token refresh enables long-lived sessions without requiring frequent re-login, which is critical for user experience. Reuse detection is a key security requirement.

**Independent Test**: Can be tested by logging in, using the refresh token, verifying new tokens are issued and the old refresh token no longer works. Reuse detection can be tested by replaying an already-rotated token and verifying the entire family is revoked.

**Acceptance Scenarios**:

1. **Given** a valid, unused refresh token, **When** the user calls POST /api/tokens, **Then** the system returns a new access token and a new refresh token, and the old refresh token is invalidated.
2. **Given** an expired refresh token, **When** the user attempts to refresh, **Then** the system returns 401 Unauthorized.
3. **Given** a refresh token that has already been rotated (reuse attempt), **When** presented again, **Then** the system revokes all tokens in that family and returns 401 Unauthorized.
4. **Given** a completely invalid or random string as a refresh token, **When** submitted, **Then** the system returns 401 Unauthorized.

---

### User Story 5 - Logout (Priority: P2)

An authenticated user logs out, which revokes their entire refresh token family. Subsequent attempts to refresh using any token from that family fail.

**Why this priority**: Logout is essential for security (stolen token mitigation) but depends on the token family model established by login and refresh.

**Independent Test**: Can be tested by logging in, calling logout, and then verifying that refresh attempts with the original token fail.

**Acceptance Scenarios**:

1. **Given** an authenticated user with a valid access token, **When** they call DELETE /api/sessions/current, **Then** the system revokes the entire refresh token family and returns success.
2. **Given** a user who has already logged out, **When** they call logout again with the same (now invalid) token, **Then** the system returns success (idempotent).
3. **Given** a revoked token family, **When** a refresh is attempted with any token from that family, **Then** the system returns 401 Unauthorized.

---

### User Story 6 - Health and Readiness Checks (Priority: P3)

Operations tooling checks whether the service is alive (/healthz) and whether it is ready to serve traffic (/readyz). The liveness check has no dependencies; the readiness check pings the database.

**Why this priority**: Operational necessity for deployment, but not core user-facing functionality.

**Independent Test**: Can be tested by calling GET /healthz and verifying 200, and calling GET /readyz and verifying it reflects actual DB connectivity.

**Acceptance Scenarios**:

1. **Given** the server is running, **When** GET /healthz is called, **Then** the system returns 200 OK with no dependency checks.
2. **Given** the database is connected, **When** GET /readyz is called, **Then** the system returns 200 OK.
3. **Given** the database is unreachable, **When** GET /readyz is called, **Then** the system returns 503 Service Unavailable.

---

### User Story 7 - Developer Onboarding (Priority: P3)

A new developer clones the repository and gets from zero to a working POST /api/sessions (login) call in under five minutes using the README, Taskfile, and docker-compose.

**Why this priority**: Critical for the boilerplate's purpose (team reuse) but is a developer experience concern, not a runtime feature.

**Independent Test**: Can be tested by following the README steps from a clean clone and timing the process to first successful POST /api/sessions response.

**Acceptance Scenarios**:

1. **Given** a developer has cloned the repo and has Docker and Task installed, **When** they run `docker compose up` (or `task dev`), **Then** the API server and all dependencies start and accept requests within 5 minutes total setup time.
2. **Given** the server is running, **When** the developer follows the README's quick-start curl examples, **Then** they can register (POST /api/users), verify email (POST /api/users/verify — using the token from the dev email log or Resend dashboard), login (POST /api/sessions), and call GET /api/users/me successfully.

---

### User Story 8 - Email Verification (Priority: P1)

After registering, a user receives a verification email containing a unique link. Clicking the link (or submitting the token via API) verifies their email address, enabling login. The system sends a welcome email upon successful verification.

**Why this priority**: Email verification is a prerequisite for login (gating mechanism). Without it, newly registered users cannot authenticate. It is co-equal with registration as a P1 requirement.

**Independent Test**: Can be tested by registering a user, extracting the verification token (from the database or email mock), calling the verify endpoint, and confirming `emailVerified` transitions to `true` and a welcome email is dispatched.

**Acceptance Scenarios**:

1. **Given** a newly registered user with an unverified email, **When** they submit a valid verification token to `POST /api/users/verify`, **Then** the system marks the email as verified, sends a welcome email, and returns success.
2. **Given** a verification token that has expired (>24 hours old), **When** submitted, **Then** the system returns 401 Unauthorized with a generic error (no detail about expiry vs. invalid).
3. **Given** a verification token that has already been used, **When** submitted again, **Then** the system returns 401 Unauthorized.
4. **Given** a completely invalid or random token string, **When** submitted, **Then** the system returns 401 Unauthorized.
5. **Given** a registered user whose verification token has expired, **When** they call `POST /api/users/resend-verification` with their email, **Then** the system invalidates any existing verification tokens, generates a new one, and sends a new verification email.
6. **Given** a burst of resend-verification requests from the same source, **When** the rate limit is exceeded, **Then** the system returns 429 Too Many Requests.
7. **Given** an email that is not registered, **When** a resend-verification request is submitted for that email, **Then** the system returns the same success response as a valid request (no user enumeration).

---

### User Story 9 - Password Reset (Priority: P2)

A user who has forgotten their password requests a reset link via email. The system sends a password reset email with a time-limited token. The user submits the token along with a new password to regain access. All existing refresh token families are revoked upon successful reset.

**Why this priority**: Password reset is essential for account recovery but depends on the email infrastructure and user model established by registration and verification.

**Independent Test**: Can be tested by creating a verified user, requesting a password reset, extracting the reset token, submitting a new password, and verifying the old password no longer works, the new password works, and all refresh token families are revoked.

**Acceptance Scenarios**:

1. **Given** a verified user with email "alice@example.com", **When** they call `POST /api/users/forgot-password` with their email, **Then** the system sends a password reset email and returns 200 OK.
2. **Given** an email that is not registered, **When** a forgot-password request is submitted, **Then** the system returns the same 200 OK response and takes comparable time (no user enumeration).
3. **Given** a valid, unused password reset token, **When** the user calls `POST /api/users/reset-password` with the token and a new valid password, **Then** the system updates the password hash, revokes all refresh token families for that user, and returns success.
4. **Given** a password reset token that has expired (>1 hour old), **When** submitted, **Then** the system returns 401 Unauthorized with a generic error.
5. **Given** a password reset token that has already been used, **When** submitted again, **Then** the system returns 401 Unauthorized.
6. **Given** a reset-password request with a weak new password, **When** submitted, **Then** the system returns field-level validation errors in RFC 9457 problem details format.
7. **Given** a burst of forgot-password requests from the same source, **When** the rate limit is exceeded, **Then** the system returns 429 Too Many Requests.

---

### User Story 10 - Welcome Email (Priority: P2)

After a user successfully verifies their email, the system sends a welcome email to acknowledge their registration and provide a starting point for using the service.

**Why this priority**: Welcome emails improve user experience and confirm that the account is fully active. It depends on verification being complete.

**Independent Test**: Can be tested by verifying a user's email and confirming that a welcome email is dispatched (via email mock or Resend API logs).

**Acceptance Scenarios**:

1. **Given** a user has just successfully verified their email, **When** the verification completes, **Then** the system sends a welcome email to the user's address.
2. **Given** the email service (Resend) is temporarily unavailable, **When** the welcome email fails to send, **Then** the verification still succeeds (welcome email is best-effort), and the failure is logged with a correlation ID.

---

### Edge Cases

- What happens when the database connection is lost mid-request? The system returns a 500 error with a correlation ID; no internal details are leaked.
- What happens when a JWT secret is missing or set to a default value in production? The application refuses to start and logs a clear error message.
- What happens when concurrent refresh requests use the same token? Only one succeeds; the others trigger reuse detection and family revocation.
- What happens when a user registers with mixed-case email (e.g., "Alice@Example.com")? The system normalizes via citext and treats it as the same account.
- What happens when the access token contains a valid structure but is signed with a different secret? The system rejects it as tampered (401).
- What happens when the Resend API key is missing or invalid in production? The application refuses to start and logs a clear error message (same pattern as JWT secret).
- What happens when the Resend API is temporarily down during registration? The user record is still created, but the verification email fails. The failure is logged with a correlation ID. The user can request a resend via POST /api/users/resend-verification.
- What happens when a user requests multiple verification tokens in quick succession? Each new request invalidates all previous verification tokens for that user. Only the latest token is valid.
- What happens when a user requests a password reset while their email is unverified? The system returns the same generic 200 OK response (no enumeration), but does not send a reset email since the email is unverified.
- What happens when concurrent password reset requests arrive for the same user? Each request invalidates all previous reset tokens. Only the latest token is valid.
- What happens when the welcome email fails to send after successful verification? The verification still succeeds. The welcome email failure is logged but does not affect the user's ability to log in.

## Clarifications

### Session 2026-09-04

- Q: What minimum password length and complexity rules should the system enforce during registration? → A: Minimum 8 characters, must include at least one uppercase letter, one lowercase letter, one digit, and one special character.
- Q: What rate limit values should apply to the /login and /register endpoints? → A: 10 requests per minute per IP address.
- Q: How should expired and revoked refresh tokens be cleaned up from the database? → A: Background goroutine that purges expired/revoked tokens on a configurable interval (e.g., hourly).
- Q: What values should the access token's iss and aud claims contain? → A: Both configurable via environment variables, with defaults (iss={{PROJECT_NAME}}, aud={{PROJECT_NAME}}-api).
- Q: Should the system use Go's stdlib log/slog or a third-party library for structured logging? → A: log/slog (stdlib) with JSON handler — zero dependencies, idiomatic Go.
- Q: Should JSON responses include self-referencing URLs and resource type identifiers per Apigee API design best practices? → A: Yes, all JSON responses MUST include `self` (resource URL) and `kind` (resource type) properties.
- Q: What HTTP status and headers should successful resource creation return? → A: 201 Created with a Location header pointing to the new resource URL.
- Q: What JSON property naming convention should the API use? → A: camelCase (e.g., `createdAt`, `refreshToken`).
- Q: What date/time format should the API use in JSON responses? → A: ISO 8601 (e.g., `2026-09-04T12:30:00Z`).
- Q: Should auth endpoints use noun-based resource URLs instead of verb-based paths? → A: Yes, per Apigee "nouns are good, verbs are bad" principle. Registration → POST /api/users, Login → POST /api/sessions, Refresh → POST /api/tokens, Logout → DELETE /api/sessions/current, Profile → GET /api/users/me.
- Q: Should the API use header-based versioning instead of URL path versioning? → A: No URL path versioning. The API uses unversioned paths (`/api/`). If versioning is ever needed, header-based versioning (`Accept-Version`) SHOULD be used.
- Q: What task runner and containerization strategy should the project use? → A: Taskfile (go-task.dev) instead of Makefile. Docker with separate dev and prod builds plus a docker-compose.yml to run the full stack (API + PostgreSQL) in both dev and prod modes.
- Q: What database migration tool should the project use? → A: golang-migrate (github.com/golang-migrate/migrate).
- Q: What linting and coding standard tool should the project use? → A: golangci-lint with a sensible default configuration.
- Q: Should the application adhere to the 12-factor app methodology? → A: Yes, all 12 factors must be followed.
- Q: What libraries should be used for HTTP routing, JWT, Swagger, validation, and rate limiting? → A: chi (go-chi/chi) for HTTP routing, golang-jwt/jwt for JWT, swaggo/swag for Swagger docs generation, go-playground/validator for validation, golang.org/x/time/rate for rate limiting.
- Q: Should the dev docker-compose use live reload? → A: Yes, use Air (cosmtrek/air) for live reload in the dev Docker Compose setup.
- Q: What libraries should be used for configuration, testing, and utilities? → A: spf13/viper for configuration management, stretchr/testify for test assertions and mocking, samber/lo for generic utility functions.
- Q: What database driver and caching layer should the project use? → A: jmoiron/sqlx (on top of pgx/v5 stdlib driver) for PostgreSQL. Redis for caching infrastructure — provide the connection and interface seam only, no premature caching. Actual cache usage deferred to when profiling identifies a need.
- Q: What dependency injection framework should the project use? → A: uber/fx. Each vertical slice and shared infrastructure component registers as an fx.Module. All inter-layer dependencies are injected via interfaces, never concrete types.
- Q: Should the project use an in-house CQRS pattern to enforce vertical slice boundaries? → A: Yes. An in-house CQRS package (improved from github.com/ssengalanto/axiaos/apps/etl/pkg/cqrs) provides CommandBus, QueryBus, and EventBus with Go generics. Improvements over the original: (1) CommandHandler returns results via two type parameters [C, R], (2) no empty marker interfaces — use `any` as generic constraint, (3) validation via generic Validatable constraint instead of runtime type assertion, (4) structured domain error types that map to RFC 9457 problem details. Each HTTP handler becomes a thin adapter that dispatches commands/queries — business logic lives in CQRS handlers.
- Q: What email sending library should the project use? → A: resend-go/v2 (github.com/resend/resend-go/v2). Resend is the transactional email provider. The email service MUST be wrapped behind an interface for testability and future provider swaps.
- Q: What email use cases should the boilerplate implement? → A: Three flows: (1) verification email on registration with a 24-hour token, (2) password reset email with a 1-hour token, (3) welcome email after successful email verification. Welcome email is best-effort (failure does not block verification).
- Q: Should the system allow login before email verification? → A: No. Users MUST verify their email before they can log in. Login attempts with correct credentials but unverified email return 403 Forbidden with a clear message.
- Q: What URL structure should email-related endpoints use? → A: Noun-based, consistent with existing API conventions. Verify email → `POST /api/users/verify`, resend verification → `POST /api/users/resend-verification`, forgot password → `POST /api/users/forgot-password`, reset password → `POST /api/users/reset-password`.
- Q: How should verification and password reset tokens be stored? → A: Same pattern as refresh tokens: 256-bit random values, stored as SHA-256 hashes in PostgreSQL. Verification tokens expire after 24 hours, reset tokens after 1 hour. Both are single-use.
- Q: What rate limits should apply to email-related endpoints? → A: Resend-verification and forgot-password: 5 requests per minute per IP address. Verify and reset-password: 10 requests per minute per IP address.
- Q: Should password reset revoke existing sessions? → A: Yes. A successful password reset MUST revoke all refresh token families for that user, forcing re-authentication on all devices.
- Q: What configuration does the email service require? → A: Environment variables: `RESEND_API_KEY` (required in production), `EMAIL_FROM_ADDRESS` (sender address), `EMAIL_FROM_NAME` (sender display name), `APP_BASE_URL` (base URL for links in emails, e.g., verification and reset links).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow users to register with an email address and password.
- **FR-002**: System MUST validate email format and password strength at the API boundary before processing. Password MUST be at least 8 characters and contain at least one uppercase letter, one lowercase letter, one digit, and one special character.
- **FR-003**: System MUST hash passwords using bcrypt (cost >= 12) before storage; passwords MUST never appear in logs, API responses, or error messages.
- **FR-004**: System MUST issue a short-lived access token (~15 min, HS256-signed JWT) and a long-lived opaque refresh token (~7 days) upon successful login. Tokens are NOT issued at registration — the user must verify their email first.
- **FR-005**: Access tokens MUST carry sub, iat, exp, jti, iss, and aud claims. The iss and aud values MUST be configurable via environment variables, defaulting to `{{PROJECT_NAME}}` and `{{PROJECT_NAME}}-api` respectively. The auth middleware MUST validate both claims on every request.
- **FR-006**: Refresh tokens MUST be opaque 256-bit random values stored in Postgres as SHA-256 hashes, never as JWTs.
- **FR-007**: System MUST rotate refresh tokens on every use: issue a new refresh token and invalidate the old one.
- **FR-008**: System MUST detect refresh token reuse (a previously-rotated token presented again) and revoke the entire token family for that user, returning 401.
- **FR-009**: System MUST return identical error responses and take comparable time for "user not found" and "wrong password" scenarios to prevent user enumeration.
- **FR-010**: System MUST rate-limit POST /api/sessions (login) and POST /api/users (registration) endpoints with an in-memory limiter at 10 requests per minute per IP address. Email-related endpoints have their own rate limits defined in FR-084 and FR-089.
- **FR-011**: System MUST use timing-safe comparison (crypto/subtle) for token hash verification.
- **FR-012**: Auth middleware MUST inject a typed user ID into request context using an unexported key type (no string keys).
- **FR-013**: All API errors MUST conform to RFC 9457 problem details format (application/problem+json).
- **FR-014**: Validation failures MUST include field-level detail.
- **FR-015**: Internal errors MUST NOT leak driver messages, SQL, or stack traces; they MUST log with a correlation ID and return a generic message plus that ID.
- **FR-016**: System MUST provide a liveness endpoint (/healthz) with no dependency checks and a readiness endpoint (/readyz) that pings the database.
- **FR-017**: System MUST support graceful shutdown, draining in-flight requests within a configurable grace period.
- **FR-021**: System MUST run a background goroutine that periodically purges expired and revoked refresh tokens, expired/used verification tokens, and expired/used password reset tokens from the database on a configurable interval (default: hourly). The goroutine MUST respect graceful shutdown.
- **FR-018**: The application MUST refuse to boot if the JWT secret is missing or is a known default value in production mode.
- **FR-019**: System MUST never log tokens (access, refresh, verification, password reset), passwords, email content, or full authorization headers; a redaction mechanism MUST be in place and tested.
- **FR-022**: System MUST use Go's stdlib `log/slog` with a JSON handler for structured logging. All log entries MUST include a correlation ID. Log output MUST be JSON-formatted in production.
- **FR-020**: System MUST provide an API documentation interface (Swagger UI) accessible in development environments.
- **FR-023**: All JSON response representations MUST include a `self` property (the canonical URL of the resource) and a `kind` property (a string identifying the resource type, e.g., "User", "Session", "TokenPair"). Nested JSON objects MUST also include `self` and `kind` where they represent distinct resources.
- **FR-024**: Successful resource creation (e.g., user registration) MUST return HTTP 201 Created with a `Location` header pointing to the created resource's canonical URL.
- **FR-025**: All JSON property names MUST use camelCase convention (e.g., `createdAt`, `refreshToken`, `tokenFamily`). Avoid hyphens, snake_case, and characters incompatible with JavaScript identifiers.
- **FR-026**: All date/time values in JSON responses MUST use ISO 8601 format with UTC timezone (e.g., `2026-09-04T12:30:00Z`).
- **FR-027**: API URLs MUST use noun-based resource naming with the `/api/` prefix per the data-oriented design paradigm. Auth endpoints: registration → `POST /api/users`, login → `POST /api/sessions`, token refresh → `POST /api/tokens`, logout → `DELETE /api/sessions/current`, user profile → `GET /api/users/me`. Email endpoints: verify email → `POST /api/users/verify`, resend verification → `POST /api/users/resend-verification`, forgot password → `POST /api/users/forgot-password`, reset password → `POST /api/users/reset-password`. HTTP methods (GET, POST, DELETE) express the action; the URL identifies the resource.
- **FR-028**: The API uses unversioned paths (`/api/`). All `self` links in responses use the `/api/` prefix. If versioning is ever needed, header-based versioning (`Accept-Version`) SHOULD be used. Clients MUST be designed to tolerate new properties appearing in responses (forward-compatible).
- **FR-029**: The project MUST use Taskfile (go-task.dev) instead of Makefile for task automation. Common tasks (build, test, lint, run, migrate, docker) MUST be defined as Task targets.
- **FR-030**: The project MUST provide two Dockerfile targets: a **dev** build (with hot-reload, debug tooling, and mounted source) and a **prod** build (multi-stage, minimal distroless/scratch image, statically linked binary, non-root user). The prod image MUST not contain build tools, source code, or unnecessary dependencies.
- **FR-031**: The project MUST provide a `docker-compose.yml` that runs the full application stack (API server + PostgreSQL + Redis) with a single `docker compose up` command. It MUST support both dev mode (with hot-reload and source mounting) and prod mode (using the production image). Environment-specific configuration MUST use compose profiles or override files.
- **FR-032**: Database schema changes MUST be managed using golang-migrate (github.com/golang-migrate/migrate). Migrations MUST be stored as versioned SQL files (up/down pairs). The Taskfile MUST include targets for creating, applying, and rolling back migrations.
- **FR-033**: The project MUST include a golangci-lint configuration (`.golangci.yml`) with a sensible set of enabled linters. The Taskfile MUST include a `lint` target. Linting MUST pass with zero issues before code is considered ready for review.
- **FR-046**: HTTP routing MUST use go-chi/chi. Middleware (auth, logging, rate limiting, CORS, request ID) MUST be composed via chi's middleware chain.
- **FR-047**: JWT token creation and validation MUST use golang-jwt/jwt (v5).
- **FR-048**: API documentation MUST be generated from Go source annotations using swaggo/swag. The Swagger UI MUST be served at a configurable path (e.g., `/swagger/`) in development environments. The Taskfile MUST include a `swagger` target to regenerate docs.
- **FR-049**: Request payload validation MUST use go-playground/validator with struct tags. Validation errors MUST be translated into RFC 9457 field-level detail (FR-014).
- **FR-050**: In-memory rate limiting MUST use golang.org/x/time/rate (token bucket algorithm). The limiter MUST be keyed per IP address with configurable rate and burst values.
- **FR-051**: The dev Docker Compose service MUST use Air (cosmtrek/air) for live reload. Source code MUST be mounted into the container, and Air MUST watch for `.go` file changes and automatically rebuild/restart the application. An `.air.toml` configuration file MUST be included in the project.
- **FR-052**: Configuration loading MUST use spf13/viper. Viper MUST read from environment variables (12-factor compliant) with support for `.env` files in development. Configuration MUST be loaded into typed structs with validation at startup. Viper's env var binding MUST be the source of truth; `.env` files are a development convenience only.
- **FR-053**: All tests MUST use stretchr/testify for assertions (`assert`, `require`) and test suite organization. Table-driven tests MUST use `assert` for non-fatal checks and `require` for preconditions that should halt the test. Testify's `mock` package MAY be used for interface mocking where appropriate.
- **FR-054**: The project MUST use samber/lo for generic utility operations (map, filter, reduce, contains, etc.) instead of hand-rolled loops where lo provides a clearer, more concise alternative.
- **FR-055**: Database access MUST use jmoiron/sqlx with the pgx/v5 stdlib-compatible driver (`github.com/jackc/pgx/v5/stdlib`). Queries MUST use named parameters or struct scanning via sqlx's `StructScan`, `NamedExec`, and `Get`/`Select` methods. Raw `database/sql` SHOULD be avoided where sqlx provides a cleaner alternative.
- **FR-056**: The project MUST include Redis (go-redis/redis) as a caching infrastructure dependency. The docker-compose MUST run a Redis instance alongside PostgreSQL. The application MUST establish a Redis connection at startup and expose a cache interface seam. However, no caching logic MUST be implemented prematurely — actual cache usage MUST be deferred until profiling or load testing identifies a concrete performance bottleneck. The interface seam ensures caching can be added to any layer without structural changes.
- **FR-057**: Dependency injection MUST use uber/fx. Each vertical slice (feature) and each shared infrastructure component (database, Redis, logger, config) MUST register as an fx.Module. All inter-layer dependencies MUST be injected via interfaces, never concrete types. The application's main function MUST construct the dependency graph via fx.New() and fx.Options().
- **FR-058**: The system MUST respond with 405 Method Not Allowed when an HTTP method is used on a valid URL that does not support that method. The 405 response MUST include an `Allow` header listing the methods the URL does support (e.g., `Allow: GET, POST`).

#### Email Infrastructure

- **FR-075**: Transactional email sending MUST use resend-go/v2 (`github.com/resend/resend-go/v2`). The Resend client MUST be wrapped behind an email service interface (e.g., `EmailSender`) for testability and provider substitution. The email service MUST register as an fx.Module.
- **FR-076**: The application MUST refuse to boot in production mode if `RESEND_API_KEY` is missing or empty (same pattern as FR-018 for JWT secret). In development mode, a no-op or logging email sender MAY be used to allow running without a Resend account.
- **FR-077**: Email service configuration MUST be read from environment variables: `RESEND_API_KEY` (API key), `EMAIL_FROM_ADDRESS` (sender email), `EMAIL_FROM_NAME` (sender display name), `APP_BASE_URL` (base URL for constructing links in emails). All MUST be validated at startup.
- **FR-078**: System MUST never log email content, verification tokens, or password reset tokens. The email service MUST log send attempts with correlation ID, recipient (redacted or hashed), and email type (verification, reset, welcome) — never the token or email body.

#### Email Verification

- **FR-079**: Upon successful registration, the system MUST create the user with `emailVerified = false` and dispatch a verification email containing a unique, time-limited token.
- **FR-080**: Verification tokens MUST be 256-bit random values stored in PostgreSQL as SHA-256 hashes (same pattern as refresh tokens, FR-006). Each token MUST have a 24-hour expiry and be single-use.
- **FR-081**: The system MUST expose `POST /api/users/verify` accepting a token in the request body. A valid, unexpired, unused token transitions the user's `emailVerified` to `true`, marks the token as used, and triggers a welcome email.
- **FR-082**: The system MUST expose `POST /api/users/resend-verification` accepting an email in the request body. If the email belongs to an unverified user, the system invalidates all existing verification tokens for that user, generates a new token, and sends a new verification email. The response MUST be identical regardless of whether the email exists or is already verified (no user enumeration).
- **FR-083**: Login (POST /api/sessions) MUST reject users whose email is not verified with 403 Forbidden. The error response MUST indicate that email verification is required, using RFC 9457 problem details format. The response time MUST be comparable to successful credential validation to prevent timing-based enumeration.
- **FR-084**: The resend-verification endpoint MUST be rate-limited at 5 requests per minute per IP address. The verify endpoint MUST be rate-limited at 10 requests per minute per IP address.
- **FR-085**: Verification token errors (expired, already used, invalid) MUST return identical 401 Unauthorized responses (no distinction between error types to prevent token enumeration).

#### Password Reset

- **FR-086**: The system MUST expose `POST /api/users/forgot-password` accepting an email in the request body. If the email belongs to a verified user, the system generates a password reset token and sends a reset email. The response MUST always be 200 OK regardless of whether the email exists, is unverified, or is unknown (no user enumeration). Response time MUST be comparable in all cases.
- **FR-087**: Password reset tokens MUST be 256-bit random values stored in PostgreSQL as SHA-256 hashes. Each token MUST have a 1-hour expiry and be single-use.
- **FR-088**: The system MUST expose `POST /api/users/reset-password` accepting a token and a new password in the request body. A valid, unexpired, unused token updates the user's password hash (bcrypt, cost >= 12), marks the token as used, and revokes ALL refresh token families for that user (forcing re-authentication on all devices).
- **FR-089**: The forgot-password endpoint MUST be rate-limited at 5 requests per minute per IP address. The reset-password endpoint MUST be rate-limited at 10 requests per minute per IP address.
- **FR-090**: Password reset MUST NOT be available for users with unverified emails. The forgot-password endpoint MUST silently skip sending the email in this case (returning the same 200 OK response).
- **FR-091**: When a new password reset token is generated for a user, all previous unused reset tokens for that user MUST be invalidated.

#### Welcome Email

- **FR-092**: Upon successful email verification, the system MUST send a welcome email to the user. The welcome email is best-effort: if sending fails, the verification MUST still succeed. The failure MUST be logged with a correlation ID.
- **FR-093**: Welcome email sending SHOULD be dispatched asynchronously via the EventBus (e.g., `EmailVerifiedEvent` → welcome email handler) to avoid blocking the verification response.

#### CQRS Infrastructure

- **FR-068**: The project MUST include an in-house CQRS package as shared infrastructure (e.g., `pkg/cqrs/`). The package MUST provide CommandBus, QueryBus, and EventBus using Go generics for compile-time type safety. The package MUST have zero external dependencies (stdlib only: `context`, `sync`, `reflect`, `fmt`, `time`).
- **FR-069**: CommandBus MUST support commands that return results. `CommandHandler[C any, R any]` MUST define `Handle(ctx context.Context, cmd C) (R, error)`. `CommandBus[C, R]` MUST accept a single handler and variadic middleware, dispatching via `Dispatch(ctx, cmd) (R, error)`. Commands that return no meaningful result MUST use a defined unit type (e.g., `type Unit = struct{}`).
- **FR-070**: QueryBus MUST support queries that return typed results. `QueryHandler[Q any, R any]` MUST define `Handle(ctx context.Context, query Q) (R, error)`. `QueryBus[Q, R]` MUST accept a single handler and variadic middleware, dispatching via `Ask(ctx, query) (R, error)`. Semantic distinction from CommandBus: queries MUST NOT mutate state.
- **FR-071**: EventBus MUST support multiple handlers per event type (pub-sub). `EventHandler[E any]` MUST define `Handle(ctx context.Context, event E) error`. EventBus MUST support both synchronous (fail-fast on first error) and asynchronous (concurrent goroutines, collect errors) dispatch modes. Handlers MUST be registered dynamically via `Register()`.
- **FR-072**: The CQRS package MUST provide a composable middleware chain for commands, queries, and events. Middleware MUST follow the `func(ctx, input, next) (result, error)` pattern with reverse-order decoration (outermost executes first). Built-in middleware MUST include: (1) **Logging** — uses slog, logs type name via `reflect.TypeFor[C]().Name()`, duration, and result; (2) **Validation** — uses a generic `Validatable` constraint (`interface{ Validate() error }`) checked at compile time, not a runtime type assertion; (3) **Recovery** — converts panics to errors.
- **FR-073**: The CQRS package MUST define structured domain error types (e.g., `NotFoundError`, `ConflictError`, `ValidationError`, `UnauthorizedError`) that carry enough context (resource type, identifier, field details) for the HTTP layer to translate them directly into RFC 9457 problem details (FR-013) without inspecting error strings. Each domain error MUST implement the `error` interface and support `errors.Is`/`errors.As` unwrapping.
- **FR-074**: Each vertical slice's HTTP handler MUST be a thin adapter: parse the HTTP request, construct the appropriate command or query struct, dispatch it via the bus, and translate the result or domain error into an HTTP response. Business logic MUST NOT reside in HTTP handlers — it MUST live exclusively in command/query handlers. The command/query handler interacts with repositories and other dependencies via injected interfaces.

#### SOLID Principles Compliance

- **FR-059 (Single Responsibility)**: Each component (HTTP handler, command/query handler, repository) MUST have a single, well-defined responsibility. HTTP handlers MUST only perform HTTP concerns (parse request, construct command/query, dispatch via bus, write response). Command/query handlers MUST only contain business logic. Repositories MUST only perform data access. No component MUST mix these concerns.
- **FR-060 (Open/Closed)**: The architecture MUST be extensible without modifying existing code. New features MUST be added as new vertical slices without changing existing slice code. Cross-cutting behavior MUST be composed via middleware at two levels: HTTP middleware (chi) for transport concerns (auth, rate limiting, CORS, request ID) and CQRS middleware for application concerns (logging, validation, recovery). Neither requires editing handlers.
- **FR-061 (Liskov Substitution)**: All interface implementations MUST be substitutable. Test doubles (mocks, stubs) MUST satisfy the same interface contracts as production implementations. No implementation MUST impose preconditions stricter than its interface specifies.
- **FR-062 (Interface Segregation)**: Go interfaces MUST be small and focused (typically 1-3 methods). Each layer boundary MUST define its own interface for the dependency it consumes (e.g., a service defines the repository interface it needs, not the other way around). No component MUST be forced to depend on methods it does not use.
- **FR-063 (Dependency Inversion)**: High-level modules (services) MUST NOT depend on low-level modules (repositories, drivers). Both MUST depend on abstractions (interfaces). Interfaces MUST be defined in the package that uses them (consumer-side), not the package that implements them. uber/fx MUST wire concrete implementations to these interfaces at the composition root.

#### Vertical Slice Architecture Compliance

- **FR-064**: Each feature MUST be organized as a vertical slice in its own directory under `features/` (e.g., `features/auth/`, `features/user/`, `features/health/`). Each slice MUST contain its own HTTP handler, command/query definitions, command/query handlers, and repository (as needed), co-located in the slice directory.
- **FR-065**: Each vertical slice MUST register itself as an fx.Module that provides its HTTP handler, command/query bus (with wired handler and middleware), and repository implementations to the dependency graph. Slices MUST NOT import from other slice directories.
- **FR-066**: Cross-cutting concerns (middleware, error handling, database connection, Redis connection, configuration, logging) MUST reside in shared infrastructure packages outside the `features/` directory (e.g., `pkg/` or `internal/`). These shared packages MUST NOT contain business logic specific to any feature.
- **FR-067**: Adding a new feature MUST require only: (1) creating a new directory under `features/`, (2) implementing the slice's handler/service/repository, (3) registering the slice's fx.Module in the application's module list. No existing feature code MUST be modified.

#### 12-Factor App Compliance

- **FR-034 (I. Codebase)**: One codebase tracked in Git, deployable to multiple environments (dev, staging, prod) without code changes.
- **FR-035 (II. Dependencies)**: All dependencies MUST be explicitly declared via `go.mod`. The application MUST NOT rely on system-wide packages or implicit OS-level dependencies. The prod Docker image MUST be self-contained.
- **FR-036 (III. Config)**: ALL environment-specific configuration (database URL, JWT secret, token durations, rate limit values, server port, log level, Resend API key, email sender address, email sender name, app base URL) MUST be read from environment variables. No config files for secrets. The application MUST fail fast with a clear error if required environment variables are missing.
- **FR-037 (IV. Backing Services)**: PostgreSQL, Redis, and Resend MUST be treated as attached resources/services, connectable via `DATABASE_URL`, `REDIS_URL`, and `RESEND_API_KEY` environment variables respectively. Swapping any backing service instance (e.g., local to managed cloud, or Resend to another email provider) MUST require only configuration changes, no business logic changes.
- **FR-038 (V. Build, Release, Run)**: Build (compile binary), release (binary + config), and run (execute process) stages MUST be strictly separated. The prod Docker multi-stage build enforces this: build stage compiles, final stage runs.
- **FR-039 (VI. Processes)**: The application MUST execute as a stateless process. All persistent state (users, tokens) MUST live in PostgreSQL. No sticky sessions or in-process state that cannot be lost on restart (the in-memory rate limiter is acceptable as a non-critical, self-recovering cache).
- **FR-040 (VII. Port Binding)**: The application MUST be self-contained and export HTTP via port binding. The listen port MUST be configurable via a `PORT` or `SERVER_ADDRESS` environment variable.
- **FR-041 (VIII. Concurrency)**: The application MUST be designed to scale horizontally by running multiple identical processes. No inter-process communication or shared memory is assumed (the in-memory rate limiter is documented as a single-instance limitation with an interface seam for Redis).
- **FR-042 (IX. Disposability)**: The application MUST start fast (sub-second to ready) and shut down gracefully (FR-017). On SIGTERM, it MUST stop accepting new requests, drain in-flight requests, close database connections, and stop the token cleanup goroutine cleanly.
- **FR-043 (X. Dev/Prod Parity)**: The docker-compose setup MUST use the same backing service type (PostgreSQL) in development and production. The time gap between dev and deploy SHOULD be minimal. Developers MUST NOT use SQLite or in-memory databases as PostgreSQL substitutes.
- **FR-044 (XI. Logs)**: The application MUST treat logs as event streams. All log output MUST go to stdout (never to log files). Log aggregation and routing is the responsibility of the execution environment, not the application.
- **FR-045 (XII. Admin Processes)**: One-off administrative tasks (database migrations, seed data) MUST be run as separate processes using the same codebase and config. Migrations via golang-migrate (FR-032) MUST be executable as standalone commands, not embedded in application startup.

### Key Entities

- **User**: Represents a registered account. Key attributes: unique identifier, email (case-insensitive, unique), password hash, email verification status (`emailVerified`), creation and update timestamps. A user owns zero or more refresh token families, verification tokens, and password reset tokens.
- **Refresh Token**: Represents a single token in a rotation chain. Key attributes: unique identifier, owning user, token hash, family identifier (groups tokens in a rotation chain), expiration, revocation status, creation timestamp. Tokens within the same family form a chain; revoking a family invalidates all tokens in it.
- **Verification Token**: Represents a single-use token for email verification. Key attributes: unique identifier, owning user, token hash (SHA-256 of the 256-bit random value), expiration (24 hours from creation), used-at timestamp (null if unused), creation timestamp. Only the latest token for a given user is valid; generating a new one invalidates all previous tokens.
- **Password Reset Token**: Represents a single-use token for password reset. Key attributes: unique identifier, owning user, token hash (SHA-256 of the 256-bit random value), expiration (1 hour from creation), used-at timestamp (null if unused), creation timestamp. Only the latest token for a given user is valid; generating a new one invalidates all previous tokens.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new developer can go from cloning the repository to a successful POST /api/sessions call in under 5 minutes, following only the README.
- **SC-002**: The complete register-verify-login-access-refresh-logout flow completes successfully in an automated end-to-end test.
- **SC-003**: Presenting a previously-rotated refresh token results in immediate revocation of the entire token family, verified by automated test.
- **SC-004**: Login with a non-existent email and login with a wrong password produce indistinguishable responses (same status code, same response body structure, comparable response time within 100ms).
- **SC-005**: No password, token value, or full authorization header appears in any log output, verified by automated test.
- **SC-006**: The application refuses to start when critical configuration (JWT secret) is missing or set to a known insecure default, verified by test.
- **SC-007**: All API error responses conform to RFC 9457 problem details format, verified by schema validation in tests.
- **SC-008**: Adding a new feature slice requires only: creating a new directory under features/, implementing handler/service/repository within it, and registering its fx.Module — no existing feature code is modified.
- **SC-009**: All JSON responses include `self` and `kind` properties, use camelCase property names, and use ISO 8601 date/time format, verified by schema validation in tests.
- **SC-010**: All inter-layer dependencies (handler→service, service→repository) are injected via interfaces using uber/fx; no concrete type is imported across layer boundaries, verified by code review or architectural linting.
- **SC-011**: Sending an unsupported HTTP method to any defined endpoint returns 405 Method Not Allowed with an `Allow` header listing supported methods, verified by automated test.
- **SC-012**: Every use case (register, login, refresh, logout, get profile, verify email, resend verification, forgot password, reset password) is implemented as a command or query handler dispatched via the CQRS bus. HTTP handlers contain no business logic — they only parse requests, dispatch, and translate responses. Verified by code review.
- **SC-013**: A user who registers but does not verify their email cannot log in (403 Forbidden), verified by automated test.
- **SC-014**: The complete forgot-password → reset-password flow successfully changes the user's password and revokes all existing refresh token families, verified by automated end-to-end test.
- **SC-015**: Forgot-password requests for non-existent and existing emails produce indistinguishable responses (same status code, same response body structure, comparable response time within 100ms), verified by automated test.
- **SC-016**: No verification token, password reset token, or email content appears in any log output, verified by automated test.
- **SC-017**: The application refuses to start in production mode when `RESEND_API_KEY` is missing, verified by test.
- **SC-018**: The welcome email is sent after successful email verification and does not block the verification response if sending fails, verified by automated test.

## Assumptions

- **Module path and project name**: The user will provide `{{MODULE_PATH}}` and `{{PROJECT_NAME}}` before implementation begins; these are template placeholders in the prompt.
- **Go version**: The latest stable Go release will be used; the implementation does not target a specific older version.
- **Single-instance deployment initially**: The in-memory rate limiter is acceptable for single-instance deployments; multi-instance deployments will require a Redis-backed limiter (interface seam is provided but Redis implementation is out of scope).
- **Resend as email provider**: Resend (via resend-go/v2) is the transactional email provider. The email service is behind an interface, so swapping providers requires only a new implementation — no business logic changes.
- **Email verification required for login**: Users must verify their email before they can log in. Registration creates the account but does not issue tokens. This is a hard gate, not a soft reminder.
- **No email templates engine**: Email content is constructed in Go code (plain text and/or simple HTML). A full template engine (e.g., Go html/template with external files) is an explicit future extension point, not implemented here.
- **No roles or permissions**: The auth system issues tokens with user identity only; role-based access control is an explicit future extension point, not implemented here.
- **PostgreSQL is the only supported database**: The data access layer uses jmoiron/sqlx with the pgx/v5 stdlib driver and does not abstract over multiple database engines.
- **Development environment**: Developers have Docker, Go, and Task (go-task.dev) installed locally. The full stack (API + Postgres) can run entirely via `docker compose up` or the API can run on the host with only Postgres in Docker.
- **HS256 JWT signing by default**: Access tokens use HS256 (symmetric) signing. RS256 (asymmetric) is documented as a switch but not the default configuration.
- **bcrypt for password hashing**: bcrypt (cost >= 12) is the chosen algorithm over argon2id for its maturity, simplicity, and wide ecosystem support; argon2id's memory-hardness advantage is noted but not required for this boilerplate's threat model.

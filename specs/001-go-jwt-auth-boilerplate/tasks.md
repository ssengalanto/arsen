# Tasks: Go Backend Boilerplate with JWT Authentication

**Input**: Design documents from `/specs/001-go-jwt-auth-boilerplate/`

**Prerequisites**: plan.md, spec.md, data-model.md, contracts/, research.md, quickstart.md

**Tests**: Included — spec explicitly requests TDD approach.

**Organization**: Tasks grouped by user story. US1+US8+US10 are combined (Register → Verify → Welcome) since they form a single flow.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story (US1, US2, etc.)
- Paths are relative to repository root

---

## Phase 1: Setup

**Purpose**: Project initialization, tooling, and configuration scaffolding

- [x] T001 Initialize Go module with `go mod init arsen` and create `cmd/api/main.go` entry point stub
- [x] T002 [P] Create `Taskfile.yml` with targets: build, test, lint, run, migrate-up, migrate-down, migrate-create, swagger, dev, docker-up, docker-down
- [x] T003 [P] Create `Dockerfile` with multi-target build: dev stage (Air hot-reload, debug tools, mounted source) and prod stage (multi-stage, distroless/scratch, statically linked binary, non-root user)
- [x] T004 [P] Create `docker-compose.yml` running API + PostgreSQL + Redis with dev/prod profiles
- [x] T005 [P] Create `.air.toml` configuration for live reload watching `.go` files
- [x] T006 [P] Create `.golangci.yml` with sensible linter set (govet, errcheck, staticcheck, unused, gosimple, ineffassign, typecheck, gocritic)
- [x] T007 [P] Create `.env.example` with all required environment variables: DATABASE_URL, REDIS_URL, SERVER_ADDRESS, JWT_SECRET, JWT_ISSUER, JWT_AUDIENCE, ACCESS_TOKEN_DURATION, REFRESH_TOKEN_DURATION, RESEND_API_KEY, EMAIL_FROM_ADDRESS, EMAIL_FROM_NAME, APP_BASE_URL, LOG_LEVEL, ENV
- [x] T008 Create directory structure per plan.md: `cmd/api/`, `features/auth/`, `features/user/`, `features/health/`, `pkg/cqrs/`, `pkg/cqrs/middleware/`, `pkg/config/`, `pkg/database/`, `pkg/redis/`, `pkg/email/`, `pkg/token/`, `pkg/server/`, `pkg/middleware/`, `pkg/response/`, `pkg/cleanup/`, `migrations/`

**Checkpoint**: Project skeleton compiles (`go build ./...` succeeds with stub main), Docker Compose starts PostgreSQL and Redis, `task lint` runs.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### CQRS Package (pkg/cqrs/) — zero external dependencies

- [x] T009 Implement `CommandBus[C, R]` and `CommandHandler[C, R]` with `Dispatch(ctx, cmd) (R, error)` in `pkg/cqrs/command.go`. Include `Unit` type alias (`type Unit = struct{}`). Include middleware chain support with reverse-order decoration.
- [x] T010 [P] Implement `QueryBus[Q, R]` and `QueryHandler[Q, R]` with `Ask(ctx, query) (R, error)` in `pkg/cqrs/query.go`. Same middleware chain pattern as CommandBus.
- [x] T011 [P] Implement `EventBus[E]` and `EventHandler[E]` with sync and async dispatch modes in `pkg/cqrs/event.go`. Support multiple handlers per event type via `Register()`.
- [x] T012 [P] Implement domain error types in `pkg/cqrs/errors.go`: `NotFoundError`, `ConflictError`, `ValidationError`, `UnauthorizedError`, `ForbiddenError`. Each carries resource type, identifier, field details. All implement `error` and support `errors.Is`/`errors.As`.
- [x] T013 [P] Implement CQRS middleware in `pkg/cqrs/middleware/logging.go` — uses slog, logs type name via `reflect.TypeFor[C]().Name()`, duration, and result
- [x] T014 [P] Implement CQRS middleware in `pkg/cqrs/middleware/validation.go` — generic `Validatable` constraint (`interface{ Validate() error }`) checked at compile time
- [x] T015 [P] Implement CQRS middleware in `pkg/cqrs/middleware/recovery.go` — converts panics to errors

### CQRS Tests

- [x] T016 [P] Write tests for CommandBus dispatch, middleware chain, and Unit return type in `pkg/cqrs/command_test.go`
- [x] T017 [P] Write tests for QueryBus dispatch and middleware chain in `pkg/cqrs/query_test.go`
- [x] T018 [P] Write tests for EventBus sync/async dispatch, multiple handlers, error collection in `pkg/cqrs/event_test.go`
- [x] T019 [P] Write tests for domain error types (Is/As unwrapping, field details) in `pkg/cqrs/errors_test.go`
- [x] T020 [P] Write tests for logging, validation, and recovery middleware in `pkg/cqrs/middleware/middleware_test.go`

### Infrastructure Modules

- [x] T021 Implement Viper-based configuration loading in `pkg/config/config.go` — typed structs for Server, Database, Redis, JWT, Email, RateLimit, Cleanup config. Validate required fields at startup. Refuse to boot if JWT_SECRET or RESEND_API_KEY missing in production mode. Register as fx.Module in `pkg/config/module.go`
- [x] T022 [P] Implement PostgreSQL connection via sqlx + pgx/v5 stdlib driver in `pkg/database/database.go`. Include ping health check method. Register as fx.Module with OnStart (connect) and OnStop (close) lifecycle hooks in `pkg/database/module.go`
- [x] T023 [P] Implement Redis connection via go-redis/v9 in `pkg/redis/redis.go`. Define cache interface seam (Get, Set, Delete). Register as fx.Module with lifecycle hooks in `pkg/redis/module.go`
- [x] T024 [P] Implement email service interface (`EmailSender` with `SendEmail(ctx, to, subject, htmlBody) error`) and Resend implementation in `pkg/email/email.go`. Implement no-op logging sender for dev mode in `pkg/email/noop.go`. Register as fx.Module (select impl based on ENV) in `pkg/email/module.go`
- [x] T025 [P] Implement secure random token generation (256-bit) and SHA-256 hashing utilities in `pkg/token/token.go`. Include timing-safe comparison via crypto/subtle.

### HTTP Infrastructure

- [x] T026 Implement RFC 9457 problem details response builder in `pkg/response/problem.go` — struct for ProblemDetail with type, title, status, detail, instance, errors fields. Include helper functions for common error types (BadRequest, Unauthorized, Forbidden, NotFound, Conflict, TooManyRequests, InternalError with correlation ID).
- [x] T027 [P] Implement JSON response writer in `pkg/response/json.go` — writes JSON with `self` and `kind` properties, camelCase field names, ISO 8601 dates. Include `RespondCreated` helper with Location header.
- [x] T028 [P] Implement domain error → RFC 9457 translator in `pkg/response/problem.go` — maps `cqrs.NotFoundError` → 404, `cqrs.ConflictError` → 409, `cqrs.ValidationError` → 400 with field details, `cqrs.UnauthorizedError` → 401, `cqrs.ForbiddenError` → 403.
- [x] T029 Implement request ID / correlation ID middleware in `pkg/middleware/requestid.go` — generates UUID, injects into context and slog, sets X-Request-ID response header
- [x] T030 [P] Implement structured request logging middleware in `pkg/middleware/logging.go` — uses slog with JSON handler, logs method, path, status, duration, correlation ID. Redacts Authorization headers and token values.
- [x] T031 [P] Implement per-IP rate limiter middleware in `pkg/middleware/ratelimit.go` — uses golang.org/x/time/rate token bucket, keyed per IP, configurable rate and burst. Returns 429 with RFC 9457 body.
- [x] T032 [P] Implement CORS middleware in `pkg/middleware/cors.go` — configurable origins, methods, headers
- [x] T033 [P] Implement JWT auth middleware in `pkg/middleware/auth.go` — validates Bearer token from Authorization header, checks exp/iss/aud claims, injects typed user ID into context using unexported key type. Returns 401 for missing/expired/tampered tokens.
- [x] T034 Implement chi router setup and middleware composition in `pkg/server/server.go` — compose request ID, logging, CORS, recovery middleware. Include 405 Method Not Allowed handler with Allow header. Register as fx.Module with OnStart (ListenAndServe) and OnStop (graceful shutdown with configurable grace period) in `pkg/server/module.go`

### Database Migrations

- [x] T035 Create migration `migrations/000001_create_users.up.sql` — enable citext extension, create users table (id UUID PK, email citext UNIQUE, password_hash text, email_verified boolean DEFAULT false, created_at timestamptz, updated_at timestamptz)
- [x] T036 [P] Create migration `migrations/000001_create_users.down.sql` — drop users table
- [x] T037 [P] Create migration `migrations/000002_create_refresh_tokens.up.sql` — create refresh_tokens table with indexes per data-model.md
- [x] T038 [P] Create migration `migrations/000002_create_refresh_tokens.down.sql` — drop refresh_tokens table
- [x] T039 [P] Create migration `migrations/000003_create_verification_tokens.up.sql` — create verification_tokens table with indexes per data-model.md
- [x] T040 [P] Create migration `migrations/000003_create_verification_tokens.down.sql` — drop verification_tokens table
- [x] T041 [P] Create migration `migrations/000004_create_password_reset_tokens.up.sql` — create password_reset_tokens table with indexes per data-model.md
- [x] T042 [P] Create migration `migrations/000004_create_password_reset_tokens.down.sql` — drop password_reset_tokens table

### Infrastructure Tests

- [x] T043 [P] Write tests for config validation (missing JWT_SECRET fails, missing RESEND_API_KEY fails in prod, succeeds in dev) in `pkg/config/config_test.go`
- [x] T044 [P] Write tests for token generation (256-bit entropy, unique), hashing (SHA-256), and timing-safe comparison in `pkg/token/token_test.go`
- [x] T045 [P] Write tests for RFC 9457 problem details builder and domain error translator in `pkg/response/problem_test.go`
- [x] T046 [P] Write tests for JSON response writer (self/kind included, camelCase, ISO 8601 dates) in `pkg/response/json_test.go`
- [x] T047 [P] Write tests for rate limiter middleware (allows within limit, returns 429 when exceeded) in `pkg/middleware/ratelimit_test.go`
- [x] T048 [P] Write tests for JWT auth middleware (valid token passes, missing/expired/tampered rejected, user ID in context) in `pkg/middleware/auth_test.go`
- [x] T049 [P] Write tests for email service (Resend mock, no-op sender logs but doesn't send) in `pkg/email/email_test.go`

**Checkpoint**: All infrastructure compiles, all pkg tests pass. `go test ./pkg/...` green. Migrations apply successfully against Docker PostgreSQL. fx dependency graph resolves. Application boots and shuts down gracefully (no routes yet).

---

## Phase 3: US1 + US8 + US10 — Registration + Email Verification + Welcome Email (Priority: P1) 🎯 MVP

**Goal**: A user can register, receive a verification email, verify their email, and receive a welcome email. Unverified users cannot log in.

**Independent Test**: Register via POST /api/users → extract verification token from logs → verify via POST /api/users/verify → confirm emailVerified=true and welcome email dispatched. Also: POST /api/users/resend-verification works for unverified users.

### Tests

- [x] T050 [P] [US1] Write tests for RegisterCommand handler in `features/user/register_command_test.go` — valid registration creates user with emailVerified=false, duplicate email returns conflict error, password validation enforced, bcrypt hash generated, verification token created, email service called
- [x] T051 [P] [US8] Write tests for VerifyEmailCommand handler in `features/user/verify_email_command_test.go` — valid token verifies email, expired token rejected, used token rejected, invalid token rejected, welcome event published
- [x] T052 [P] [US8] Write tests for ResendVerificationCommand handler in `features/user/resend_verification_command_test.go` — invalidates old tokens, creates new token, sends email, returns same response for non-existent email (no enumeration)
- [x] T053 [P] [US1] Write HTTP handler tests for POST /api/users in `features/user/handler_test.go` — 201 response with Location header, validation errors return RFC 9457, rate limiting returns 429
- [x] T054 [P] [US8] Write HTTP handler tests for POST /api/users/verify and POST /api/users/resend-verification in `features/user/handler_test.go` — success responses, 401 for bad tokens, rate limiting

### Implementation

- [x] T055 [US1] Implement user repository interface and sqlx implementation in `features/user/repository.go` — Create(ctx, user), GetByEmail(ctx, email), GetByID(ctx, id), UpdateEmailVerified(ctx, id, verified), UpdatePasswordHash(ctx, id, hash). Include verification token methods: CreateVerificationToken(ctx, token), GetVerificationTokenByHash(ctx, hash), InvalidateUserVerificationTokens(ctx, userID)
- [x] T056 [US1] Implement RegisterCommand and RegisterCommandHandler in `features/user/register_command.go` — validate input (email format, password strength via custom validator), check email uniqueness, bcrypt hash password (cost >= 12), create user with emailVerified=false, generate verification token (256-bit), store token hash, call EmailSender with verification link, return created user. Anti-enumeration: same error for duplicate email.
- [x] T057 [US8] Implement VerifyEmailCommand and VerifyEmailCommandHandler in `features/user/verify_email_command.go` — hash submitted token, look up in DB, check not expired (24h) and not used, mark token used, set user emailVerified=true, publish EmailVerifiedEvent via EventBus. All error cases return identical UnauthorizedError.
- [x] T058 [US8] Implement ResendVerificationCommand and ResendVerificationCommandHandler in `features/user/resend_verification_command.go` — look up user by email, if exists and unverified: invalidate existing tokens, generate new token, send verification email. Always return success (no enumeration).
- [x] T059 [US10] Implement EmailVerifiedEvent and welcome email handler in `features/user/welcome_email_handler.go` — EventHandler[EmailVerifiedEvent] that sends welcome email via EmailSender. Best-effort: log failure but don't propagate error. Register on EventBus (async dispatch).
- [x] T060 [US1] Implement HTTP handlers in `features/user/handler.go` — POST /api/users (register), POST /api/users/verify, POST /api/users/resend-verification. Each handler: parse request, construct command, dispatch via bus, translate result/error to HTTP response. Apply rate limiting per endpoint (10/min register, 10/min verify, 5/min resend-verification).
- [x] T061 [US1] Implement fx.Module registration in `features/user/module.go` — provide repository, command buses (with logging+validation+recovery middleware), event bus with welcome email handler, HTTP handler. Wire into router.
- [x] T062 [US1] Wire user feature module into `cmd/api/main.go` — add to fx.Options()

**Checkpoint**: `go test ./features/user/...` green. Can register via curl, see verification token in dev logs, verify email, confirm welcome email logged. Resend-verification works. Duplicate email returns generic error.

---

## Phase 4: US2 — User Login (Priority: P1)

**Goal**: A verified user can log in with email/password and receive access + refresh tokens. Unverified users get 403. Wrong credentials get generic 401 with constant-time response.

**Independent Test**: Create verified user → POST /api/sessions with correct credentials → receive accessToken + refreshToken. Test unverified user → 403. Test wrong password → 401 indistinguishable from non-existent user.

### Tests

- [x] T063 [P] [US2] Write tests for LoginCommand handler in `features/auth/login_command_test.go` — valid credentials return tokens, wrong password returns unauthorized, non-existent user returns same error with comparable timing, unverified email returns forbidden, bcrypt comparison timing is constant
- [x] T064 [P] [US2] Write HTTP handler tests for POST /api/sessions in `features/auth/handler_test.go` — 200 with accessToken/refreshToken/expiresIn, 401 for bad credentials, 403 for unverified, rate limiting 429

### Implementation

- [x] T065 [US2] Implement auth repository interface and sqlx implementation in `features/auth/repository.go` — CreateRefreshToken(ctx, token), GetRefreshTokenByHash(ctx, hash), RevokeRefreshTokenFamily(ctx, familyID), RevokeAllUserRefreshTokens(ctx, userID), GetRefreshTokenByFamilyLatest(ctx, familyID)
- [x] T066 [US2] Implement JWT service (access token creation + validation) as shared utility used by auth — create HS256-signed JWT with sub, iat, exp, jti, iss, aud claims. Validate and parse tokens. Use golang-jwt/jwt/v5. Place in `pkg/jwt/jwt.go` with fx.Module in `pkg/jwt/module.go`
- [x] T067 [US2] Implement LoginCommand and LoginCommandHandler in `features/auth/login_command.go` — get user by email (if not found, still run bcrypt compare against dummy hash for constant timing), verify password via bcrypt, check emailVerified (return ForbiddenError if false), generate access token (JWT, ~15 min), generate refresh token (256-bit random, new family_id), store refresh token hash, return token pair. Anti-enumeration: identical error for not-found and wrong-password.
- [x] T068 [US2] Implement HTTP handler for POST /api/sessions in `features/auth/handler.go` — parse login request, dispatch LoginCommand, translate to JSON response with self, kind, accessToken, refreshToken, tokenType, expiresIn. Apply rate limiting (10/min).
- [x] T069 [US2] Implement fx.Module registration in `features/auth/module.go` — provide repository, login command bus, HTTP handler. Wire into router.
- [x] T070 [US2] Wire auth feature module into `cmd/api/main.go` — add to fx.Options()
- [x] T071 [P] [US2] Write tests for JWT service (create, validate, expired rejection, tampered rejection, wrong issuer/audience rejection) in `pkg/jwt/jwt_test.go`

**Checkpoint**: `go test ./features/auth/...` and `go test ./pkg/jwt/...` green. Full register → verify → login flow works via curl. Unverified login returns 403. Wrong password indistinguishable from non-existent user.

---

## Phase 5: US3 — Access Protected Resources (Priority: P1)

**Goal**: An authenticated user can access GET /api/users/me with a valid access token and see their profile.

**Independent Test**: Login → use access token → GET /api/users/me returns profile with id, email, emailVerified, createdAt. Missing/expired/tampered tokens return 401.

### Tests

- [x] T072 [P] [US3] Write tests for GetProfileQuery handler in `features/user/get_profile_query_test.go` — valid user ID returns profile, non-existent user returns not-found error
- [x] T073 [P] [US3] Write HTTP handler tests for GET /api/users/me in `features/user/handler_test.go` — 200 with self/kind/id/email/emailVerified/createdAt, 401 for missing/expired/tampered token

### Implementation

- [x] T074 [US3] Implement GetProfileQuery and GetProfileQueryHandler in `features/user/get_profile_query.go` — extract user ID from context, query repository by ID, return user profile (excluding password_hash)
- [x] T075 [US3] Add GET /api/users/me route to user HTTP handler in `features/user/handler.go` — protected by auth middleware, extract user ID from context, dispatch GetProfileQuery, respond with self/kind/id/email/emailVerified/createdAt
- [x] T076 [US3] Update user fx.Module in `features/user/module.go` to wire GetProfileQuery bus

**Checkpoint**: `go test ./features/user/...` green. Full register → verify → login → GET /api/users/me flow works. Auth middleware rejects bad tokens.

---

## Phase 6: US4 — Token Refresh with Rotation (Priority: P2)

**Goal**: A user can exchange a valid refresh token for new access + refresh tokens. Old refresh token is invalidated. Reuse of rotated token revokes entire family.

**Independent Test**: Login → use refresh token → get new tokens → old refresh token fails. Replay rotated token → entire family revoked.

### Tests

- [x] T077 [P] [US4] Write tests for RefreshTokenCommand handler in `features/auth/refresh_command_test.go` — valid token rotates (new pair returned, old revoked), expired token rejected, already-revoked token triggers family revocation (reuse detection), invalid token rejected, timing-safe comparison used

### Implementation

- [x] T078 [US4] Implement RefreshTokenCommand and RefreshTokenCommandHandler in `features/auth/refresh_command.go` — hash submitted token, look up in DB via timing-safe comparison, check not expired and not revoked. If revoked (reuse detection): revoke entire family, return UnauthorizedError. If valid: revoke old token, create new refresh token with same family_id, create new access token, return token pair.
- [x] T079 [US4] Add POST /api/tokens route to auth HTTP handler in `features/auth/handler.go` — parse refresh token from body, dispatch RefreshTokenCommand, respond with new token pair
- [x] T080 [US4] Update auth fx.Module in `features/auth/module.go` to wire RefreshTokenCommand bus

**Checkpoint**: `go test ./features/auth/...` green. Login → refresh → get new tokens. Old token fails. Replayed rotated token revokes family.

---

## Phase 7: US5 — Logout (Priority: P2)

**Goal**: An authenticated user can log out, revoking their entire refresh token family.

**Independent Test**: Login → logout → refresh with any token from that family fails. Logout is idempotent.

### Tests

- [ ] T081 [P] [US5] Write tests for LogoutCommand handler in `features/auth/logout_command_test.go` — revokes entire family, idempotent (second logout succeeds), refresh after logout fails

### Implementation

- [ ] T082 [US5] Implement LogoutCommand and LogoutCommandHandler in `features/auth/logout_command.go` — extract user ID from context, find active refresh token family for user's current session, revoke entire family. Idempotent: return success even if already revoked.
- [ ] T083 [US5] Add DELETE /api/sessions/current route to auth HTTP handler in `features/auth/handler.go` — protected by auth middleware, dispatch LogoutCommand, respond with 204 No Content
- [ ] T084 [US5] Update auth fx.Module in `features/auth/module.go` to wire LogoutCommand bus

**Checkpoint**: `go test ./features/auth/...` green. Login → logout → refresh fails. Second logout returns 204.

---

## Phase 8: US9 — Password Reset (Priority: P2)

**Goal**: A user can request a password reset email, then use the token to set a new password. All refresh token families are revoked on reset.

**Independent Test**: Create verified user → forgot-password → extract reset token → reset-password with new password → login with new password succeeds → login with old password fails → all refresh tokens revoked.

### Tests

- [ ] T085 [P] [US9] Write tests for ForgotPasswordCommand handler in `features/user/forgot_password_command_test.go` — sends reset email for verified user, does NOT send for unverified user, returns same result for non-existent email (no enumeration), invalidates previous reset tokens, rate limiting
- [ ] T086 [P] [US9] Write tests for ResetPasswordCommand handler in `features/user/reset_password_command_test.go` — valid token resets password (bcrypt hash updated), expired token rejected, used token rejected, weak password rejected with validation errors, all refresh token families revoked
- [ ] T087 [P] [US9] Write HTTP handler tests for POST /api/users/forgot-password and POST /api/users/reset-password in `features/user/handler_test.go`

### Implementation

- [ ] T088 [US9] Add password reset token repository methods to `features/user/repository.go` — CreatePasswordResetToken(ctx, token), GetPasswordResetTokenByHash(ctx, hash), InvalidateUserPasswordResetTokens(ctx, userID)
- [ ] T089 [US9] Implement ForgotPasswordCommand and ForgotPasswordCommandHandler in `features/user/forgot_password_command.go` — look up user by email, if verified: invalidate previous reset tokens, generate new 256-bit token, store hash (1-hour expiry), send reset email. If unverified or not found: do dummy work for constant timing, return success. Always return same response.
- [ ] T090 [US9] Implement ResetPasswordCommand and ResetPasswordCommandHandler in `features/user/reset_password_command.go` — hash submitted token, look up in DB, check not expired and not used, validate new password strength, bcrypt hash new password, update user password_hash, mark token used, revoke ALL refresh token families for user (call auth repository), return success. All error cases return identical UnauthorizedError.
- [ ] T091 [US9] Add POST /api/users/forgot-password and POST /api/users/reset-password routes to user HTTP handler in `features/user/handler.go` — apply rate limiting (5/min forgot, 10/min reset)
- [ ] T092 [US9] Update user fx.Module in `features/user/module.go` to wire ForgotPasswordCommand and ResetPasswordCommand buses

**Checkpoint**: `go test ./features/user/...` green. Forgot-password → reset-password flow works. Old password fails. All refresh tokens revoked. Non-existent email returns identical response.

---

## Phase 9: US6 — Health and Readiness Checks (Priority: P3)

**Goal**: /healthz returns 200 always. /readyz returns 200 if database reachable, 503 otherwise.

**Independent Test**: GET /healthz → 200. GET /readyz with DB up → 200. GET /readyz with DB down → 503.

### Tests

- [ ] T093 [P] [US6] Write tests for ReadinessQuery handler and HTTP handlers in `features/health/handler_test.go` — healthz returns 200 with status ok, readyz returns 200 when DB pings, readyz returns 503 when DB unreachable

### Implementation

- [ ] T094 [US6] Implement ReadinessQuery and ReadinessQueryHandler in `features/health/readiness_query.go` — ping database, return ready/unavailable status
- [ ] T095 [US6] Implement HTTP handlers for GET /healthz and GET /readyz in `features/health/handler.go` — healthz returns static 200 with self/kind/status, readyz dispatches ReadinessQuery
- [ ] T096 [US6] Implement fx.Module in `features/health/module.go` — provide handler, wire into router at root level (not under /api/)
- [ ] T097 [US6] Wire health feature module into `cmd/api/main.go`

**Checkpoint**: `go test ./features/health/...` green. /healthz and /readyz work via curl.

---

## Phase 10: US7 — Developer Onboarding (Priority: P3)

**Goal**: A new developer goes from clone to successful login in under 5 minutes following the README.

**Independent Test**: Follow README from clean clone, time to first POST /api/sessions response.

### Implementation

- [ ] T098 [US7] Create `README.md` with: project overview, prerequisites (Docker, Task, Go), quick start (clone, cp .env.example .env, task dev or docker compose up), curl examples covering full register → verify → login → profile flow, architecture overview, project structure, development workflow (task targets), testing instructions
- [ ] T099 [US7] Configure Swagger annotations on all HTTP handlers and generate OpenAPI spec via swaggo/swag. Serve Swagger UI at /swagger/ in dev mode via swaggo/http-swagger. Add `swagger` target to Taskfile.yml.
- [ ] T100 [US7] Implement background token cleanup goroutine in `pkg/cleanup/cleanup.go` — periodically purge expired/revoked refresh tokens, expired/used verification tokens, expired/used password reset tokens. Configurable interval (default: hourly). Respect graceful shutdown. Register as fx.Module with lifecycle hooks in `pkg/cleanup/module.go`. Wire into `cmd/api/main.go`.

**Checkpoint**: `docker compose up` starts everything. README curl examples work end-to-end. Swagger UI accessible at /swagger/. Token cleanup runs on schedule.

---

## Phase 11: Polish & Cross-Cutting Concerns

**Purpose**: Security hardening, log redaction verification, and final integration validation

- [ ] T101 [P] Implement log redaction tests — verify no passwords, tokens (access, refresh, verification, reset), email content, or full Authorization headers appear in log output. Write in `pkg/middleware/logging_test.go` and `pkg/email/email_test.go`.
- [ ] T102 [P] Implement anti-enumeration timing tests — verify login with non-existent user and wrong password take comparable time (within 100ms). Write in `features/auth/handler_test.go`.
- [ ] T103 [P] Implement 405 Method Not Allowed tests — verify all defined endpoints return 405 with Allow header for unsupported methods. Write in `pkg/server/server_test.go`.
- [ ] T104 Write end-to-end integration test covering the complete flow: register → verify email → login → access profile → refresh token → logout → verify refresh revoked → forgot password → reset password → login with new password. Write in `tests/integration/auth_flow_test.go`.
- [ ] T105 [P] Verify all JSON responses include `self` and `kind` properties, use camelCase, and use ISO 8601 dates — add schema validation assertions to existing handler tests
- [ ] T106 [P] Verify config startup validation — test that app refuses to boot with missing JWT_SECRET, missing RESEND_API_KEY in prod, known insecure defaults. Write in `pkg/config/config_test.go` (extend T043).
- [ ] T107 Run `task lint` and fix all golangci-lint issues across the entire codebase
- [ ] T108 Run full quickstart.md validation scenarios manually against running Docker stack — verify all 7 scenarios pass

**Checkpoint**: All tests pass (`go test ./...`). Lint clean. Quickstart validated. Application is production-ready.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **US1+US8+US10 (Phase 3)**: Depends on Foundational — MVP registration flow
- **US2 (Phase 4)**: Depends on Foundational + needs user repository from Phase 3
- **US3 (Phase 5)**: Depends on Phase 4 (needs login to get access token)
- **US4 (Phase 6)**: Depends on Phase 4 (needs login to get refresh token)
- **US5 (Phase 7)**: Depends on Phase 4 (needs login to get access token)
- **US9 (Phase 8)**: Depends on Phase 3 (needs verified user) + Phase 4 (revokes refresh tokens)
- **US6 (Phase 9)**: Depends on Foundational only — can run in parallel with any user story
- **US7 (Phase 10)**: Depends on all user stories being complete
- **Polish (Phase 11)**: Depends on all phases complete

### User Story Dependencies

```text
Phase 1 (Setup)
  └─→ Phase 2 (Foundational) ──BLOCKS──┐
                                        ├─→ Phase 3 (US1+US8+US10: Register/Verify/Welcome) ──┐
                                        │                                                       ├─→ Phase 4 (US2: Login) ──┐
                                        │                                                       │                           ├─→ Phase 5 (US3: Profile)
                                        │                                                       │                           ├─→ Phase 6 (US4: Refresh)
                                        │                                                       │                           ├─→ Phase 7 (US5: Logout)
                                        │                                                       ├─→ Phase 8 (US9: Password Reset)
                                        ├─→ Phase 9 (US6: Health) ── independent ──────────────┤
                                        └──────────────────────────────────────────────────────→ Phase 10 (US7: Onboarding)
                                                                                                  └─→ Phase 11 (Polish)
```

### Within Each User Story

- Tests MUST be written first and FAIL before implementation
- Repository before command/query handlers
- Command/query handlers before HTTP handlers
- HTTP handlers before fx.Module wiring
- fx.Module wiring before main.go integration

### Parallel Opportunities

**Phase 2** (highest parallelism):
- T009-T015: All CQRS components can be built in parallel
- T016-T020: All CQRS tests in parallel
- T021-T025: All infrastructure modules in parallel (after CQRS)
- T026-T034: All HTTP infrastructure in parallel
- T035-T042: All migrations in parallel
- T043-T049: All infrastructure tests in parallel

**Across phases** (after Foundational):
- Phase 9 (Health) can run in parallel with any user story phase
- Phases 5, 6, 7 can run in parallel (all depend on Phase 4 but not each other)

---

## Parallel Example: Phase 2 Foundational

```bash
# Wave 1: CQRS package (all parallel)
Agent: "Implement CommandBus in pkg/cqrs/command.go"
Agent: "Implement QueryBus in pkg/cqrs/query.go"
Agent: "Implement EventBus in pkg/cqrs/event.go"
Agent: "Implement domain errors in pkg/cqrs/errors.go"
Agent: "Implement logging middleware in pkg/cqrs/middleware/logging.go"
Agent: "Implement validation middleware in pkg/cqrs/middleware/validation.go"
Agent: "Implement recovery middleware in pkg/cqrs/middleware/recovery.go"

# Wave 2: Infrastructure + CQRS tests (all parallel)
Agent: "Implement config in pkg/config/"
Agent: "Implement database in pkg/database/"
Agent: "Implement redis in pkg/redis/"
Agent: "Implement email service in pkg/email/"
Agent: "Implement token utils in pkg/token/"
Agent: "Write CQRS tests in pkg/cqrs/*_test.go"

# Wave 3: HTTP infra + migrations (all parallel)
Agent: "Implement response helpers in pkg/response/"
Agent: "Implement HTTP middleware in pkg/middleware/"
Agent: "Implement server in pkg/server/"
Agent: "Create all migrations in migrations/"
```

---

## Implementation Strategy

### MVP First (Phase 1 → 2 → 3 → 4 → 5)

1. Complete Phase 1: Setup — project compiles
2. Complete Phase 2: Foundational — infrastructure ready
3. Complete Phase 3: US1+US8+US10 — registration + verification working
4. Complete Phase 4: US2 — login working
5. Complete Phase 5: US3 — profile endpoint working
6. **STOP and VALIDATE**: Register → verify → login → profile works end-to-end
7. Deploy/demo MVP

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1+US8+US10 → Registration + verification (MVP core)
3. Add US2 → Login (MVP complete)
4. Add US3 → Profile (MVP+)
5. Add US4+US5 → Token refresh + logout (session management)
6. Add US9 → Password reset (account recovery)
7. Add US6 → Health checks (operational readiness)
8. Add US7 → Developer onboarding (documentation)
9. Polish → Production-ready

### Parallel Team Strategy

With multiple developers after Foundational is complete:

- **Developer A**: Phase 3 (Register/Verify/Welcome) → Phase 4 (Login) → Phase 5 (Profile)
- **Developer B**: Phase 9 (Health) → Phase 6 (Token Refresh) → Phase 7 (Logout)
- **Developer C**: Phase 8 (Password Reset) → Phase 10 (Onboarding) → Phase 11 (Polish)

---

## Notes

- [P] tasks = different files, no dependencies on incomplete tasks
- [Story] label maps task to specific user story for traceability
- TDD enforced: write tests first, verify they fail, then implement
- Commit after each task or logical group
- Stop at any checkpoint to validate independently
- `{{MODULE_PATH}}` and `{{PROJECT_NAME}}` must be provided before T001

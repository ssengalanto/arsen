# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

This is the Go API workspace. For monorepo-wide commands and architecture overview, see the root [CLAUDE.md](../../CLAUDE.md).

## Commands (run from repo root)

```bash
task api:test                        # go test ./...
task api:test -- -run TestName       # Single test
task api:test -- -v ./features/auth/ # Verbose, one package
task api:lint                        # golangci-lint
task api:lint-fix                    # Auto-fix
task api:build                       # Binary → apps/api/bin/arsen
task api:swagger                     # Regenerate Swagger docs (swag init)
task api:migrate-up                  # Apply pending migrations
task api:migrate-down                # Rollback last migration
task api:migrate-create -- name      # New migration pair
```

## Request Lifecycle

1. Chi router applies global middleware: RequestID → Logging → Recovery → CORS
2. Per-route middleware applied inline (e.g., `middleware.Auth`, `middleware.RateLimit`)
3. Handler decodes JSON body into a request struct
4. Handler dispatches a command/query via its typed bus: `bus.Dispatch(ctx, cmd)` or `bus.Ask(ctx, query)`
5. CQRS middleware chain runs: Recovery → Logging → Validation (calls `cmd.Validate()`)
6. Command/query handler executes business logic, returns result or typed error
7. Handler calls `response.HandleError(w, r, err)` which maps CQRS errors → RFC 9457 status codes:
   - `*cqrs.ValidationError` → 400, `*cqrs.UnauthorizedError` → 401, `*cqrs.ForbiddenError` → 403, `*cqrs.NotFoundError` → 404, `*cqrs.ConflictError` → 409, anything else → 500

## Adding a New Feature

1. Create `features/<name>/` with these files:
   - `module.go` — `fx.Module` that provides repository, handlers, command/query buses with middleware, and invokes route registration
   - `handler.go` — HTTP handlers + `RegisterRoutes(r chi.Router, jwtService *jwt.Service)`
   - `<action>_command.go` — Command struct with `Validate() error`, handler implementing `cqrs.CommandHandler[C, R]`
   - `repository.go` — Interface + `SQLRepository` implementation
2. Wire command buses in `module.go` with the standard middleware triple: `Recovery`, `Logging`, `Validation`
3. Register `<name>.Module` in `cmd/api/main.go`'s `fx.New()` call

Use `cqrs.Unit` as the return type for commands with no meaningful result (logout, resend-verification, etc.).

## Adding a New Command/Query to an Existing Feature

1. Create `<action>_command.go` with the command struct + handler
2. Add `fx.Provide(NewXCommandHandler)` in `module.go`
3. Add a new `fx.Provide` for the `*cqrs.CommandBus[XCommand, XResult]` with middleware
4. Inject the bus into the handler, add the HTTP route in `RegisterRoutes`

## Cross-Feature Data Access

Features must not import another feature's `Repository` interface. When feature A needs data owned by feature B, feature B exposes a query through its `QueryBus`:

1. In the owning feature, create `get_<entity>_by_<field>_query.go` with the query struct, result struct, and handler
2. Wire the handler and `QueryBus` in the owning feature's `module.go`
3. In the consuming feature, inject `*cqrs.QueryBus[owner.XQuery, *owner.XResult]` and call `.Ask(ctx, query)`

Example: auth needs user data for login. User feature exposes `GetUserByEmailQuery` via its query bus. Auth injects `*cqrs.QueryBus[user.GetUserByEmailQuery, *user.GetUserByEmailResult]` — never `user.Repository`.

In tests, mock the `cqrs.QueryHandler` interface and wrap it in a real `QueryBus`:
```go
mock := &mockUserByEmailHandler{users: make(map[string]*user.User)}
bus := cqrs.NewQueryBus[user.GetUserByEmailQuery, *user.GetUserByEmailResult](mock)
handler := NewLoginCommandHandler(bus, authRepo, jwtSvc, cfg)
```

## Anti-Enumeration Pattern

Register, forgot-password, and resend-verification endpoints must return identical responses regardless of whether the email exists. In handlers, catch `ConflictError`/`NotFoundError` and normalize to a generic success response. See `features/user/handler.go` for the pattern.

## Event Bus

For post-command side effects (e.g., sending a welcome email after verification), use `cqrs.EventBus[E]`. Register event handlers in `module.go` and publish events from command handlers. See `features/user/module.go` for the `EmailVerifiedEvent` example.

## Auth Context

Protected routes use `middleware.Auth(jwtService)`. Inside handlers, extract the user ID with:
```go
userID, ok := middleware.GetUserID(r.Context())
```

## Rate Limiting

Applied per-route via `middleware.RateLimit(rps, burst)`. The first argument is requests-per-second (e.g., `10.0/60.0` for 10 per minute), second is burst capacity.

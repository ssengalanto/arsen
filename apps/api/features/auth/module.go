package auth

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"

	"arsen/pkg/cqrs"
	cqrsmw "arsen/pkg/cqrs/middleware"
	"arsen/pkg/server"
)

// Module wires the auth feature into the fx dependency graph.
var Module = fx.Module("auth",
	// Provide the repository (concrete -> interface).
	fx.Provide(func(db *sqlx.DB) Repository {
		return NewSQLRepository(db)
	}),

	// Provide the command handlers.
	fx.Provide(NewLoginCommandHandler),
	fx.Provide(NewRefreshTokenCommandHandler),
	fx.Provide(NewLogoutCommandHandler),

	// Provide the HTTP handler.
	fx.Provide(NewHandler),

	// Command buses with middleware.
	fx.Provide(func(h *LoginCommandHandler) *cqrs.CommandBus[LoginCommand, *LoginResult] {
		return cqrs.NewCommandBus[LoginCommand, *LoginResult](h,
			cqrsmw.Recovery[LoginCommand, *LoginResult](),
			cqrsmw.Logging[LoginCommand, *LoginResult](slog.Default()),
			cqrsmw.Validation[LoginCommand, *LoginResult](),
		)
	}),
	fx.Provide(func(h *RefreshTokenCommandHandler) *cqrs.CommandBus[RefreshTokenCommand, *RefreshTokenResult] {
		return cqrs.NewCommandBus[RefreshTokenCommand, *RefreshTokenResult](h,
			cqrsmw.Recovery[RefreshTokenCommand, *RefreshTokenResult](),
			cqrsmw.Logging[RefreshTokenCommand, *RefreshTokenResult](slog.Default()),
			cqrsmw.Validation[RefreshTokenCommand, *RefreshTokenResult](),
		)
	}),
	fx.Provide(func(h *LogoutCommandHandler) *cqrs.CommandBus[LogoutCommand, cqrs.Unit] {
		return cqrs.NewCommandBus[LogoutCommand, cqrs.Unit](h,
			cqrsmw.Recovery[LogoutCommand, cqrs.Unit](),
			cqrsmw.Logging[LogoutCommand, cqrs.Unit](slog.Default()),
			cqrsmw.Validation[LogoutCommand, cqrs.Unit](),
		)
	}),

	// Wire routes into the server.
	fx.Invoke(func(srv *server.Server, h *Handler) {
		h.RegisterRoutes(srv.Router, srv.JWTService)
	}),
)

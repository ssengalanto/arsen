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

	// Provide the command handler.
	fx.Provide(NewLoginCommandHandler),

	// Provide the HTTP handler.
	fx.Provide(NewHandler),

	// Command bus with middleware.
	fx.Provide(func(h *LoginCommandHandler) *cqrs.CommandBus[LoginCommand, *LoginResult] {
		return cqrs.NewCommandBus[LoginCommand, *LoginResult](h,
			cqrsmw.Recovery[LoginCommand, *LoginResult](),
			cqrsmw.Logging[LoginCommand, *LoginResult](slog.Default()),
			cqrsmw.Validation[LoginCommand, *LoginResult](),
		)
	}),

	// Wire routes into the server.
	fx.Invoke(func(srv *server.Server, h *Handler) {
		h.RegisterRoutes(srv.Router)
	}),
)

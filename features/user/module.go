package user

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"

	"arsen/pkg/cqrs"
	cqrsmw "arsen/pkg/cqrs/middleware"
	"arsen/pkg/server"
)

// Module wires the user feature into the fx dependency graph.
var Module = fx.Module("user",
	// Provide the repository (concrete → interface).
	fx.Provide(func(db *sqlx.DB) Repository {
		return NewSQLRepository(db)
	}),

	// Provide command handlers.
	fx.Provide(
		NewRegisterCommandHandler,
		NewVerifyEmailCommandHandler,
		NewResendVerificationCommandHandler,
		NewWelcomeEmailHandler,
	),

	// Provide the HTTP handler.
	fx.Provide(NewHandler),

	// Command buses with middleware.
	fx.Provide(func(h *RegisterCommandHandler) *cqrs.CommandBus[RegisterCommand, *RegisterResult] {
		return cqrs.NewCommandBus[RegisterCommand, *RegisterResult](h,
			cqrsmw.Recovery[RegisterCommand, *RegisterResult](),
			cqrsmw.Logging[RegisterCommand, *RegisterResult](slog.Default()),
			cqrsmw.Validation[RegisterCommand, *RegisterResult](),
		)
	}),
	fx.Provide(func(h *VerifyEmailCommandHandler) *cqrs.CommandBus[VerifyEmailCommand, *VerifyEmailResult] {
		return cqrs.NewCommandBus[VerifyEmailCommand, *VerifyEmailResult](h,
			cqrsmw.Recovery[VerifyEmailCommand, *VerifyEmailResult](),
			cqrsmw.Logging[VerifyEmailCommand, *VerifyEmailResult](slog.Default()),
			cqrsmw.Validation[VerifyEmailCommand, *VerifyEmailResult](),
		)
	}),
	fx.Provide(func(h *ResendVerificationCommandHandler) *cqrs.CommandBus[ResendVerificationCommand, cqrs.Unit] {
		return cqrs.NewCommandBus[ResendVerificationCommand, cqrs.Unit](h,
			cqrsmw.Recovery[ResendVerificationCommand, cqrs.Unit](),
			cqrsmw.Logging[ResendVerificationCommand, cqrs.Unit](slog.Default()),
			cqrsmw.Validation[ResendVerificationCommand, cqrs.Unit](),
		)
	}),

	// Event bus for post-verification events.
	fx.Provide(func(h *WelcomeEmailHandler) *cqrs.EventBus[EmailVerifiedEvent] {
		bus := cqrs.NewEventBus[EmailVerifiedEvent]()
		bus.Register(h)
		return bus
	}),

	// Wire routes into the server.
	fx.Invoke(func(srv *server.Server, h *Handler) {
		h.RegisterRoutes(srv.Router)
	}),
)

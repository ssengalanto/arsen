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

	// Provide the refresh token revoker (avoids circular import with auth).
	fx.Provide(func(db *sqlx.DB) RefreshTokenRevoker { //nolint:gocritic // lambda needed for interface conversion
		return NewSQLRefreshTokenRevoker(db)
	}),

	// Provide command/query handlers.
	fx.Provide(
		NewRegisterCommandHandler,
		NewVerifyEmailCommandHandler,
		NewResendVerificationCommandHandler,
		NewWelcomeEmailHandler,
		NewGetProfileQueryHandler,
		NewForgotPasswordCommandHandler,
		NewResetPasswordCommandHandler,
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
	fx.Provide(func(h *ForgotPasswordCommandHandler) *cqrs.CommandBus[ForgotPasswordCommand, cqrs.Unit] {
		return cqrs.NewCommandBus[ForgotPasswordCommand, cqrs.Unit](h,
			cqrsmw.Recovery[ForgotPasswordCommand, cqrs.Unit](),
			cqrsmw.Logging[ForgotPasswordCommand, cqrs.Unit](slog.Default()),
			cqrsmw.Validation[ForgotPasswordCommand, cqrs.Unit](),
		)
	}),
	fx.Provide(func(h *ResetPasswordCommandHandler) *cqrs.CommandBus[ResetPasswordCommand, cqrs.Unit] {
		return cqrs.NewCommandBus[ResetPasswordCommand, cqrs.Unit](h,
			cqrsmw.Recovery[ResetPasswordCommand, cqrs.Unit](),
			cqrsmw.Logging[ResetPasswordCommand, cqrs.Unit](slog.Default()),
			cqrsmw.Validation[ResetPasswordCommand, cqrs.Unit](),
		)
	}),

	// Query bus with middleware.
	fx.Provide(func(h *GetProfileQueryHandler) *cqrs.QueryBus[GetProfileQuery, *GetProfileResult] {
		return cqrs.NewQueryBus[GetProfileQuery, *GetProfileResult](h,
			cqrsmw.Recovery[GetProfileQuery, *GetProfileResult](),
			cqrsmw.Logging[GetProfileQuery, *GetProfileResult](slog.Default()),
			cqrsmw.Validation[GetProfileQuery, *GetProfileResult](),
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
		h.RegisterRoutes(srv.Router, srv.JWTService)
	}),
)

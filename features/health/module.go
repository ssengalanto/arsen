package health

import (
	"log/slog"

	"go.uber.org/fx"

	"arsen/pkg/cqrs"
	cqrsmw "arsen/pkg/cqrs/middleware"
	"arsen/pkg/server"
)

var Module = fx.Module("health",
	fx.Provide(NewReadinessQueryHandler),
	fx.Provide(NewHandler),

	fx.Provide(func(h *ReadinessQueryHandler) *cqrs.QueryBus[ReadinessQuery, *ReadinessResult] {
		return cqrs.NewQueryBus[ReadinessQuery, *ReadinessResult](h,
			cqrsmw.Recovery[ReadinessQuery, *ReadinessResult](),
			cqrsmw.Logging[ReadinessQuery, *ReadinessResult](slog.Default()),
		)
	}),

	fx.Invoke(func(srv *server.Server, h *Handler) {
		h.RegisterRoutes(srv.Router)
	}),
)

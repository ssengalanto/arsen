package cleanup

import (
	"context"

	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"

	"arsen/pkg/config"
)

var Module = fx.Module("cleanup",
	fx.Provide(func(db *sqlx.DB, cfg *config.Config) *Service {
		return NewService(db, cfg.Cleanup.Interval)
	}),
	fx.Invoke(func(lc fx.Lifecycle, svc *Service) {
		ctx, cancel := context.WithCancel(context.Background())
		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				go svc.Start(ctx)
				return nil
			},
			OnStop: func(_ context.Context) error {
				cancel()
				return nil
			},
		})
	}),
)

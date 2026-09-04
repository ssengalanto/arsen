package database

import (
	"context"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"
)

var Module = fx.Module("database",
	fx.Provide(New),
	fx.Invoke(func(lc fx.Lifecycle, db *sqlx.DB) {
		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				slog.Info("database connected")
				return nil
			},
			OnStop: func(_ context.Context) error {
				return db.Close()
			},
		})
	}),
)

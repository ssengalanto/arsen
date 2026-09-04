package redis

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

var Module = fx.Module("redis",
	fx.Provide(New),
	fx.Invoke(func(lc fx.Lifecycle, client *redis.Client) {
		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				slog.Info("redis connected")
				return nil
			},
			OnStop: func(_ context.Context) error {
				return client.Close()
			},
		})
	}),
)

package middleware

import (
	"context"
	"log/slog"
	"reflect"
	"time"
)

func Logging[C any, R any](logger *slog.Logger) func(ctx context.Context, input C, next func(context.Context, C) (R, error)) (R, error) {
	return func(ctx context.Context, input C, next func(context.Context, C) (R, error)) (R, error) {
		name := reflect.TypeOf(input).Name()
		start := time.Now()

		result, err := next(ctx, input)
		duration := time.Since(start)

		if err != nil {
			logger.ErrorContext(ctx, "handler failed",
				slog.String("type", name),
				slog.Duration("duration", duration),
				slog.String("error", err.Error()),
			)
		} else {
			logger.InfoContext(ctx, "handler succeeded",
				slog.String("type", name),
				slog.Duration("duration", duration),
			)
		}

		return result, err
	}
}

package middleware

import (
	"context"
	"fmt"
)

func Recovery[C any, R any]() func(ctx context.Context, input C, next func(context.Context, C) (R, error)) (R, error) {
	return func(ctx context.Context, input C, next func(context.Context, C) (R, error)) (R, error) {
		var (
			result R
			err    error
		)

		func() {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic recovered: %v", r)
				}
			}()
			result, err = next(ctx, input)
		}()

		return result, err
	}
}

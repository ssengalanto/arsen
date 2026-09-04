package middleware

import "context"

type Validatable interface {
	Validate() error
}

func Validation[C Validatable, R any]() func(ctx context.Context, input C, next func(context.Context, C) (R, error)) (R, error) {
	return func(ctx context.Context, input C, next func(context.Context, C) (R, error)) (R, error) {
		if err := input.Validate(); err != nil {
			var zero R
			return zero, err
		}
		return next(ctx, input)
	}
}

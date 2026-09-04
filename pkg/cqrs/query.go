package cqrs

import "context"

type QueryHandler[Q any, R any] interface {
	Handle(ctx context.Context, query Q) (R, error)
}

type QueryMiddleware[Q any, R any] func(ctx context.Context, query Q, next func(context.Context, Q) (R, error)) (R, error)

type QueryBus[Q any, R any] struct {
	handler     QueryHandler[Q, R]
	middlewares []QueryMiddleware[Q, R]
}

func NewQueryBus[Q, R any](handler QueryHandler[Q, R], middlewares ...QueryMiddleware[Q, R]) *QueryBus[Q, R] {
	return &QueryBus[Q, R]{
		handler:     handler,
		middlewares: middlewares,
	}
}

func (b *QueryBus[Q, R]) Ask(ctx context.Context, query Q) (R, error) {
	next := b.handler.Handle

	for i := len(b.middlewares) - 1; i >= 0; i-- {
		mw := b.middlewares[i]
		inner := next
		next = func(ctx context.Context, q Q) (R, error) {
			return mw(ctx, q, inner)
		}
	}

	return next(ctx, query)
}

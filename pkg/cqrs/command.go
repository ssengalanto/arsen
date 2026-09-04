package cqrs

import "context"

type Unit = struct{}

type CommandHandler[C any, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}

type CommandMiddleware[C any, R any] func(ctx context.Context, cmd C, next func(context.Context, C) (R, error)) (R, error)

type CommandBus[C any, R any] struct {
	handler     CommandHandler[C, R]
	middlewares []CommandMiddleware[C, R]
}

func NewCommandBus[C, R any](handler CommandHandler[C, R], middlewares ...CommandMiddleware[C, R]) *CommandBus[C, R] {
	return &CommandBus[C, R]{
		handler:     handler,
		middlewares: middlewares,
	}
}

func (b *CommandBus[C, R]) Dispatch(ctx context.Context, cmd C) (R, error) {
	next := b.handler.Handle

	// Wrap in reverse so the first middleware in the slice executes outermost.
	for i := len(b.middlewares) - 1; i >= 0; i-- {
		mw := b.middlewares[i]
		inner := next
		next = func(ctx context.Context, c C) (R, error) {
			return mw(ctx, c, inner)
		}
	}

	return next(ctx, cmd)
}

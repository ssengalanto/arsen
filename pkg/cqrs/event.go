package cqrs

import (
	"context"
	"sync"
)

type EventHandler[E any] interface {
	Handle(ctx context.Context, event E) error
}

type EventBus[E any] struct {
	mu       sync.Mutex
	handlers []EventHandler[E]
}

func NewEventBus[E any]() *EventBus[E] {
	return &EventBus[E]{}
}

func (b *EventBus[E]) Register(handler EventHandler[E]) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = append(b.handlers, handler)
}

func (b *EventBus[E]) Publish(ctx context.Context, event E) error {
	b.mu.Lock()
	handlers := make([]EventHandler[E], len(b.handlers))
	copy(handlers, b.handlers)
	b.mu.Unlock()

	for _, h := range handlers {
		if err := h.Handle(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (b *EventBus[E]) PublishAsync(ctx context.Context, event E) []error {
	b.mu.Lock()
	handlers := make([]EventHandler[E], len(b.handlers))
	copy(handlers, b.handlers)
	b.mu.Unlock()

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	wg.Add(len(handlers))
	for _, h := range handlers {
		go func(handler EventHandler[E]) {
			defer wg.Done()
			if err := handler.Handle(ctx, event); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(h)
	}
	wg.Wait()

	return errs
}

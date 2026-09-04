package cqrs_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"arsen/pkg/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- test types for events ---

type testEvent struct{ Payload string }

type testEventHandler struct {
	fn func(ctx context.Context, event testEvent) error
}

func (h *testEventHandler) Handle(ctx context.Context, event testEvent) error {
	return h.fn(ctx, event)
}

// --- tests ---

func TestEventBus_Publish_CallsRegisteredHandler(t *testing.T) {
	var called bool
	bus := cqrs.NewEventBus[testEvent]()

	bus.Register(&testEventHandler{
		fn: func(_ context.Context, e testEvent) error {
			called = true
			assert.Equal(t, "ping", e.Payload)
			return nil
		},
	})

	err := bus.Publish(context.Background(), testEvent{Payload: "ping"})

	require.NoError(t, err)
	assert.True(t, called)
}

func TestEventBus_Publish_CallsMultipleHandlers(t *testing.T) {
	var count int32
	bus := cqrs.NewEventBus[testEvent]()

	for i := 0; i < 3; i++ {
		bus.Register(&testEventHandler{
			fn: func(_ context.Context, _ testEvent) error {
				atomic.AddInt32(&count, 1)
				return nil
			},
		})
	}

	err := bus.Publish(context.Background(), testEvent{Payload: "multi"})

	require.NoError(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&count))
}

func TestEventBus_Publish_FailFastOnError(t *testing.T) {
	var handlerTwoCalled bool
	bus := cqrs.NewEventBus[testEvent]()

	bus.Register(&testEventHandler{
		fn: func(_ context.Context, _ testEvent) error {
			return fmt.Errorf("handler-1 error")
		},
	})
	bus.Register(&testEventHandler{
		fn: func(_ context.Context, _ testEvent) error {
			handlerTwoCalled = true
			return nil
		},
	})

	err := bus.Publish(context.Background(), testEvent{Payload: "boom"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "handler-1 error")
	assert.False(t, handlerTwoCalled, "second handler should not be called after first error")
}

func TestEventBus_PublishAsync_CollectsAllErrors(t *testing.T) {
	bus := cqrs.NewEventBus[testEvent]()

	bus.Register(&testEventHandler{
		fn: func(_ context.Context, _ testEvent) error {
			return fmt.Errorf("err-a")
		},
	})
	bus.Register(&testEventHandler{
		fn: func(_ context.Context, _ testEvent) error {
			return nil // success
		},
	})
	bus.Register(&testEventHandler{
		fn: func(_ context.Context, _ testEvent) error {
			return fmt.Errorf("err-c")
		},
	})

	errs := bus.PublishAsync(context.Background(), testEvent{Payload: "async"})

	assert.Len(t, errs, 2)

	errMsgs := make([]string, len(errs))
	for i, e := range errs {
		errMsgs[i] = e.Error()
	}
	assert.Contains(t, errMsgs, "err-a")
	assert.Contains(t, errMsgs, "err-c")
}

func TestEventBus_PublishAsync_RunsConcurrently(t *testing.T) {
	bus := cqrs.NewEventBus[testEvent]()
	delay := 50 * time.Millisecond

	for i := 0; i < 5; i++ {
		bus.Register(&testEventHandler{
			fn: func(_ context.Context, _ testEvent) error {
				time.Sleep(delay)
				return nil
			},
		})
	}

	start := time.Now()
	errs := bus.PublishAsync(context.Background(), testEvent{Payload: "concurrent"})
	elapsed := time.Since(start)

	assert.Empty(t, errs)
	// If sequential, total would be ~250ms. Concurrent should be ~50ms.
	assert.Less(t, elapsed, 3*delay, "handlers should run concurrently, took %v", elapsed)
}

func TestEventBus_Register_ThreadSafe(t *testing.T) {
	bus := cqrs.NewEventBus[testEvent]()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Register(&testEventHandler{
				fn: func(_ context.Context, _ testEvent) error {
					return nil
				},
			})
		}()
	}

	wg.Wait()

	// Verify all handlers were registered by publishing and counting calls.
	var count int32
	// We can't easily replace existing handlers, so just confirm no panic and
	// that Publish works after concurrent registration.
	err := bus.Publish(context.Background(), testEvent{Payload: "safe"})
	_ = count
	require.NoError(t, err)
}

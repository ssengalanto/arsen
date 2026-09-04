package cqrs_test

import (
	"context"
	"fmt"
	"testing"

	"arsen/pkg/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- test types for commands ---

type testCommand struct{ Value string }
type testResult struct{ Output string }

type testCommandHandler struct {
	fn func(ctx context.Context, cmd testCommand) (testResult, error)
}

func (h *testCommandHandler) Handle(ctx context.Context, cmd testCommand) (testResult, error) {
	return h.fn(ctx, cmd)
}

// --- tests ---

func TestCommandBus_Dispatch_CallsHandlerAndReturnsResult(t *testing.T) {
	handler := &testCommandHandler{
		fn: func(_ context.Context, cmd testCommand) (testResult, error) {
			return testResult{Output: "echo:" + cmd.Value}, nil
		},
	}

	bus := cqrs.NewCommandBus[testCommand, testResult](handler)
	result, err := bus.Dispatch(context.Background(), testCommand{Value: "hello"})

	require.NoError(t, err)
	assert.Equal(t, "echo:hello", result.Output)
}

func TestCommandBus_Dispatch_ReturnsHandlerError(t *testing.T) {
	handler := &testCommandHandler{
		fn: func(_ context.Context, _ testCommand) (testResult, error) {
			return testResult{}, fmt.Errorf("handler boom")
		},
	}

	bus := cqrs.NewCommandBus[testCommand, testResult](handler)
	_, err := bus.Dispatch(context.Background(), testCommand{Value: "x"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "handler boom")
}

func TestCommandBus_MiddlewareAppliedInCorrectOrder(t *testing.T) {
	var order []string

	mw1 := func(ctx context.Context, cmd testCommand, next func(context.Context, testCommand) (testResult, error)) (testResult, error) {
		order = append(order, "mw1-before")
		r, err := next(ctx, cmd)
		order = append(order, "mw1-after")
		return r, err
	}

	mw2 := func(ctx context.Context, cmd testCommand, next func(context.Context, testCommand) (testResult, error)) (testResult, error) {
		order = append(order, "mw2-before")
		r, err := next(ctx, cmd)
		order = append(order, "mw2-after")
		return r, err
	}

	handler := &testCommandHandler{
		fn: func(_ context.Context, _ testCommand) (testResult, error) {
			order = append(order, "handler")
			return testResult{}, nil
		},
	}

	bus := cqrs.NewCommandBus[testCommand, testResult](handler, mw1, mw2)
	_, err := bus.Dispatch(context.Background(), testCommand{})

	require.NoError(t, err)
	assert.Equal(t, []string{
		"mw1-before",
		"mw2-before",
		"handler",
		"mw2-after",
		"mw1-after",
	}, order)
}

func TestCommandBus_UnitReturnType(t *testing.T) {
	bus := cqrs.NewCommandBus[testCommand, cqrs.Unit](&unitCommandHandler{})
	result, err := bus.Dispatch(context.Background(), testCommand{Value: "fire-and-forget"})

	require.NoError(t, err)
	assert.Equal(t, cqrs.Unit{}, result)
}

type unitCommandHandler struct{}

func (h *unitCommandHandler) Handle(_ context.Context, _ testCommand) (cqrs.Unit, error) {
	return cqrs.Unit{}, nil
}

func TestCommandBus_Dispatch_NilHandlerPanics(t *testing.T) {
	// The bus requires a handler at construction time; passing nil should
	// cause a panic on Dispatch because Handle is called on a nil receiver.
	assert.Panics(t, func() {
		bus := cqrs.NewCommandBus[testCommand, testResult](nil)
		_, _ = bus.Dispatch(context.Background(), testCommand{})
	})
}

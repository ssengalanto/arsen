package middleware_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"arsen/pkg/cqrs/middleware"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── Logging middleware tests ────────────────────────────────────────────────

type logCmd struct{ Value string }
type logResult struct{ Output string }

func TestLogging_PassesThroughResult(t *testing.T) {
	logger := slog.Default()
	mw := middleware.Logging[logCmd, logResult](logger)

	next := func(_ context.Context, cmd logCmd) (logResult, error) {
		return logResult{Output: "ok:" + cmd.Value}, nil
	}

	result, err := mw(context.Background(), logCmd{Value: "test"}, next)

	require.NoError(t, err)
	assert.Equal(t, "ok:test", result.Output)
}

func TestLogging_PassesThroughError(t *testing.T) {
	logger := slog.Default()
	mw := middleware.Logging[logCmd, logResult](logger)

	next := func(_ context.Context, _ logCmd) (logResult, error) {
		return logResult{}, fmt.Errorf("inner error")
	}

	_, err := mw(context.Background(), logCmd{Value: "x"}, next)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "inner error")
}

// ─── Validation middleware tests ─────────────────────────────────────────────

type validatable struct {
	valid bool
}

func (v validatable) Validate() error {
	if !v.valid {
		return fmt.Errorf("invalid")
	}
	return nil
}

type valResult struct{ Data string }

func TestValidation_ValidCommandPassesThrough(t *testing.T) {
	mw := middleware.Validation[validatable, valResult]()

	var handlerCalled bool
	next := func(_ context.Context, _ validatable) (valResult, error) {
		handlerCalled = true
		return valResult{Data: "processed"}, nil
	}

	result, err := mw(context.Background(), validatable{valid: true}, next)

	require.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, "processed", result.Data)
}

func TestValidation_InvalidCommandShortCircuits(t *testing.T) {
	mw := middleware.Validation[validatable, valResult]()

	var handlerCalled bool
	next := func(_ context.Context, _ validatable) (valResult, error) {
		handlerCalled = true
		return valResult{Data: "should not happen"}, nil
	}

	_, err := mw(context.Background(), validatable{valid: false}, next)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
	assert.False(t, handlerCalled, "handler should not be called when validation fails")
}

// ─── Recovery middleware tests ───────────────────────────────────────────────

type recCmd struct{ Value string }
type recResult struct{ Output string }

func TestRecovery_NormalExecutionPassesThrough(t *testing.T) {
	mw := middleware.Recovery[recCmd, recResult]()

	next := func(_ context.Context, cmd recCmd) (recResult, error) {
		return recResult{Output: "safe:" + cmd.Value}, nil
	}

	result, err := mw(context.Background(), recCmd{Value: "ok"}, next)

	require.NoError(t, err)
	assert.Equal(t, "safe:ok", result.Output)
}

func TestRecovery_PanicIsRecoveredAsError(t *testing.T) {
	mw := middleware.Recovery[recCmd, recResult]()

	next := func(_ context.Context, _ recCmd) (recResult, error) {
		panic("unexpected crash")
	}

	result, err := mw(context.Background(), recCmd{Value: "boom"}, next)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "panic recovered")
	assert.Contains(t, err.Error(), "unexpected crash")
	assert.Equal(t, recResult{}, result)
}

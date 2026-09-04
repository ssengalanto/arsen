package cqrs_test

import (
	"context"
	"fmt"
	"testing"

	"arsen/pkg/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- test types for queries ---

type testQuery struct{ Term string }
type testQueryResult struct{ Items []string }

type testQueryHandler struct {
	fn func(ctx context.Context, q testQuery) (testQueryResult, error)
}

func (h *testQueryHandler) Handle(ctx context.Context, q testQuery) (testQueryResult, error) {
	return h.fn(ctx, q)
}

// --- tests ---

func TestQueryBus_Ask_CallsHandlerAndReturnsResult(t *testing.T) {
	handler := &testQueryHandler{
		fn: func(_ context.Context, q testQuery) (testQueryResult, error) {
			return testQueryResult{Items: []string{q.Term}}, nil
		},
	}

	bus := cqrs.NewQueryBus[testQuery, testQueryResult](handler)
	result, err := bus.Ask(context.Background(), testQuery{Term: "foo"})

	require.NoError(t, err)
	assert.Equal(t, []string{"foo"}, result.Items)
}

func TestQueryBus_Ask_ReturnsHandlerError(t *testing.T) {
	handler := &testQueryHandler{
		fn: func(_ context.Context, _ testQuery) (testQueryResult, error) {
			return testQueryResult{}, fmt.Errorf("query failed")
		},
	}

	bus := cqrs.NewQueryBus[testQuery, testQueryResult](handler)
	_, err := bus.Ask(context.Background(), testQuery{Term: "x"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "query failed")
}

func TestQueryBus_MiddlewareIsApplied(t *testing.T) {
	var called bool

	mw := func(ctx context.Context, q testQuery, next func(context.Context, testQuery) (testQueryResult, error)) (testQueryResult, error) {
		called = true
		return next(ctx, q)
	}

	handler := &testQueryHandler{
		fn: func(_ context.Context, _ testQuery) (testQueryResult, error) {
			return testQueryResult{Items: []string{"result"}}, nil
		},
	}

	bus := cqrs.NewQueryBus[testQuery, testQueryResult](handler, mw)
	result, err := bus.Ask(context.Background(), testQuery{Term: "bar"})

	require.NoError(t, err)
	assert.True(t, called, "middleware should have been called")
	assert.Equal(t, []string{"result"}, result.Items)
}

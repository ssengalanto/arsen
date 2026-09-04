package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"arsen/pkg/cqrs"
)

// ---------------------------------------------------------------------------
// Mock readiness query handler
// ---------------------------------------------------------------------------

type mockReadinessHandler struct {
	result *ReadinessResult
	err    error
}

func (m *mockReadinessHandler) Handle(_ context.Context, _ ReadinessQuery) (*ReadinessResult, error) {
	return m.result, m.err
}

// ---------------------------------------------------------------------------
// Helper: build a Handler wired to mock, mounted on chi.
// ---------------------------------------------------------------------------

func setupHealthRouter(readinessMock *mockReadinessHandler) http.Handler {
	bus := cqrs.NewQueryBus[ReadinessQuery, *ReadinessResult](readinessMock)
	r := chi.NewRouter()
	h := NewHandler(bus)
	h.RegisterRoutes(r)
	return r
}

// ---------------------------------------------------------------------------
// T093-1: GET /healthz — Liveness probe
// ---------------------------------------------------------------------------

func TestHandler_Healthz_ReturnsOK(t *testing.T) {
	// The liveness endpoint should not touch readiness at all, so the mock
	// can be empty.
	router := setupHealthRouter(&mockReadinessHandler{})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp healthResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	assert.Equal(t, "/healthz", resp.Self)
	assert.Equal(t, "Health", resp.Kind)
	assert.Equal(t, "ok", resp.Status)
}

// ---------------------------------------------------------------------------
// T093-2: GET /readyz — Database reachable
// ---------------------------------------------------------------------------

func TestHandler_Readyz_Ready(t *testing.T) {
	readinessMock := &mockReadinessHandler{
		result: &ReadinessResult{Ready: true, Database: "reachable"},
	}

	router := setupHealthRouter(readinessMock)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp readinessResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	assert.Equal(t, "/readyz", resp.Self)
	assert.Equal(t, "Readiness", resp.Kind)
	assert.Equal(t, "ready", resp.Status)
}

// ---------------------------------------------------------------------------
// T093-3: GET /readyz — Database unreachable
// ---------------------------------------------------------------------------

func TestHandler_Readyz_Unavailable(t *testing.T) {
	readinessMock := &mockReadinessHandler{
		result: &ReadinessResult{Ready: false, Database: "unreachable"},
	}

	router := setupHealthRouter(readinessMock)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var resp readinessResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	assert.Equal(t, "/readyz", resp.Self)
	assert.Equal(t, "Readiness", resp.Kind)
	assert.Equal(t, "unavailable", resp.Status)
	require.Contains(t, resp.Checks, "database")
	assert.Equal(t, "unreachable", resp.Checks["database"])
}

// ---------------------------------------------------------------------------
// T093-4: GET /readyz — Query bus returns an error
// ---------------------------------------------------------------------------

func TestHandler_Readyz_QueryError(t *testing.T) {
	readinessMock := &mockReadinessHandler{
		err: errors.New("unexpected failure"),
	}

	router := setupHealthRouter(readinessMock)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

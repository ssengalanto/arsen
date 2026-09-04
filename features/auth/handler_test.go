package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"arsen/pkg/cqrs"
	"arsen/pkg/response"
)

// ---------------------------------------------------------------------------
// Mock command handler for HTTP handler tests
// ---------------------------------------------------------------------------

type mockLoginHandler struct {
	result *LoginResult
	err    error
}

func (m *mockLoginHandler) Handle(_ context.Context, _ LoginCommand) (*LoginResult, error) {
	return m.result, m.err
}

// ---------------------------------------------------------------------------
// Helper: build a Handler wired to a mock, mounted on chi.
// ---------------------------------------------------------------------------

func setupAuthRouter(loginMock *mockLoginHandler) http.Handler {
	loginBus := cqrs.NewCommandBus[LoginCommand, *LoginResult](loginMock)
	r := chi.NewRouter()
	h := NewHandler(loginBus)
	h.RegisterRoutes(r)
	return r
}

// ---------------------------------------------------------------------------
// T064: POST /api/sessions — Login endpoint
// ---------------------------------------------------------------------------

func TestHandler_Login_Success(t *testing.T) {
	loginMock := &mockLoginHandler{
		result: &LoginResult{
			AccessToken:  "jwt-token",
			RefreshToken: "refresh-token",
			ExpiresIn:    900,
		},
	}

	router := setupAuthRouter(loginMock)

	body := `{"email":"alice@example.com","password":"Test123!!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp sessionResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	assert.Equal(t, "/api/sessions/current", resp.Self)
	assert.Equal(t, "Session", resp.Kind)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, "jwt-token", resp.AccessToken)
	assert.Equal(t, "refresh-token", resp.RefreshToken)
	assert.Equal(t, 900, resp.ExpiresIn)
}

func TestHandler_Login_Unauthorized(t *testing.T) {
	loginMock := &mockLoginHandler{
		err: cqrs.NewUnauthorizedError("invalid email or password"),
	}

	router := setupAuthRouter(loginMock)

	body := `{"email":"alice@example.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/problem+json")

	var prob response.ProblemDetail
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&prob))

	assert.Equal(t, http.StatusUnauthorized, prob.Status)
}

func TestHandler_Login_Forbidden(t *testing.T) {
	loginMock := &mockLoginHandler{
		err: cqrs.NewForbiddenError("you must verify your email address before logging in"),
	}

	router := setupAuthRouter(loginMock)

	body := `{"email":"alice@example.com","password":"Test123!!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/problem+json")

	var prob response.ProblemDetail
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&prob))

	assert.Equal(t, http.StatusForbidden, prob.Status)
}

func TestHandler_Login_InvalidJSON(t *testing.T) {
	router := setupAuthRouter(&mockLoginHandler{})

	body := `{this is not json}`
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var prob response.ProblemDetail
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&prob))

	assert.Equal(t, http.StatusBadRequest, prob.Status)
	assert.Equal(t, "Invalid request body.", prob.Detail)
}

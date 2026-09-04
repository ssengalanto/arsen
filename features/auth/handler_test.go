package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"arsen/pkg/cqrs"
	"arsen/pkg/jwt"
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

type mockRefreshHandler struct {
	result *RefreshTokenResult
	err    error
}

func (m *mockRefreshHandler) Handle(_ context.Context, _ RefreshTokenCommand) (*RefreshTokenResult, error) {
	return m.result, m.err
}

type mockLogoutHandler struct {
	err error
}

func (m *mockLogoutHandler) Handle(_ context.Context, _ LogoutCommand) (cqrs.Unit, error) {
	return cqrs.Unit{}, m.err
}

// ---------------------------------------------------------------------------
// Helper: build a Handler wired to mocks, mounted on chi.
// ---------------------------------------------------------------------------

// testJWTService returns a jwt.Service configured with deterministic test
// credentials.  Reused across all handler-level helpers.
func testJWTService() *jwt.Service {
	return jwt.NewService("test-secret-at-least-32-chars-long!!", "test-issuer", "test-audience")
}

func setupAuthRouter(loginMock *mockLoginHandler) http.Handler {
	return setupAuthRouterWithRefresh(loginMock, &mockRefreshHandler{})
}

func setupAuthRouterWithRefresh(loginMock *mockLoginHandler, refreshMock *mockRefreshHandler) http.Handler {
	return setupFullAuthRouter(loginMock, refreshMock, &mockLogoutHandler{})
}

func setupFullAuthRouter(loginMock *mockLoginHandler, refreshMock *mockRefreshHandler, logoutMock *mockLogoutHandler) http.Handler {
	loginBus := cqrs.NewCommandBus[LoginCommand, *LoginResult](loginMock)
	refreshBus := cqrs.NewCommandBus[RefreshTokenCommand, *RefreshTokenResult](refreshMock)
	logoutBus := cqrs.NewCommandBus[LogoutCommand, cqrs.Unit](logoutMock)
	jwtSvc := testJWTService()
	r := chi.NewRouter()
	h := NewHandler(loginBus, refreshBus, logoutBus)
	h.RegisterRoutes(r, jwtSvc)
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

// ---------------------------------------------------------------------------
// T077: POST /api/tokens — Refresh endpoint
// ---------------------------------------------------------------------------

func TestHandler_Refresh_Success(t *testing.T) {
	refreshMock := &mockRefreshHandler{
		result: &RefreshTokenResult{
			AccessToken:  "new-jwt-token",
			RefreshToken: "new-refresh-token",
			ExpiresIn:    900,
		},
	}

	router := setupAuthRouterWithRefresh(&mockLoginHandler{}, refreshMock)

	body := `{"refreshToken":"old-refresh-token"}`
	req := httptest.NewRequest(http.MethodPost, "/api/tokens", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp sessionResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	assert.Equal(t, "/api/tokens", resp.Self)
	assert.Equal(t, "TokenPair", resp.Kind)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, "new-jwt-token", resp.AccessToken)
	assert.Equal(t, "new-refresh-token", resp.RefreshToken)
	assert.Equal(t, 900, resp.ExpiresIn)
}

func TestHandler_Refresh_Unauthorized(t *testing.T) {
	refreshMock := &mockRefreshHandler{
		err: cqrs.NewUnauthorizedError("invalid or expired refresh token"),
	}

	router := setupAuthRouterWithRefresh(&mockLoginHandler{}, refreshMock)

	body := `{"refreshToken":"bad-token"}`
	req := httptest.NewRequest(http.MethodPost, "/api/tokens", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/problem+json")

	var prob response.ProblemDetail
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&prob))

	assert.Equal(t, http.StatusUnauthorized, prob.Status)
}

// ---------------------------------------------------------------------------
// DELETE /api/sessions/current — Logout endpoint
// ---------------------------------------------------------------------------

func TestHandler_Logout_Success(t *testing.T) {
	logoutMock := &mockLogoutHandler{}
	router := setupFullAuthRouter(&mockLoginHandler{}, &mockRefreshHandler{}, logoutMock)

	// Generate a valid access token to pass through the Auth middleware.
	jwtSvc := testJWTService()
	accessToken, err := jwtSvc.CreateToken("user-123", 15*time.Minute)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/api/sessions/current", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.String(), "expected empty response body for 204")
}

func TestHandler_Logout_Unauthorized_NoToken(t *testing.T) {
	router := setupFullAuthRouter(&mockLoginHandler{}, &mockRefreshHandler{}, &mockLogoutHandler{})

	req := httptest.NewRequest(http.MethodDelete, "/api/sessions/current", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/problem+json")

	var prob response.ProblemDetail
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&prob))

	assert.Equal(t, http.StatusUnauthorized, prob.Status)
}

func TestHandler_Logout_HandlerError(t *testing.T) {
	logoutMock := &mockLogoutHandler{
		err: errors.New("unexpected database failure"),
	}
	router := setupFullAuthRouter(&mockLoginHandler{}, &mockRefreshHandler{}, logoutMock)

	jwtSvc := testJWTService()
	accessToken, err := jwtSvc.CreateToken("user-123", 15*time.Minute)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/api/sessions/current", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// An unrecognised error falls through HandleError to a 500 Internal Server Error.
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/problem+json")

	var prob response.ProblemDetail
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&prob))

	assert.Equal(t, http.StatusInternalServerError, prob.Status)
}

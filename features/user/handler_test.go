package user

import (
	"context"
	"encoding/json"
	"fmt"
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
// Mock command handlers for HTTP handler tests
// ---------------------------------------------------------------------------

type mockRegisterHandler struct {
	result *RegisterResult
	err    error
}

func (m *mockRegisterHandler) Handle(_ context.Context, _ RegisterCommand) (*RegisterResult, error) {
	return m.result, m.err
}

type mockVerifyEmailHandler struct {
	result *VerifyEmailResult
	err    error
}

func (m *mockVerifyEmailHandler) Handle(_ context.Context, _ VerifyEmailCommand) (*VerifyEmailResult, error) {
	return m.result, m.err
}

type mockResendVerificationHandler struct {
	err error
}

func (m *mockResendVerificationHandler) Handle(_ context.Context, _ ResendVerificationCommand) (cqrs.Unit, error) {
	return cqrs.Unit{}, m.err
}

type mockForgotPasswordHandler struct {
	err error
}

func (m *mockForgotPasswordHandler) Handle(_ context.Context, _ ForgotPasswordCommand) (cqrs.Unit, error) {
	return cqrs.Unit{}, m.err
}

type mockResetPasswordHandler struct {
	err error
}

func (m *mockResetPasswordHandler) Handle(_ context.Context, _ ResetPasswordCommand) (cqrs.Unit, error) {
	return cqrs.Unit{}, m.err
}

// ---------------------------------------------------------------------------
// Helper: build a Handler wired to mock command handlers, mounted on chi.
// ---------------------------------------------------------------------------

func setupRouter(
	regHandler *mockRegisterHandler,
	verifyHandler *mockVerifyEmailHandler,
	resendHandler *mockResendVerificationHandler,
) http.Handler {
	registerBus := cqrs.NewCommandBus[RegisterCommand, *RegisterResult](regHandler)
	verifyBus := cqrs.NewCommandBus[VerifyEmailCommand, *VerifyEmailResult](verifyHandler)
	resendBus := cqrs.NewCommandBus[ResendVerificationCommand, cqrs.Unit](resendHandler)
	// Use no-op handlers for buses not exercised by existing tests.
	profileBus := cqrs.NewQueryBus[GetProfileQuery, *GetProfileResult](&mockGetProfileHandler{})
	forgotBus := cqrs.NewCommandBus[ForgotPasswordCommand, cqrs.Unit](&mockForgotPasswordHandler{})
	resetBus := cqrs.NewCommandBus[ResetPasswordCommand, cqrs.Unit](&mockResetPasswordHandler{})

	r := chi.NewRouter()
	h := NewHandler(registerBus, verifyBus, resendBus, profileBus, forgotBus, resetBus)
	// Pass a JWT service for route registration (existing tests don't hit authenticated routes).
	jwtSvc := jwt.NewService("test-secret-32-chars-long-enough", "test", "test")
	h.RegisterRoutes(r, jwtSvc)
	return r
}

// ---------------------------------------------------------------------------
// T053: POST /api/users — Register endpoint
// ---------------------------------------------------------------------------

func TestHandler_Register_Success(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	regMock := &mockRegisterHandler{
		result: &RegisterResult{
			User: &User{
				ID:        "test-id",
				Email:     "alice@example.com",
				CreatedAt: now,
			},
		},
	}

	router := setupRouter(regMock, &mockVerifyEmailHandler{}, &mockResendVerificationHandler{})

	body := `{"email":"alice@example.com","password":"Test123!!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "/api/users/test-id", rec.Header().Get("Location"))

	var resp userResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	assert.Equal(t, "/api/users/test-id", resp.Self)
	assert.Equal(t, "User", resp.Kind)
	assert.Equal(t, "test-id", resp.ID)
	assert.Equal(t, "alice@example.com", resp.Email)
	assert.False(t, resp.EmailVerified)
	assert.Equal(t, now.Format(time.RFC3339), resp.CreatedAt)
	assert.Contains(t, resp.Message, "verify")
}

func TestHandler_Register_ValidationError(t *testing.T) {
	regMock := &mockRegisterHandler{
		err: cqrs.NewValidationError([]cqrs.FieldError{
			{Field: "email", Detail: "is required"},
			{Field: "password", Detail: "is required"},
		}),
	}

	router := setupRouter(regMock, &mockVerifyEmailHandler{}, &mockResendVerificationHandler{})

	body := `{"email":"","password":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/problem+json")

	var prob response.ProblemDetail
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&prob))

	assert.Equal(t, http.StatusBadRequest, prob.Status)
	assert.NotEmpty(t, prob.Errors, "expected validation field errors in response")
	assert.Len(t, prob.Errors, 2)
}

// ---------------------------------------------------------------------------
// T105: JSON schema validation — Register response
// ---------------------------------------------------------------------------

func TestHandler_Register_ResponseSchema(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	regMock := &mockRegisterHandler{
		result: &RegisterResult{
			User: &User{ID: "test-id", Email: "a@b.com", CreatedAt: now},
		},
	}
	router := setupRouter(regMock, &mockVerifyEmailHandler{}, &mockResendVerificationHandler{})

	body := `{"email":"a@b.com","password":"Test123!!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var raw map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&raw))

	assert.Contains(t, raw, "self")
	assert.Contains(t, raw, "kind")
	assert.Equal(t, "User", raw["kind"])
	// Verify camelCase
	for key := range raw {
		assert.NotContains(t, key, "_", "response key %q should be camelCase", key)
	}
	// Verify createdAt is ISO 8601
	createdAt, ok := raw["createdAt"].(string)
	require.True(t, ok)
	_, err := time.Parse(time.RFC3339, createdAt)
	assert.NoError(t, err, "createdAt should be ISO 8601")
}

func TestHandler_Register_ConflictNormalizedTo400(t *testing.T) {
	regMock := &mockRegisterHandler{
		err: cqrs.NewConflictError("User", "alice@example.com", "email already registered"),
	}

	router := setupRouter(regMock, &mockVerifyEmailHandler{}, &mockResendVerificationHandler{})

	body := `{"email":"alice@example.com","password":"Test123!!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Anti-enumeration: conflict is normalized to 400, NOT 409.
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var prob response.ProblemDetail
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&prob))

	assert.Equal(t, "Registration failed.", prob.Detail)
}

func TestHandler_Register_InvalidJSON(t *testing.T) {
	router := setupRouter(&mockRegisterHandler{}, &mockVerifyEmailHandler{}, &mockResendVerificationHandler{})

	body := `{this is not json}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
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
// T054: POST /api/users/verify — Verify email endpoint
// ---------------------------------------------------------------------------

func TestHandler_VerifyEmail_Success(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	verifyMock := &mockVerifyEmailHandler{
		result: &VerifyEmailResult{
			User: &User{
				ID:            "verified-id",
				Email:         "alice@example.com",
				EmailVerified: true,
				CreatedAt:     now,
			},
		},
	}

	router := setupRouter(&mockRegisterHandler{}, verifyMock, &mockResendVerificationHandler{})

	body := `{"token":"some-valid-token"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users/verify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp userResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	assert.Equal(t, "/api/users/verified-id", resp.Self)
	assert.Equal(t, "User", resp.Kind)
	assert.True(t, resp.EmailVerified)
	assert.Equal(t, "Email verified successfully.", resp.Message)
}

func TestHandler_VerifyEmail_Unauthorized(t *testing.T) {
	verifyMock := &mockVerifyEmailHandler{
		err: cqrs.NewUnauthorizedError("invalid or expired verification token"),
	}

	router := setupRouter(&mockRegisterHandler{}, verifyMock, &mockResendVerificationHandler{})

	body := `{"token":"bad-token"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users/verify", strings.NewReader(body))
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
// T054: POST /api/users/resend-verification — Resend verification endpoint
// ---------------------------------------------------------------------------

func TestHandler_ResendVerification_Success(t *testing.T) {
	router := setupRouter(
		&mockRegisterHandler{},
		&mockVerifyEmailHandler{},
		&mockResendVerificationHandler{},
	)

	body := `{"email":"alice@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users/resend-verification", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp acknowledgmentResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	assert.Equal(t, "Acknowledgment", resp.Kind)
	assert.Contains(t, resp.Message, "verification email has been sent")
}

func TestHandler_ResendVerification_ErrorStillReturns200(t *testing.T) {
	resendMock := &mockResendVerificationHandler{
		err: fmt.Errorf("some internal error"),
	}

	router := setupRouter(
		&mockRegisterHandler{},
		&mockVerifyEmailHandler{},
		resendMock,
	)

	body := `{"email":"unknown@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users/resend-verification", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Anti-enumeration: always returns 200 even when an error occurs.
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp acknowledgmentResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	assert.Equal(t, "Acknowledgment", resp.Kind)
}

// ---------------------------------------------------------------------------
// Mock query handler for GET /api/users/me
// ---------------------------------------------------------------------------

type mockGetProfileHandler struct {
	result *GetProfileResult
	err    error
}

func (m *mockGetProfileHandler) Handle(_ context.Context, _ GetProfileQuery) (*GetProfileResult, error) {
	return m.result, m.err
}

// ---------------------------------------------------------------------------
// Helper: build a Handler with profile query bus, mounted on chi with JWT auth.
// ---------------------------------------------------------------------------

func setupRouterWithAuth(
	regHandler *mockRegisterHandler,
	verifyHandler *mockVerifyEmailHandler,
	resendHandler *mockResendVerificationHandler,
	profileHandler *mockGetProfileHandler,
	jwtSvc *jwt.Service,
) http.Handler {
	registerBus := cqrs.NewCommandBus[RegisterCommand, *RegisterResult](regHandler)
	verifyBus := cqrs.NewCommandBus[VerifyEmailCommand, *VerifyEmailResult](verifyHandler)
	resendBus := cqrs.NewCommandBus[ResendVerificationCommand, cqrs.Unit](resendHandler)
	profileBus := cqrs.NewQueryBus[GetProfileQuery, *GetProfileResult](profileHandler)
	forgotBus := cqrs.NewCommandBus[ForgotPasswordCommand, cqrs.Unit](&mockForgotPasswordHandler{})
	resetBus := cqrs.NewCommandBus[ResetPasswordCommand, cqrs.Unit](&mockResetPasswordHandler{})

	r := chi.NewRouter()
	h := NewHandler(registerBus, verifyBus, resendBus, profileBus, forgotBus, resetBus)
	h.RegisterRoutes(r, jwtSvc)
	return r
}

// setupRouterWithPasswordReset creates a router with custom forgot/reset password mocks
// and default no-op mocks for the other handlers.
func setupRouterWithPasswordReset(
	forgotMock *mockForgotPasswordHandler,
	resetMock *mockResetPasswordHandler,
) http.Handler {
	registerBus := cqrs.NewCommandBus[RegisterCommand, *RegisterResult](&mockRegisterHandler{})
	verifyBus := cqrs.NewCommandBus[VerifyEmailCommand, *VerifyEmailResult](&mockVerifyEmailHandler{})
	resendBus := cqrs.NewCommandBus[ResendVerificationCommand, cqrs.Unit](&mockResendVerificationHandler{})
	profileBus := cqrs.NewQueryBus[GetProfileQuery, *GetProfileResult](&mockGetProfileHandler{})
	forgotBus := cqrs.NewCommandBus[ForgotPasswordCommand, cqrs.Unit](forgotMock)
	resetBus := cqrs.NewCommandBus[ResetPasswordCommand, cqrs.Unit](resetMock)

	r := chi.NewRouter()
	jwtSvc := jwt.NewService("test-secret-32-chars-long-enough", "test", "test")
	h := NewHandler(registerBus, verifyBus, resendBus, profileBus, forgotBus, resetBus)
	h.RegisterRoutes(r, jwtSvc)
	return r
}

// ---------------------------------------------------------------------------
// T073: GET /api/users/me — Get profile endpoint
// ---------------------------------------------------------------------------

func TestHandler_GetProfile_Success(t *testing.T) {
	jwtSvc := jwt.NewService("test-secret-32-chars-long-enough", "test", "test")
	accessToken, err := jwtSvc.CreateToken("user-123", 15*time.Minute)
	require.NoError(t, err)

	now := time.Now().UTC().Truncate(time.Second)
	profileMock := &mockGetProfileHandler{
		result: &GetProfileResult{
			User: &User{
				ID:            "user-123",
				Email:         "alice@example.com",
				EmailVerified: true,
				CreatedAt:     now,
			},
		},
	}

	router := setupRouterWithAuth(
		&mockRegisterHandler{},
		&mockVerifyEmailHandler{},
		&mockResendVerificationHandler{},
		profileMock,
		jwtSvc,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/users/me", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp userResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	assert.Equal(t, "/api/users/me", resp.Self)
	assert.Equal(t, "User", resp.Kind)
	assert.Equal(t, "user-123", resp.ID)
	assert.Equal(t, "alice@example.com", resp.Email)
	assert.True(t, resp.EmailVerified)
	assert.Equal(t, now.Format(time.RFC3339), resp.CreatedAt)
}

func TestHandler_GetProfile_MissingToken(t *testing.T) {
	jwtSvc := jwt.NewService("test-secret-32-chars-long-enough", "test", "test")

	profileMock := &mockGetProfileHandler{}

	router := setupRouterWithAuth(
		&mockRegisterHandler{},
		&mockVerifyEmailHandler{},
		&mockResendVerificationHandler{},
		profileMock,
		jwtSvc,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/users/me", http.NoBody)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/problem+json")

	var prob response.ProblemDetail
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&prob))

	assert.Equal(t, http.StatusUnauthorized, prob.Status)
}

func TestHandler_GetProfile_ExpiredToken(t *testing.T) {
	jwtSvc := jwt.NewService("test-secret-32-chars-long-enough", "test", "test")
	// Create a token that expired 1 minute ago.
	expiredToken, err := jwtSvc.CreateToken("user-123", -1*time.Minute)
	require.NoError(t, err)

	profileMock := &mockGetProfileHandler{}

	router := setupRouterWithAuth(
		&mockRegisterHandler{},
		&mockVerifyEmailHandler{},
		&mockResendVerificationHandler{},
		profileMock,
		jwtSvc,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/users/me", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/problem+json")

	var prob response.ProblemDetail
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&prob))

	assert.Equal(t, http.StatusUnauthorized, prob.Status)
}

// ---------------------------------------------------------------------------
// T087: POST /api/users/forgot-password — Forgot password endpoint
// ---------------------------------------------------------------------------

func TestHandler_ForgotPassword_Success(t *testing.T) {
	forgotMock := &mockForgotPasswordHandler{}

	router := setupRouterWithPasswordReset(forgotMock, &mockResetPasswordHandler{})

	body := `{"email":"alice@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users/forgot-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp acknowledgmentResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	assert.Equal(t, "Acknowledgment", resp.Kind)
	assert.NotEmpty(t, resp.Message)
}

func TestHandler_ForgotPassword_AlwaysReturns200OnError(t *testing.T) {
	forgotMock := &mockForgotPasswordHandler{
		err: fmt.Errorf("some internal error"),
	}

	router := setupRouterWithPasswordReset(forgotMock, &mockResetPasswordHandler{})

	body := `{"email":"unknown@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users/forgot-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Anti-enumeration: always returns 200 even when an error occurs.
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp acknowledgmentResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	assert.Equal(t, "Acknowledgment", resp.Kind)
}

// ---------------------------------------------------------------------------
// T087: POST /api/users/reset-password — Reset password endpoint
// ---------------------------------------------------------------------------

func TestHandler_ResetPassword_Success(t *testing.T) {
	resetMock := &mockResetPasswordHandler{}

	router := setupRouterWithPasswordReset(&mockForgotPasswordHandler{}, resetMock)

	body := `{"token":"somevalidtoken","password":"N3w!S3cure"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users/reset-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp acknowledgmentResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	assert.Equal(t, "Acknowledgment", resp.Kind)
}

func TestHandler_ResetPassword_Unauthorized(t *testing.T) {
	resetMock := &mockResetPasswordHandler{
		err: cqrs.NewUnauthorizedError("invalid or expired reset token"),
	}

	router := setupRouterWithPasswordReset(&mockForgotPasswordHandler{}, resetMock)

	body := `{"token":"badtoken","password":"N3w!S3cure"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users/reset-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/problem+json")

	var prob response.ProblemDetail
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&prob))

	assert.Equal(t, http.StatusUnauthorized, prob.Status)
}

func TestHandler_ResetPassword_ValidationError(t *testing.T) {
	resetMock := &mockResetPasswordHandler{
		err: cqrs.NewValidationError([]cqrs.FieldError{
			{Field: "password", Detail: "must be at least 8 characters"},
			{Field: "password", Detail: "must contain at least 1 uppercase letter"},
		}),
	}

	router := setupRouterWithPasswordReset(&mockForgotPasswordHandler{}, resetMock)

	body := `{"token":"x","password":"weak"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users/reset-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/problem+json")

	var prob response.ProblemDetail
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&prob))

	assert.Equal(t, http.StatusBadRequest, prob.Status)
	assert.NotEmpty(t, prob.Errors, "expected validation field errors in response")
	assert.Len(t, prob.Errors, 2)
}

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

	r := chi.NewRouter()
	h := NewHandler(registerBus, verifyBus, resendBus)
	h.RegisterRoutes(r)
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

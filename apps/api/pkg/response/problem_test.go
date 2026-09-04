package response_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"arsen/pkg/cqrs"
	"arsen/pkg/response"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Helper tests ---

func TestBadRequest_SetsCorrectFields(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/users", http.NoBody)

	fieldErrors := []response.FieldError{
		{Field: "email", Detail: "is required"},
		{Field: "name", Detail: "must be at least 2 characters"},
	}

	response.BadRequest(w, r, "Request validation failed", fieldErrors)

	res := w.Result()
	defer res.Body.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/problem+json", res.Header.Get("Content-Type"))

	var body response.ProblemDetail
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

	assert.Equal(t, "about:blank", body.Type)
	assert.Equal(t, "Validation Failed", body.Title)
	assert.Equal(t, http.StatusBadRequest, body.Status)
	assert.Equal(t, "Request validation failed", body.Detail)
	assert.Equal(t, "/api/users", body.Instance)
	assert.Len(t, body.Errors, 2)
	assert.Equal(t, "email", body.Errors[0].Field)
	assert.Equal(t, "is required", body.Errors[0].Detail)
}

func TestUnauthorized_Sets401(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/me", http.NoBody)

	response.Unauthorized(w, r, "Invalid credentials")

	res := w.Result()
	defer res.Body.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)

	var body response.ProblemDetail
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

	assert.Equal(t, "Authentication Failed", body.Title)
	assert.Equal(t, http.StatusUnauthorized, body.Status)
}

func TestNotFound_Sets404(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/users/999", http.NoBody)

	response.NotFound(w, r, "User not found")

	res := w.Result()
	defer res.Body.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	var body response.ProblemDetail
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

	assert.Equal(t, "Not Found", body.Title)
	assert.Equal(t, http.StatusNotFound, body.Status)
}

func TestInternalError_WithCorrelationID(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/data", http.NoBody)

	response.InternalError(w, r, "abc-123-def")

	res := w.Result()
	defer res.Body.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	var body response.ProblemDetail
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

	assert.Contains(t, body.Detail, "abc-123-def")
	assert.Contains(t, body.Detail, "correlation_id")
}

func TestInternalError_WithoutCorrelationID(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/data", http.NoBody)

	response.InternalError(w, r, "")

	res := w.Result()
	defer res.Body.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	var body response.ProblemDetail
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

	assert.Equal(t, "an unexpected error occurred", body.Detail)
	assert.NotContains(t, body.Detail, "correlation_id")
}

// --- HandleError tests ---

func TestHandleError_NotFoundError_Returns404(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/users/123", http.NoBody)

	err := cqrs.NewNotFoundError("User", "123")
	response.HandleError(w, r, err)

	res := w.Result()
	defer res.Body.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	var body response.ProblemDetail
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

	assert.Equal(t, "Not Found", body.Title)
	assert.Contains(t, body.Detail, "User")
}

func TestHandleError_ConflictError_Returns409(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/users", http.NoBody)

	err := cqrs.NewConflictError("User", "alice@example.com", "email already exists")
	response.HandleError(w, r, err)

	res := w.Result()
	defer res.Body.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, http.StatusConflict, res.StatusCode)

	var body response.ProblemDetail
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

	assert.Equal(t, "Conflict", body.Title)
}

func TestHandleError_ValidationError_Returns400WithFieldErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/users", http.NoBody)

	err := cqrs.NewValidationError([]cqrs.FieldError{
		{Field: "email", Detail: "is required"},
		{Field: "password", Detail: "too short"},
	})
	response.HandleError(w, r, err)

	res := w.Result()
	defer res.Body.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	var body response.ProblemDetail
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

	assert.Equal(t, "Validation Failed", body.Title)
	assert.Len(t, body.Errors, 2)
	assert.Equal(t, "email", body.Errors[0].Field)
	assert.Equal(t, "is required", body.Errors[0].Detail)
	assert.Equal(t, "password", body.Errors[1].Field)
	assert.Equal(t, "too short", body.Errors[1].Detail)
}

func TestHandleError_UnauthorizedError_Returns401(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/me", http.NoBody)

	err := cqrs.NewUnauthorizedError("invalid token")
	response.HandleError(w, r, err)

	res := w.Result()
	defer res.Body.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)

	var body response.ProblemDetail
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

	assert.Equal(t, "Authentication Failed", body.Title)
}

func TestHandleError_ForbiddenError_Returns403(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/admin/users/1", http.NoBody)

	err := cqrs.NewForbiddenError("admin access required")
	response.HandleError(w, r, err)

	res := w.Result()
	defer res.Body.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, http.StatusForbidden, res.StatusCode)

	var body response.ProblemDetail
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

	assert.Equal(t, "Forbidden", body.Title)
}

func TestHandleError_UnknownError_Returns500(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/data", http.NoBody)

	err := errors.New("something completely unexpected")
	response.HandleError(w, r, err)

	res := w.Result()
	defer res.Body.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	var body response.ProblemDetail
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

	assert.Equal(t, "Internal Server Error", body.Title)
	assert.Equal(t, "an unexpected error occurred", body.Detail)
}

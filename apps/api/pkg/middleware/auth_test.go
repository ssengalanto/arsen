package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"arsen/pkg/jwt"
	"arsen/pkg/middleware"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestJWTService() *jwt.Service {
	return jwt.NewService("test-secret", "test-issuer", "test-audience")
}

func TestAuth_ValidToken_PassesAndSetsUserID(t *testing.T) {
	svc := newTestJWTService()
	token, err := svc.CreateToken("user-42", 5*time.Minute)
	require.NoError(t, err)

	var capturedUserID string
	var found bool

	handler := middleware.Auth(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID, found = middleware.GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/me", http.NoBody)
	r.Header.Set("Authorization", "Bearer "+token)

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, found)
	assert.Equal(t, "user-42", capturedUserID)
}

func TestAuth_MissingAuthorizationHeader_Returns401(t *testing.T) {
	svc := newTestJWTService()

	handler := middleware.Auth(svc)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/me", http.NoBody)
	// No Authorization header set.

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_MalformedAuthorizationHeader_Returns401(t *testing.T) {
	svc := newTestJWTService()

	handler := middleware.Auth(svc)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/me", http.NoBody)
	r.Header.Set("Authorization", "Basic abc123")

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_ExpiredToken_Returns401(t *testing.T) {
	svc := newTestJWTService()
	token, err := svc.CreateToken("user-42", -1*time.Minute)
	require.NoError(t, err)

	handler := middleware.Auth(svc)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/me", http.NoBody)
	r.Header.Set("Authorization", "Bearer "+token)

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_TamperedToken_Returns401(t *testing.T) {
	svc := newTestJWTService()
	token, err := svc.CreateToken("user-42", 5*time.Minute)
	require.NoError(t, err)

	// Tamper with the token by modifying a character in the signature (last part).
	tampered := token[:len(token)-1] + "X"
	if tampered == token {
		tampered = token[:len(token)-1] + "Y"
	}

	handler := middleware.Auth(svc)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/me", http.NoBody)
	r.Header.Set("Authorization", "Bearer "+tampered)

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_WrongIssuer_Returns401(t *testing.T) {
	// Create token with one issuer.
	wrongSvc := jwt.NewService("test-secret", "wrong-issuer", "test-audience")
	token, err := wrongSvc.CreateToken("user-42", 5*time.Minute)
	require.NoError(t, err)

	// Validate with a different issuer.
	correctSvc := jwt.NewService("test-secret", "correct-issuer", "test-audience")

	handler := middleware.Auth(correctSvc)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/me", http.NoBody)
	r.Header.Set("Authorization", "Bearer "+token)

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

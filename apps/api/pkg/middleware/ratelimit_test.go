package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"arsen/pkg/middleware"
	"arsen/pkg/response"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimit_AllowsRequestsWithinLimit(t *testing.T) {
	handler := middleware.RateLimit(10, 10)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/test", http.NoBody)
		r.RemoteAddr = "192.168.1.1:12345"

		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code, "request %d should be allowed", i+1)
	}
}

func TestRateLimit_Returns429WhenExceeded(t *testing.T) {
	handler := middleware.RateLimit(1, 1)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request should pass.
	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest(http.MethodGet, "/api/test", http.NoBody)
	r1.RemoteAddr = "10.0.0.1:54321"
	handler.ServeHTTP(w1, r1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second immediate request should be rate-limited.
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest(http.MethodGet, "/api/test", http.NoBody)
	r2.RemoteAddr = "10.0.0.1:54322"
	handler.ServeHTTP(w2, r2)

	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	assert.Equal(t, "application/problem+json", w2.Result().Header.Get("Content-Type"))

	var body response.ProblemDetail
	require.NoError(t, json.NewDecoder(w2.Body).Decode(&body))

	assert.Equal(t, http.StatusTooManyRequests, body.Status)
	assert.Equal(t, "Too Many Requests", body.Title)
}

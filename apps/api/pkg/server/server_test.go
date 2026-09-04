package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"arsen/pkg/config"
	"arsen/pkg/jwt"
	"arsen/pkg/response"
)

func testServer() *Server {
	cfg := &config.Config{
		Env:    "test",
		Server: config.ServerConfig{Address: ":0"},
	}
	jwtSvc := jwt.NewService("test-secret-32-chars-long-enough", "test", "test")
	srv := New(cfg, jwtSvc)
	// Register a test route so chi knows about the path.
	srv.Router.Post("/test-endpoint", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return srv
}

func TestServer_MethodNotAllowed_Returns405(t *testing.T) {
	srv := testServer()

	req := httptest.NewRequest(http.MethodGet, "/test-endpoint", http.NoBody)
	rec := httptest.NewRecorder()

	srv.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))

	var body response.ProblemDetail
	err := json.NewDecoder(rec.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, body.Status)
}

func TestServer_MethodNotAllowed_ProblemDetail(t *testing.T) {
	srv := testServer()

	req := httptest.NewRequest(http.MethodGet, "/test-endpoint", http.NoBody)
	rec := httptest.NewRecorder()

	srv.Router.ServeHTTP(rec, req)

	var body response.ProblemDetail
	err := json.NewDecoder(rec.Body).Decode(&body)
	require.NoError(t, err)

	assert.Equal(t, "about:blank", body.Type)
	assert.Equal(t, "Method Not Allowed", body.Title)
	assert.Equal(t, http.StatusMethodNotAllowed, body.Status)
	assert.Equal(t, "Method GET is not allowed for this resource.", body.Detail)
	assert.Equal(t, "/test-endpoint", body.Instance)
}

func TestServer_NotFound_Returns404(t *testing.T) {
	srv := testServer()

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", http.NoBody)
	rec := httptest.NewRecorder()

	srv.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))

	var body response.ProblemDetail
	err := json.NewDecoder(rec.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, body.Status)
}

func TestServer_NotFound_ProblemDetail(t *testing.T) {
	srv := testServer()

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", http.NoBody)
	rec := httptest.NewRecorder()

	srv.Router.ServeHTTP(rec, req)

	var body response.ProblemDetail
	err := json.NewDecoder(rec.Body).Decode(&body)
	require.NoError(t, err)

	assert.Equal(t, "about:blank", body.Type)
	assert.Equal(t, "Not Found", body.Title)
	assert.Equal(t, http.StatusNotFound, body.Status)
	assert.Equal(t, "The requested resource was not found.", body.Detail)
	assert.Equal(t, "/nonexistent", body.Instance)
}

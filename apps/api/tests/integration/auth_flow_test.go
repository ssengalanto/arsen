//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"arsen/features/auth"
	"arsen/features/health"
	"arsen/features/user"
	"arsen/pkg/cleanup"
	"arsen/pkg/config"
	"arsen/pkg/database"
	emailpkg "arsen/pkg/email"
	"arsen/pkg/jwt"
	"arsen/pkg/redis"
	"arsen/pkg/server"
)

// capturingSender captures all sent emails for test assertions.
type capturingSender struct {
	mu     sync.Mutex
	emails []emailpkg.SendParams
}

func (c *capturingSender) Send(_ context.Context, params emailpkg.SendParams) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.emails = append(c.emails, params)
	return nil
}

func (c *capturingSender) lastEmail() emailpkg.SendParams {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.emails[len(c.emails)-1]
}

// extractToken extracts a token from an email HTML body by finding the
// token= query parameter. Tokens are 64-character hex strings produced by
// crypto/rand → hex.EncodeToString (see pkg/token/token.go).
func extractToken(html string) string {
	re := regexp.MustCompile(`token=([0-9a-f]{64})`)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// readJSON decodes resp.Body into the target map and closes the body.
func readJSON(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &m))
	return m
}

func TestAuthFlow_E2E(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}
	if os.Getenv("REDIS_URL") == "" {
		t.Skip("REDIS_URL not set, skipping integration test")
	}

	// Set required environment variables for the config module.
	t.Setenv("DATABASE_URL", os.Getenv("DATABASE_URL"))
	t.Setenv("REDIS_URL", os.Getenv("REDIS_URL"))
	t.Setenv("JWT_SECRET", "integration-test-secret-32-chars!!")
	t.Setenv("JWT_ISSUER", "integration-test")
	t.Setenv("JWT_AUDIENCE", "integration-test")
	t.Setenv("APP_BASE_URL", "http://localhost:8080")
	t.Setenv("ENV", "dev")

	emailCapture := &capturingSender{}

	var srv *server.Server

	app := fxtest.New(t,
		config.Module,
		database.Module,
		redis.Module,
		jwt.Module,
		cleanup.Module,
		// Provide a capturing email sender instead of the real email module.
		// We skip email.Module entirely and supply our own Sender.
		fx.Provide(func() emailpkg.Sender {
			return emailCapture
		}),
		// Provide the server constructor without the lifecycle listener that
		// binds the real TCP port. We only need the Server struct so we can
		// wrap its Router with httptest.
		fx.Provide(server.New),
		// Feature modules register routes via fx.Invoke, which uses
		// *server.Server — that is satisfied by server.New above.
		user.Module,
		auth.Module,
		health.Module,
		fx.Populate(&srv),
	)
	app.RequireStart()
	defer app.RequireStop()

	// Use httptest to serve the chi router without binding a real port.
	ts := httptest.NewServer(srv.Router)
	defer ts.Close()

	client := ts.Client()

	// doJSON is a helper that sends a JSON request and returns the response.
	doJSON := func(method, path string, body interface{}, headers ...string) *http.Response {
		t.Helper()
		var bodyReader *bytes.Reader
		if body != nil {
			b, err := json.Marshal(body)
			require.NoError(t, err)
			bodyReader = bytes.NewReader(b)
		} else {
			bodyReader = bytes.NewReader(nil)
		}
		req, err := http.NewRequest(method, ts.URL+path, bodyReader)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		for i := 0; i < len(headers); i += 2 {
			req.Header.Set(headers[i], headers[i+1])
		}
		resp, err := client.Do(req)
		require.NoError(t, err)
		return resp
	}

	// ----------------------------------------------------------------
	// Step 1: Register a new user.
	// ----------------------------------------------------------------
	resp := doJSON("POST", "/api/users", map[string]string{
		"email":    "integration-test@example.com",
		"password": "Test123!!",
	})
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// ----------------------------------------------------------------
	// Step 2: Extract the verification token from the captured email.
	// ----------------------------------------------------------------
	verifyToken := extractToken(emailCapture.lastEmail().HTML)
	require.NotEmpty(t, verifyToken, "should have captured a verification token from the email HTML")

	// ----------------------------------------------------------------
	// Step 3: Verify the email address.
	// ----------------------------------------------------------------
	resp = doJSON("POST", "/api/users/verify", map[string]string{
		"token": verifyToken,
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := readJSON(t, resp)
	assert.Equal(t, true, body["emailVerified"])

	// ----------------------------------------------------------------
	// Step 4: Login with credentials — receive access & refresh tokens.
	// ----------------------------------------------------------------
	resp = doJSON("POST", "/api/sessions", map[string]string{
		"email":    "integration-test@example.com",
		"password": "Test123!!",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	loginBody := readJSON(t, resp)

	accessToken, ok := loginBody["accessToken"].(string)
	require.True(t, ok, "response should contain accessToken")
	require.NotEmpty(t, accessToken)

	refreshToken, ok := loginBody["refreshToken"].(string)
	require.True(t, ok, "response should contain refreshToken")
	require.NotEmpty(t, refreshToken)

	// ----------------------------------------------------------------
	// Step 5: Access the profile with the access token.
	// ----------------------------------------------------------------
	resp = doJSON("GET", "/api/users/me", nil,
		"Authorization", "Bearer "+accessToken,
	)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	profileBody := readJSON(t, resp)
	assert.Equal(t, "integration-test@example.com", profileBody["email"])
	assert.Equal(t, true, profileBody["emailVerified"])

	// ----------------------------------------------------------------
	// Step 6: Refresh the token pair.
	// ----------------------------------------------------------------
	resp = doJSON("POST", "/api/tokens", map[string]string{
		"refreshToken": refreshToken,
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	refreshBody := readJSON(t, resp)

	newAccessToken, ok := refreshBody["accessToken"].(string)
	require.True(t, ok, "refreshed response should contain accessToken")
	require.NotEmpty(t, newAccessToken)

	newRefreshToken, ok := refreshBody["refreshToken"].(string)
	require.True(t, ok, "refreshed response should contain refreshToken")
	require.NotEmpty(t, newRefreshToken)

	// ----------------------------------------------------------------
	// Step 7: Logout using the new access token.
	// ----------------------------------------------------------------
	resp = doJSON("DELETE", "/api/sessions/current", nil,
		"Authorization", "Bearer "+newAccessToken,
	)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// ----------------------------------------------------------------
	// Step 8: The old refresh token should now be rejected (family revoked).
	// ----------------------------------------------------------------
	resp = doJSON("POST", "/api/tokens", map[string]string{
		"refreshToken": refreshToken,
	})
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// The new refresh token should also be rejected (logout revoked all).
	resp = doJSON("POST", "/api/tokens", map[string]string{
		"refreshToken": newRefreshToken,
	})
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// ----------------------------------------------------------------
	// Step 9: Request a password reset (forgot-password).
	// ----------------------------------------------------------------
	resp = doJSON("POST", "/api/users/forgot-password", map[string]string{
		"email": "integration-test@example.com",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// ----------------------------------------------------------------
	// Step 10: Extract the reset token and set a new password.
	// ----------------------------------------------------------------
	resetToken := extractToken(emailCapture.lastEmail().HTML)
	require.NotEmpty(t, resetToken, "should have captured a password-reset token from the email HTML")

	resp = doJSON("POST", "/api/users/reset-password", map[string]string{
		"token":    resetToken,
		"password": "NewPass456!!",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// ----------------------------------------------------------------
	// Step 11: Login with the new password — should succeed.
	// ----------------------------------------------------------------
	resp = doJSON("POST", "/api/sessions", map[string]string{
		"email":    "integration-test@example.com",
		"password": "NewPass456!!",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// ----------------------------------------------------------------
	// Step 12: Login with the old password — should fail.
	// ----------------------------------------------------------------
	resp = doJSON("POST", "/api/sessions", map[string]string{
		"email":    "integration-test@example.com",
		"password": "Test123!!",
	})
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}

package config

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetViper(t *testing.T) {
	t.Helper()
	viper.Reset()
}

func TestNew_MissingJWTSecret(t *testing.T) {
	resetViper(t)
	t.Setenv("DATABASE_URL", "postgres://localhost/testdb")
	t.Setenv("REDIS_URL", "redis://localhost:6379")

	cfg, err := New()
	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET")
}

func TestNew_MissingDatabaseURL(t *testing.T) {
	resetViper(t)
	t.Setenv("JWT_SECRET", "supersecret")
	t.Setenv("REDIS_URL", "redis://localhost:6379")

	cfg, err := New()
	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL")
}

func TestNew_MissingRedisURL(t *testing.T) {
	resetViper(t)
	t.Setenv("JWT_SECRET", "supersecret")
	t.Setenv("DATABASE_URL", "postgres://localhost/testdb")

	cfg, err := New()
	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "REDIS_URL")
}

func TestNew_MissingResendAPIKeyFailsInProd(t *testing.T) {
	resetViper(t)
	t.Setenv("ENV", "prod")
	t.Setenv("JWT_SECRET", "supersecret")
	t.Setenv("DATABASE_URL", "postgres://localhost/testdb")
	t.Setenv("REDIS_URL", "redis://localhost:6379")

	cfg, err := New()
	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "RESEND_API_KEY")
}

func TestNew_MissingResendAPIKeySucceedsInDev(t *testing.T) {
	resetViper(t)
	t.Setenv("ENV", "dev")
	t.Setenv("JWT_SECRET", "supersecret")
	t.Setenv("DATABASE_URL", "postgres://localhost/testdb")
	t.Setenv("REDIS_URL", "redis://localhost:6379")

	cfg, err := New()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestNew_ValidConfigLoadsAllValues(t *testing.T) {
	resetViper(t)
	t.Setenv("ENV", "prod")
	t.Setenv("SERVER_ADDRESS", ":9090")
	t.Setenv("DATABASE_URL", "postgres://localhost/testdb")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("JWT_SECRET", "supersecret")
	t.Setenv("JWT_ISSUER", "myapp")
	t.Setenv("JWT_AUDIENCE", "myaudience")
	t.Setenv("ACCESS_TOKEN_DURATION", "30m")
	t.Setenv("REFRESH_TOKEN_DURATION", "72h")
	t.Setenv("RESEND_API_KEY", "re_123456")
	t.Setenv("EMAIL_FROM_ADDRESS", "no-reply@example.com")
	t.Setenv("EMAIL_FROM_NAME", "MyApp")
	t.Setenv("RATE_LIMIT_REQUESTS", "50")
	t.Setenv("RATE_LIMIT_BURST", "100")
	t.Setenv("CLEANUP_INTERVAL", "2h")
	t.Setenv("APP_BASE_URL", "https://example.com")

	cfg, err := New()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "prod", cfg.Env)
	assert.Equal(t, ":9090", cfg.Server.Address)
	assert.Equal(t, "postgres://localhost/testdb", cfg.Database.URL)
	assert.Equal(t, "redis://localhost:6379", cfg.Redis.URL)
	assert.Equal(t, "supersecret", cfg.JWT.Secret)
	assert.Equal(t, "myapp", cfg.JWT.Issuer)
	assert.Equal(t, "myaudience", cfg.JWT.Audience)
	assert.Equal(t, 30*time.Minute, cfg.JWT.AccessTokenDuration)
	assert.Equal(t, 72*time.Hour, cfg.JWT.RefreshTokenDuration)
	assert.Equal(t, "re_123456", cfg.Email.ResendAPIKey)
	assert.Equal(t, "no-reply@example.com", cfg.Email.FromAddress)
	assert.Equal(t, "MyApp", cfg.Email.FromName)
	assert.Equal(t, 50, cfg.RateLimit.Requests)
	assert.Equal(t, 100, cfg.RateLimit.Burst)
	assert.Equal(t, 2*time.Hour, cfg.Cleanup.Interval)
	assert.Equal(t, "https://example.com", cfg.AppBaseURL)
}

package auth

import (
	"bytes"
	"context"
	"crypto/subtle"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"arsen/pkg/config"
	"arsen/pkg/cqrs"
	"arsen/pkg/jwt"
	"arsen/pkg/token"
)

// ---------------------------------------------------------------------------
// refreshMockAuthRepo — a richer mock for refresh-token tests.
//
// Unlike the simpler mockAuthRepo in login_command_test.go, this mock supports
// hash-based lookup (GetRefreshTokenByHash) and family revocation tracking so
// we can assert rotation behaviour.
// ---------------------------------------------------------------------------

type refreshMockAuthRepo struct {
	tokens          []*RefreshToken
	createErr       error
	getByHashErr    error
	revokeFamilyErr error
}

func newRefreshMockAuthRepo() *refreshMockAuthRepo {
	return &refreshMockAuthRepo{}
}

func (m *refreshMockAuthRepo) CreateRefreshToken(_ context.Context, t *RefreshToken) error {
	if m.createErr != nil {
		return m.createErr
	}
	t.ID = "rt-mock-id"
	t.CreatedAt = time.Now()
	m.tokens = append(m.tokens, t)
	return nil
}

func (m *refreshMockAuthRepo) GetRefreshTokenByHash(_ context.Context, hash []byte) (*RefreshToken, error) {
	if m.getByHashErr != nil {
		return nil, m.getByHashErr
	}
	for _, rt := range m.tokens {
		if subtle.ConstantTimeCompare(rt.TokenHash, hash) == 1 && !rt.Revoked {
			return rt, nil
		}
	}
	return nil, cqrs.NewNotFoundError("RefreshToken", "hash")
}

func (m *refreshMockAuthRepo) RevokeRefreshTokenFamily(_ context.Context, familyID string) error {
	if m.revokeFamilyErr != nil {
		return m.revokeFamilyErr
	}
	for _, rt := range m.tokens {
		if rt.FamilyID == familyID {
			rt.Revoked = true
		}
	}
	return nil
}

func (m *refreshMockAuthRepo) RevokeAllUserRefreshTokens(_ context.Context, userID string) error {
	for _, rt := range m.tokens {
		if rt.UserID == userID {
			rt.Revoked = true
		}
	}
	return nil
}

// revokedTokens returns all tokens in the mock that have been marked revoked.
func (m *refreshMockAuthRepo) revokedTokens() []*RefreshToken {
	var out []*RefreshToken
	for _, rt := range m.tokens {
		if rt.Revoked {
			out = append(out, rt)
		}
	}
	return out
}

// tokensByFamily returns all tokens sharing the given family ID.
func (m *refreshMockAuthRepo) tokensByFamily(familyID string) []*RefreshToken {
	var out []*RefreshToken
	for _, rt := range m.tokens {
		if rt.FamilyID == familyID {
			out = append(out, rt)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Helper: seed a valid refresh token into the mock and return the raw string.
// ---------------------------------------------------------------------------

func seedRefreshToken(t *testing.T, repo *refreshMockAuthRepo, userID, familyID string, expiresAt time.Time) string { //nolint:unparam // test helper parameterized for clarity
	t.Helper()
	raw, hash, err := token.Generate()
	require.NoError(t, err)

	rt := &RefreshToken{
		ID:        "seed-rt-id",
		UserID:    userID,
		TokenHash: hash,
		FamilyID:  familyID,
		ExpiresAt: expiresAt,
		Revoked:   false,
		CreatedAt: time.Now(),
	}
	repo.tokens = append(repo.tokens, rt)
	return raw
}

// ---------------------------------------------------------------------------
// Test setup helper
// ---------------------------------------------------------------------------

func setupRefreshTest(t *testing.T) (*RefreshTokenCommandHandler, *refreshMockAuthRepo) {
	t.Helper()
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:               "test-secret-at-least-32-chars-long!!",
			Issuer:               "test-issuer",
			Audience:             "test-audience",
			AccessTokenDuration:  15 * time.Minute,
			RefreshTokenDuration: 7 * 24 * time.Hour,
		},
	}
	jwtSvc := jwt.NewService(cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.Audience)
	repo := newRefreshMockAuthRepo()
	handler := NewRefreshTokenCommandHandler(repo, jwtSvc, cfg)
	return handler, repo
}

// ---------------------------------------------------------------------------
// T077: RefreshTokenCommandHandler tests
// ---------------------------------------------------------------------------

func TestRefreshTokenCommandHandler_ValidTokenRotates(t *testing.T) {
	handler, repo := setupRefreshTest(t)

	rawToken := seedRefreshToken(t, repo, "user-123", "family-abc", time.Now().Add(24*time.Hour))

	result, err := handler.Handle(context.Background(), RefreshTokenCommand{Token: rawToken})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.NotEmpty(t, result.AccessToken, "expected non-empty access token")
	assert.NotEmpty(t, result.RefreshToken, "expected non-empty refresh token")
	assert.NotEqual(t, rawToken, result.RefreshToken, "rotated token must differ from original")
}

func TestRefreshTokenCommandHandler_ExpiredTokenRejected(t *testing.T) {
	handler, repo := setupRefreshTest(t)

	rawToken := seedRefreshToken(t, repo, "user-123", "family-abc", time.Now().Add(-1*time.Hour))

	result, err := handler.Handle(context.Background(), RefreshTokenCommand{Token: rawToken})
	assert.Nil(t, result)
	require.Error(t, err)

	var unauthorizedErr *cqrs.UnauthorizedError
	require.True(t, errors.As(err, &unauthorizedErr), "expected UnauthorizedError, got %T: %v", err, err)
}

func TestRefreshTokenCommandHandler_InvalidUnknownTokenRejected(t *testing.T) {
	handler, _ := setupRefreshTest(t)

	// Generate a token that is NOT stored in the repo.
	raw, _, err := token.Generate()
	require.NoError(t, err)

	result, err := handler.Handle(context.Background(), RefreshTokenCommand{Token: raw})
	assert.Nil(t, result)
	require.Error(t, err)

	var unauthorizedErr *cqrs.UnauthorizedError
	require.True(t, errors.As(err, &unauthorizedErr), "expected UnauthorizedError, got %T: %v", err, err)
}

func TestRefreshTokenCommandHandler_OldTokenRevokedAfterRotation(t *testing.T) {
	handler, repo := setupRefreshTest(t)

	rawToken := seedRefreshToken(t, repo, "user-123", "family-abc", time.Now().Add(24*time.Hour))

	// Capture the original token hash for comparison.
	origHash, err := token.Hash(rawToken)
	require.NoError(t, err)

	_, err = handler.Handle(context.Background(), RefreshTokenCommand{Token: rawToken})
	require.NoError(t, err)

	// Verify the original token was revoked.
	revoked := repo.revokedTokens()
	require.NotEmpty(t, revoked, "expected at least one revoked token")

	foundOriginal := false
	for _, rt := range revoked {
		if bytes.Equal(rt.TokenHash, origHash) {
			foundOriginal = true
			break
		}
	}
	assert.True(t, foundOriginal, "original token should be revoked after rotation")
}

func TestRefreshTokenCommandHandler_NewTokenHasSameFamilyID(t *testing.T) {
	handler, repo := setupRefreshTest(t)

	const familyID = "family-abc"
	rawToken := seedRefreshToken(t, repo, "user-123", familyID, time.Now().Add(24*time.Hour))

	result, err := handler.Handle(context.Background(), RefreshTokenCommand{Token: rawToken})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Find the newly created (non-revoked) token in the family.
	familyTokens := repo.tokensByFamily(familyID)
	require.GreaterOrEqual(t, len(familyTokens), 2, "expected at least 2 tokens in family (old + new)")

	// The new token is the one that is not revoked and whose hash matches the
	// rotated raw token.
	newHash, err := token.Hash(result.RefreshToken)
	require.NoError(t, err)

	var newToken *RefreshToken
	for _, rt := range familyTokens {
		if bytes.Equal(rt.TokenHash, newHash) {
			newToken = rt
			break
		}
	}
	require.NotNil(t, newToken, "expected to find the newly created token in the repo")
	assert.Equal(t, familyID, newToken.FamilyID, "new token must share the same family ID")
}

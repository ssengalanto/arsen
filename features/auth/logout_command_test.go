package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// logoutMockRepo — a lightweight mock for logout-specific tests.
//
// Only RevokeAllUserRefreshTokens is meaningful here; the other methods are
// no-op stubs satisfying the Repository interface.
// ---------------------------------------------------------------------------

type logoutMockRepo struct {
	revokedUserIDs []string
	revokeErr      error
}

func (m *logoutMockRepo) CreateRefreshToken(_ context.Context, _ *RefreshToken) error {
	return nil
}

func (m *logoutMockRepo) GetRefreshTokenByHash(_ context.Context, _ []byte) (*RefreshToken, error) {
	return nil, nil
}

func (m *logoutMockRepo) RevokeRefreshTokenFamily(_ context.Context, _ string) error {
	return nil
}

func (m *logoutMockRepo) RevokeAllUserRefreshTokens(_ context.Context, userID string) error {
	if m.revokeErr != nil {
		return m.revokeErr
	}
	m.revokedUserIDs = append(m.revokedUserIDs, userID)
	return nil
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestLogoutCommandHandler_RevokesAllUserTokens(t *testing.T) {
	repo := &logoutMockRepo{}
	handler := NewLogoutCommandHandler(repo)

	const userID = "user-123"

	result, err := handler.Handle(context.Background(), LogoutCommand{UserID: userID})
	require.NoError(t, err)
	assert.Equal(t, result, struct{}{})

	require.Len(t, repo.revokedUserIDs, 1, "expected RevokeAllUserRefreshTokens to be called once")
	assert.Equal(t, userID, repo.revokedUserIDs[0])
}

func TestLogoutCommandHandler_IdempotentOnNoTokens(t *testing.T) {
	repo := &logoutMockRepo{}
	handler := NewLogoutCommandHandler(repo)

	// Call Handle for a user that has no tokens — should succeed without error.
	result, err := handler.Handle(context.Background(), LogoutCommand{UserID: "user-no-tokens"})
	require.NoError(t, err)
	assert.Equal(t, result, struct{}{})

	require.Len(t, repo.revokedUserIDs, 1)
	assert.Equal(t, "user-no-tokens", repo.revokedUserIDs[0])
}

func TestLogoutCommandHandler_RepoErrorPropagates(t *testing.T) {
	repoErr := errors.New("database connection lost")
	repo := &logoutMockRepo{revokeErr: repoErr}
	handler := NewLogoutCommandHandler(repo)

	_, err := handler.Handle(context.Background(), LogoutCommand{UserID: "user-123"})
	require.Error(t, err)
	assert.ErrorIs(t, err, repoErr, "expected the underlying repo error to be wrapped")
	assert.Contains(t, err.Error(), "logout: failed to revoke refresh tokens")
}

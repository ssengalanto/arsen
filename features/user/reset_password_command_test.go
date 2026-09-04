package user

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"arsen/pkg/cqrs"
	"arsen/pkg/token"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// mockRefreshTokenRevoker implements RefreshTokenRevoker for testing.
type mockRefreshTokenRevoker struct {
	revokedUserIDs []string
	revokeErr      error
}

func (m *mockRefreshTokenRevoker) RevokeAllUserRefreshTokens(_ context.Context, userID string) error {
	if m.revokeErr != nil {
		return m.revokeErr
	}
	m.revokedUserIDs = append(m.revokedUserIDs, userID)
	return nil
}

func setupResetPasswordTest(t *testing.T) (*ResetPasswordCommandHandler, *mockRepository, *mockRefreshTokenRevoker, string) {
	t.Helper()
	repo := newMockRepository()
	revoker := &mockRefreshTokenRevoker{}
	handler := NewResetPasswordCommandHandler(repo, revoker)

	// Seed a user.
	repo.users["user-1"] = &User{
		ID:            "user-1",
		Email:         "alice@example.com",
		PasswordHash:  "$2a$12$LJ3m4ys3Lf0JoRHlVh7mle0m/Fwv2kSOwW9wOoCeeVqFkYl6n.rTa",
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	repo.usersByEmail["alice@example.com"] = repo.users["user-1"]

	// Seed a valid reset token.
	rawToken, tokenHash, err := token.Generate()
	require.NoError(t, err)
	key := hex.EncodeToString(tokenHash)
	repo.passwordResetTokens[key] = &PasswordResetToken{
		ID:        "prt-1",
		UserID:    "user-1",
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	return handler, repo, revoker, rawToken
}

func TestResetPasswordCommandHandler_ValidTokenResetsPassword(t *testing.T) {
	handler, repo, revoker, rawToken := setupResetPasswordTest(t)

	newPassword := "N3w!S3cure"
	result, err := handler.Handle(context.Background(), ResetPasswordCommand{
		Token:    rawToken,
		Password: newPassword,
	})
	require.NoError(t, err)
	assert.Equal(t, struct{}{}, result)

	// Assert user's password was changed and bcrypt verifies with the new password.
	user := repo.users["user-1"]
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(newPassword))
	assert.NoError(t, err, "bcrypt comparison should succeed with the new password")

	// Assert reset tokens were invalidated.
	assert.Contains(t, repo.invalidatedResetUserIDs, "user-1",
		"InvalidateUserPasswordResetTokens should have been called")

	// Assert refresh tokens were revoked.
	assert.Contains(t, revoker.revokedUserIDs, "user-1",
		"RevokeAllUserRefreshTokens should have been called")
}

func TestResetPasswordCommandHandler_ExpiredTokenRejected(t *testing.T) {
	repo := newMockRepository()
	revoker := &mockRefreshTokenRevoker{}
	handler := NewResetPasswordCommandHandler(repo, revoker)

	// Seed a user.
	repo.users["user-1"] = &User{
		ID:            "user-1",
		Email:         "alice@example.com",
		PasswordHash:  "$2a$12$LJ3m4ys3Lf0JoRHlVh7mle0m/Fwv2kSOwW9wOoCeeVqFkYl6n.rTa",
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	repo.usersByEmail["alice@example.com"] = repo.users["user-1"]

	// Seed an expired reset token.
	rawToken, tokenHash, err := token.Generate()
	require.NoError(t, err)
	key := hex.EncodeToString(tokenHash)
	repo.passwordResetTokens[key] = &PasswordResetToken{
		ID:        "prt-expired",
		UserID:    "user-1",
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // expired 1 hour ago
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	_, err = handler.Handle(context.Background(), ResetPasswordCommand{
		Token:    rawToken,
		Password: "N3w!S3cure",
	})
	require.Error(t, err)

	var unauthorizedErr *cqrs.UnauthorizedError
	assert.True(t, errors.As(err, &unauthorizedErr), "expected UnauthorizedError, got %T: %v", err, err)
}

func TestResetPasswordCommandHandler_UsedTokenRejected(t *testing.T) {
	repo := newMockRepository()
	revoker := &mockRefreshTokenRevoker{}
	handler := NewResetPasswordCommandHandler(repo, revoker)

	// Seed a user.
	repo.users["user-1"] = &User{
		ID:            "user-1",
		Email:         "alice@example.com",
		PasswordHash:  "$2a$12$LJ3m4ys3Lf0JoRHlVh7mle0m/Fwv2kSOwW9wOoCeeVqFkYl6n.rTa",
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	repo.usersByEmail["alice@example.com"] = repo.users["user-1"]

	// Seed a used reset token (UsedAt is set).
	rawToken, tokenHash, err := token.Generate()
	require.NoError(t, err)
	now := time.Now()
	key := hex.EncodeToString(tokenHash)
	repo.passwordResetTokens[key] = &PasswordResetToken{
		ID:        "prt-used",
		UserID:    "user-1",
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UsedAt:    &now, // already used
		CreatedAt: time.Now(),
	}

	_, err = handler.Handle(context.Background(), ResetPasswordCommand{
		Token:    rawToken,
		Password: "N3w!S3cure",
	})
	require.Error(t, err)

	var unauthorizedErr *cqrs.UnauthorizedError
	assert.True(t, errors.As(err, &unauthorizedErr), "expected UnauthorizedError, got %T: %v", err, err)
}

func TestResetPasswordCommandHandler_InvalidTokenRejected(t *testing.T) {
	repo := newMockRepository()
	revoker := &mockRefreshTokenRevoker{}
	handler := NewResetPasswordCommandHandler(repo, revoker)

	// Call with a completely unknown token (no seeded tokens at all).
	_, err := handler.Handle(context.Background(), ResetPasswordCommand{
		Token:    "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		Password: "N3w!S3cure",
	})
	require.Error(t, err)

	var unauthorizedErr *cqrs.UnauthorizedError
	assert.True(t, errors.As(err, &unauthorizedErr), "expected UnauthorizedError, got %T: %v", err, err)
}

func TestResetPasswordCommand_WeakPasswordValidation(t *testing.T) {
	cmd := ResetPasswordCommand{Token: "sometoken", Password: "abc"}
	err := cmd.Validate()
	require.Error(t, err)

	var valErr *cqrs.ValidationError
	require.True(t, errors.As(err, &valErr), "expected ValidationError, got %T: %v", err, err)

	// Check that password field errors are present.
	var passwordErrors []string
	for _, fe := range valErr.Fields {
		if fe.Field == "password" {
			passwordErrors = append(passwordErrors, fe.Detail)
		}
	}
	assert.NotEmpty(t, passwordErrors, "should have password validation errors")
}

func TestResetPasswordCommand_MissingTokenValidation(t *testing.T) {
	cmd := ResetPasswordCommand{Token: "", Password: "N3w!S3cure"}
	err := cmd.Validate()
	require.Error(t, err)

	var valErr *cqrs.ValidationError
	require.True(t, errors.As(err, &valErr), "expected ValidationError, got %T: %v", err, err)

	found := false
	for _, fe := range valErr.Fields {
		if fe.Field == "token" {
			found = true
			break
		}
	}
	assert.True(t, found, "should have token validation error")
}

func TestResetPasswordCommandHandler_RevokesAllRefreshTokens(t *testing.T) {
	handler, _, revoker, rawToken := setupResetPasswordTest(t)

	_, err := handler.Handle(context.Background(), ResetPasswordCommand{
		Token:    rawToken,
		Password: "N3w!S3cure",
	})
	require.NoError(t, err)

	assert.Contains(t, revoker.revokedUserIDs, "user-1",
		"RevokeAllUserRefreshTokens should have been called for the user")
}

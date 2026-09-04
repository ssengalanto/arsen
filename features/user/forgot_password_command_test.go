package user

import (
	"context"
	"testing"
	"time"

	"arsen/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupForgotPasswordTest(t *testing.T) (*ForgotPasswordCommandHandler, *mockRepository, *mockEmailSender) {
	t.Helper()
	repo := newMockRepository()
	emailSender := &mockEmailSender{}
	cfg := &config.Config{AppBaseURL: "http://localhost:8080"}
	handler := NewForgotPasswordCommandHandler(repo, emailSender, cfg)
	return handler, repo, emailSender
}

func seedVerifiedUser(repo *mockRepository) *User {
	u := &User{
		ID:            "user-1",
		Email:         "alice@example.com",
		PasswordHash:  "$2a$12$dummyhashfortest000000000000000000000000000000000000",
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	repo.users[u.ID] = u
	repo.usersByEmail[u.Email] = u
	return u
}

func TestForgotPasswordCommandHandler_SendsResetForVerifiedUser(t *testing.T) {
	handler, repo, emailSender := setupForgotPasswordTest(t)
	user := seedVerifiedUser(repo)

	result, err := handler.Handle(context.Background(), ForgotPasswordCommand{Email: user.Email})
	require.NoError(t, err)
	assert.Equal(t, struct{}{}, result)

	// Assert email was sent.
	require.Len(t, emailSender.sent, 1, "exactly one email should be sent")
	assert.Equal(t, "alice@example.com", emailSender.sent[0].To)
	assert.Equal(t, "Reset your password", emailSender.sent[0].Subject)
	assert.Contains(t, emailSender.sent[0].HTML, "http://localhost:8080")

	// Assert a password reset token was created in the repo.
	found := false
	for _, prt := range repo.passwordResetTokens {
		if prt.UserID == user.ID {
			found = true
			assert.NotEmpty(t, prt.TokenHash, "token hash should not be empty")
			assert.False(t, prt.ExpiresAt.IsZero(), "expires_at should be set")
			assert.WithinDuration(t, time.Now().Add(1*time.Hour), prt.ExpiresAt, 5*time.Second,
				"token should expire in approximately 1 hour")
			break
		}
	}
	assert.True(t, found, "password reset token should be created for the user")

	// Assert previous tokens were invalidated.
	assert.Contains(t, repo.invalidatedResetUserIDs, user.ID,
		"InvalidateUserPasswordResetTokens should have been called for the user")
}

func TestForgotPasswordCommandHandler_NoEmailForUnverifiedUser(t *testing.T) {
	handler, repo, emailSender := setupForgotPasswordTest(t)

	// Seed an unverified user.
	u := &User{
		ID:            "user-2",
		Email:         "unverified@example.com",
		PasswordHash:  "$2a$12$dummyhash",
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	repo.users[u.ID] = u
	repo.usersByEmail[u.Email] = u

	result, err := handler.Handle(context.Background(), ForgotPasswordCommand{Email: u.Email})
	require.NoError(t, err, "should return no error for unverified user (anti-enumeration)")
	assert.Equal(t, struct{}{}, result)

	// No email should be sent.
	assert.Empty(t, emailSender.sent, "no email should be sent for unverified user")

	// No reset token should be created.
	for _, prt := range repo.passwordResetTokens {
		assert.NotEqual(t, u.ID, prt.UserID, "no reset token should be created for unverified user")
	}
}

func TestForgotPasswordCommandHandler_NoEnumerationForNonExistent(t *testing.T) {
	handler, _, emailSender := setupForgotPasswordTest(t)

	result, err := handler.Handle(context.Background(), ForgotPasswordCommand{Email: "nobody@example.com"})
	require.NoError(t, err, "should return no error for non-existent email (anti-enumeration)")
	assert.Equal(t, struct{}{}, result)

	// No email should be sent.
	assert.Empty(t, emailSender.sent, "no email should be sent for non-existent user")
}

func TestForgotPasswordCommandHandler_InvalidatesPreviousTokens(t *testing.T) {
	handler, repo, _ := setupForgotPasswordTest(t)
	user := seedVerifiedUser(repo)

	// Call Handle twice.
	_, err := handler.Handle(context.Background(), ForgotPasswordCommand{Email: user.Email})
	require.NoError(t, err)

	_, err = handler.Handle(context.Background(), ForgotPasswordCommand{Email: user.Email})
	require.NoError(t, err)

	// The invalidatedResetUserIDs should contain the user ID (at least twice,
	// once per call, since each call invalidates previous tokens).
	count := 0
	for _, id := range repo.invalidatedResetUserIDs {
		if id == user.ID {
			count++
		}
	}
	assert.GreaterOrEqual(t, count, 2,
		"InvalidateUserPasswordResetTokens should be called on each Handle invocation")
}

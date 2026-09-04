package user

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"arsen/pkg/config"
	"arsen/pkg/token"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResendVerification_UnverifiedUser(t *testing.T) {
	repo := newMockRepository()
	sender := &mockEmailSender{}
	cfg := &config.Config{AppBaseURL: "http://localhost:8080"}

	// Pre-populate an unverified user.
	user := &User{
		ID:            "user-1",
		Email:         "resend@example.com",
		PasswordHash:  "hash",
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	repo.users[user.ID] = user
	repo.usersByEmail[user.Email] = user

	handler := NewResendVerificationCommandHandler(repo, sender, cfg)

	result, err := handler.Handle(context.Background(), ResendVerificationCommand{Email: "resend@example.com"})
	require.NoError(t, err)
	assert.Equal(t, struct{}{}, result)

	// Verify a new verification token was created.
	found := false
	for _, vt := range repo.verificationTokens {
		if vt.UserID == user.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "new verification token should be created")

	// Verify email was sent.
	require.Len(t, sender.sent, 1, "one verification email should be sent")
	assert.Equal(t, "resend@example.com", sender.sent[0].To)
	assert.Equal(t, "Verify your email", sender.sent[0].Subject)
	assert.Contains(t, sender.sent[0].HTML, "http://localhost:8080/api/users/verify?token=")
}

func TestResendVerification_InvalidatesOldTokens(t *testing.T) {
	repo := newMockRepository()
	sender := &mockEmailSender{}
	cfg := &config.Config{AppBaseURL: "http://localhost:8080"}

	// Pre-populate user.
	user := &User{
		ID:            "user-1",
		Email:         "resend@example.com",
		PasswordHash:  "hash",
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	repo.users[user.ID] = user
	repo.usersByEmail[user.Email] = user

	// Pre-populate an existing verification token.
	_, tokenHash, err := token.Generate()
	require.NoError(t, err)
	oldToken := &VerificationToken{
		ID:        "old-vt",
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
	key := hex.EncodeToString(tokenHash)
	repo.verificationTokens[key] = oldToken

	handler := NewResendVerificationCommandHandler(repo, sender, cfg)

	_, err = handler.Handle(context.Background(), ResendVerificationCommand{Email: "resend@example.com"})
	require.NoError(t, err)

	// Verify old token was invalidated.
	assert.Contains(t, repo.invalidatedUserIDs, user.ID, "InvalidateUserVerificationTokens should have been called for the user")
	assert.NotNil(t, oldToken.UsedAt, "old token should have UsedAt set after invalidation")

	// Verify a new token was also created (should be 2 total: old invalidated + new).
	newTokenCount := 0
	for _, vt := range repo.verificationTokens {
		if vt.UserID == user.ID {
			newTokenCount++
		}
	}
	assert.Equal(t, 2, newTokenCount, "should have old + new verification token")
}

func TestResendVerification_NonExistentEmail(t *testing.T) {
	repo := newMockRepository()
	sender := &mockEmailSender{}
	cfg := &config.Config{AppBaseURL: "http://localhost:8080"}
	handler := NewResendVerificationCommandHandler(repo, sender, cfg)

	result, err := handler.Handle(context.Background(), ResendVerificationCommand{Email: "unknown@example.com"})
	require.NoError(t, err, "should not return error for unknown email (no enumeration)")
	assert.Equal(t, struct{}{}, result)

	// Verify no email was sent.
	assert.Empty(t, sender.sent, "no email should be sent for non-existent user")
}

func TestResendVerification_AlreadyVerified(t *testing.T) {
	repo := newMockRepository()
	sender := &mockEmailSender{}
	cfg := &config.Config{AppBaseURL: "http://localhost:8080"}

	// Pre-populate a verified user.
	user := &User{
		ID:            "user-1",
		Email:         "verified@example.com",
		PasswordHash:  "hash",
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	repo.users[user.ID] = user
	repo.usersByEmail[user.Email] = user

	handler := NewResendVerificationCommandHandler(repo, sender, cfg)

	result, err := handler.Handle(context.Background(), ResendVerificationCommand{Email: "verified@example.com"})
	require.NoError(t, err)
	assert.Equal(t, struct{}{}, result)

	// Verify no email was sent and no new token created.
	assert.Empty(t, sender.sent, "no email should be sent for already verified user")
	assert.Empty(t, repo.verificationTokens, "no verification token should be created for already verified user")
}

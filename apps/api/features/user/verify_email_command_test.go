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
)

func newVerifySetup(t *testing.T) (repo *mockRepository, rawToken string, user *User) {
	t.Helper()

	repo = newMockRepository()

	// Create user.
	user = &User{
		ID:            "user-1",
		Email:         "verify@example.com",
		PasswordHash:  "hash",
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	repo.users[user.ID] = user
	repo.usersByEmail[user.Email] = user

	// Generate and store a valid verification token.
	raw, tokenHash, err := token.Generate()
	require.NoError(t, err)

	vt := &VerificationToken{
		ID:        "vt-1",
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
	key := hex.EncodeToString(tokenHash)
	repo.verificationTokens[key] = vt

	return repo, raw, user
}

func TestVerifyEmailCommand_ValidToken(t *testing.T) {
	repo, rawToken, user := newVerifySetup(t)
	eventBus := cqrs.NewEventBus[EmailVerifiedEvent]()
	handler := NewVerifyEmailCommandHandler(repo, eventBus)

	result, err := handler.Handle(context.Background(), VerifyEmailCommand{Token: rawToken})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.User)

	assert.True(t, result.User.EmailVerified, "user should be verified after valid token")
	assert.Equal(t, user.ID, result.User.ID)
}

func TestVerifyEmailCommand_ExpiredToken(t *testing.T) {
	repo, rawToken, _ := newVerifySetup(t)
	eventBus := cqrs.NewEventBus[EmailVerifiedEvent]()
	handler := NewVerifyEmailCommandHandler(repo, eventBus)

	// Expire the token.
	for _, vt := range repo.verificationTokens {
		vt.ExpiresAt = time.Now().Add(-1 * time.Hour)
	}

	_, err := handler.Handle(context.Background(), VerifyEmailCommand{Token: rawToken})
	require.Error(t, err)

	var unauthErr *cqrs.UnauthorizedError
	assert.True(t, errors.As(err, &unauthErr), "expected UnauthorizedError, got %T: %v", err, err)
}

func TestVerifyEmailCommand_InvalidToken(t *testing.T) {
	repo := newMockRepository()

	// Create user but no matching token.
	user := &User{
		ID:            "user-1",
		Email:         "verify@example.com",
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	repo.users[user.ID] = user
	repo.usersByEmail[user.Email] = user

	eventBus := cqrs.NewEventBus[EmailVerifiedEvent]()
	handler := NewVerifyEmailCommandHandler(repo, eventBus)

	// Use a random but valid hex string (64 hex chars = 32 bytes).
	_, err := handler.Handle(context.Background(), VerifyEmailCommand{
		Token: "aabbccddee00112233445566778899aabbccddee00112233445566778899aabb",
	})
	require.Error(t, err)

	var unauthErr *cqrs.UnauthorizedError
	assert.True(t, errors.As(err, &unauthErr), "expected UnauthorizedError, got %T: %v", err, err)
}

func TestVerifyEmailCommand_WelcomeEventPublished(t *testing.T) {
	repo, rawToken, user := newVerifySetup(t)

	eventBus := cqrs.NewEventBus[EmailVerifiedEvent]()
	recorder := &testEventHandler{}
	eventBus.Register(recorder)

	handler := NewVerifyEmailCommandHandler(repo, eventBus)

	_, err := handler.Handle(context.Background(), VerifyEmailCommand{Token: rawToken})
	require.NoError(t, err)

	require.Len(t, recorder.events, 1, "one EmailVerifiedEvent should be published")
	assert.Equal(t, user.ID, recorder.events[0].UserID)
	assert.Equal(t, user.Email, recorder.events[0].Email)
}

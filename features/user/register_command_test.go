package user

import (
	"context"
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"arsen/pkg/config"
	"arsen/pkg/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validRegisterCmd() RegisterCommand {
	return RegisterCommand{
		Email:    "test@example.com",
		Password: "Str0ng!Pass",
	}
}

func testConfig() *config.Config {
	return &config.Config{AppBaseURL: "http://localhost:8080"}
}

func TestRegisterCommand_ValidRegistration(t *testing.T) {
	repo := newMockRepository()
	sender := &mockEmailSender{}
	handler := NewRegisterCommandHandler(repo, sender, testConfig())

	result, err := handler.Handle(context.Background(), validRegisterCmd())
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.User)

	assert.False(t, result.User.EmailVerified, "new user should have emailVerified=false")
	assert.Equal(t, "test@example.com", result.User.Email)
	assert.NotEmpty(t, result.User.ID)
}

func TestRegisterCommand_DuplicateEmail(t *testing.T) {
	repo := newMockRepository()
	sender := &mockEmailSender{}
	handler := NewRegisterCommandHandler(repo, sender, testConfig())

	// Pre-populate an existing user.
	repo.usersByEmail["test@example.com"] = &User{
		ID:    "existing-id",
		Email: "test@example.com",
	}
	repo.users["existing-id"] = repo.usersByEmail["test@example.com"]

	_, err := handler.Handle(context.Background(), validRegisterCmd())
	require.Error(t, err)

	var conflictErr *cqrs.ConflictError
	assert.True(t, errors.As(err, &conflictErr), "expected ConflictError, got %T: %v", err, err)
}

func TestRegisterCommand_WeakPassword(t *testing.T) {
	cmd := RegisterCommand{Email: "a@b.com", Password: "short"}
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

func TestRegisterCommand_InvalidEmail(t *testing.T) {
	cmd := RegisterCommand{Email: "not-an-email", Password: "Str0ng!Pass"}
	err := cmd.Validate()
	require.Error(t, err)

	var valErr *cqrs.ValidationError
	require.True(t, errors.As(err, &valErr), "expected ValidationError, got %T: %v", err, err)

	found := false
	for _, fe := range valErr.Fields {
		if fe.Field == "email" {
			found = true
			break
		}
	}
	assert.True(t, found, "should have email validation error")
}

func TestRegisterCommand_BcryptHash(t *testing.T) {
	repo := newMockRepository()
	sender := &mockEmailSender{}
	handler := NewRegisterCommandHandler(repo, sender, testConfig())

	result, err := handler.Handle(context.Background(), validRegisterCmd())
	require.NoError(t, err)

	// Verify bcrypt prefix.
	assert.True(t,
		strings.HasPrefix(result.User.PasswordHash, "$2a$") || strings.HasPrefix(result.User.PasswordHash, "$2b$"),
		"password hash should have bcrypt prefix, got: %s", result.User.PasswordHash,
	)

	// Verify the hash matches the original password.
	err = bcrypt.CompareHashAndPassword([]byte(result.User.PasswordHash), []byte("Str0ng!Pass"))
	assert.NoError(t, err, "bcrypt comparison should succeed for the original password")
}

func TestRegisterCommand_VerificationTokenCreated(t *testing.T) {
	repo := newMockRepository()
	sender := &mockEmailSender{}
	handler := NewRegisterCommandHandler(repo, sender, testConfig())

	result, err := handler.Handle(context.Background(), validRegisterCmd())
	require.NoError(t, err)

	// Verify a verification token was stored for this user.
	found := false
	for _, vt := range repo.verificationTokens {
		if vt.UserID == result.User.ID {
			found = true
			assert.NotEmpty(t, vt.TokenHash, "token hash should not be empty")
			assert.False(t, vt.ExpiresAt.IsZero(), "expires_at should be set")
			break
		}
	}
	assert.True(t, found, "verification token should be created for the new user")
	assert.NotEmpty(t, result.RawToken, "raw token should be returned in the result")
}

func TestRegisterCommand_EmailSent(t *testing.T) {
	repo := newMockRepository()
	sender := &mockEmailSender{}
	handler := NewRegisterCommandHandler(repo, sender, testConfig())

	_, err := handler.Handle(context.Background(), validRegisterCmd())
	require.NoError(t, err)

	require.Len(t, sender.sent, 1, "exactly one email should be sent")
	assert.Equal(t, "test@example.com", sender.sent[0].To)
	assert.Equal(t, "Verify your email", sender.sent[0].Subject)
	assert.Contains(t, sender.sent[0].HTML, "http://localhost:8080/api/users/verify?token=")
}

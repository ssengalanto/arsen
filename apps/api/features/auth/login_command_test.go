package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"arsen/features/user"
	"arsen/pkg/config"
	"arsen/pkg/cqrs"
	"arsen/pkg/jwt"
)

// ---------------------------------------------------------------------------
// Mock: user.GetUserByEmailQuery handler
// ---------------------------------------------------------------------------

type mockUserByEmailHandler struct {
	users         map[string]*user.User // keyed by email
	getByEmailErr error
}

func (m *mockUserByEmailHandler) Handle(_ context.Context, query user.GetUserByEmailQuery) (*user.GetUserByEmailResult, error) {
	if m.getByEmailErr != nil {
		return nil, m.getByEmailErr
	}
	u, ok := m.users[query.Email]
	if !ok {
		return nil, cqrs.NewNotFoundError("User", query.Email)
	}
	return &user.GetUserByEmailResult{User: u}, nil
}

// ---------------------------------------------------------------------------
// Mock: auth.Repository
// ---------------------------------------------------------------------------

type mockAuthRepo struct {
	tokens           []*RefreshToken
	createErr        error
	getByHashErr     error
	revokeFamilyErr  error
	revokeAllUserErr error
}

func (m *mockAuthRepo) CreateRefreshToken(_ context.Context, token *RefreshToken) error {
	if m.createErr != nil {
		return m.createErr
	}
	token.ID = "rt-mock-id"
	token.CreatedAt = time.Now()
	m.tokens = append(m.tokens, token)
	return nil
}

func (m *mockAuthRepo) GetRefreshTokenByHash(_ context.Context, _ []byte) (*RefreshToken, error) {
	if m.getByHashErr != nil {
		return nil, m.getByHashErr
	}
	return nil, cqrs.NewNotFoundError("RefreshToken", "hash")
}

func (m *mockAuthRepo) RevokeRefreshTokenFamily(_ context.Context, _ string) error {
	return m.revokeFamilyErr
}

func (m *mockAuthRepo) RevokeAllUserRefreshTokens(_ context.Context, _ string) error {
	return m.revokeAllUserErr
}

// ---------------------------------------------------------------------------
// Test setup helper
// ---------------------------------------------------------------------------

func setupLoginTest(t *testing.T) (*LoginCommandHandler, *mockUserByEmailHandler, *mockAuthRepo) {
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
	mockHandler := &mockUserByEmailHandler{users: make(map[string]*user.User)}
	userByEmail := cqrs.NewQueryBus[user.GetUserByEmailQuery, *user.GetUserByEmailResult](mockHandler)
	authRepo := &mockAuthRepo{}
	handler := NewLoginCommandHandler(userByEmail, authRepo, jwtSvc, cfg)
	return handler, mockHandler, authRepo
}

func createTestUser(t *testing.T, email, password string, verified bool) *user.User { //nolint:unparam // test helper parameterized for clarity
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	require.NoError(t, err)
	return &user.User{
		ID:            "user-123",
		Email:         email,
		PasswordHash:  string(hash),
		EmailVerified: verified,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// ---------------------------------------------------------------------------
// T063: LoginCommandHandler tests
// ---------------------------------------------------------------------------

func TestLoginCommandHandler_ValidCredentials(t *testing.T) {
	handler, userMock, _ := setupLoginTest(t)

	testUser := createTestUser(t, "alice@example.com", "Test123!!", true)
	userMock.users[testUser.Email] = testUser

	result, err := handler.Handle(context.Background(), LoginCommand{
		Email:    "alice@example.com",
		Password: "Test123!!",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken, "expected non-empty access token")
	assert.NotEmpty(t, result.RefreshToken, "expected non-empty refresh token")
	assert.Equal(t, 900, result.ExpiresIn, "expected ExpiresIn to equal 15 minutes in seconds")
}

func TestLoginCommandHandler_WrongPassword(t *testing.T) {
	handler, userMock, _ := setupLoginTest(t)

	testUser := createTestUser(t, "alice@example.com", "Test123!!", true)
	userMock.users[testUser.Email] = testUser

	result, err := handler.Handle(context.Background(), LoginCommand{
		Email:    "alice@example.com",
		Password: "WrongPassword!!",
	})

	assert.Nil(t, result)
	require.Error(t, err)

	var unauthorizedErr *cqrs.UnauthorizedError
	require.True(t, errors.As(err, &unauthorizedErr), "expected UnauthorizedError, got %T: %v", err, err)
	assert.Equal(t, "invalid email or password", unauthorizedErr.Detail)
}

func TestLoginCommandHandler_NonExistentUser(t *testing.T) {
	handler, _, _ := setupLoginTest(t)

	result, err := handler.Handle(context.Background(), LoginCommand{
		Email:    "nobody@example.com",
		Password: "Test123!!",
	})

	assert.Nil(t, result)
	require.Error(t, err)

	var unauthorizedErr *cqrs.UnauthorizedError
	require.True(t, errors.As(err, &unauthorizedErr), "expected UnauthorizedError, got %T: %v", err, err)
	// Should return the same message as wrong password to prevent user enumeration.
	assert.Equal(t, "invalid email or password", unauthorizedErr.Detail)
}

func TestLoginCommandHandler_UnverifiedEmail(t *testing.T) {
	handler, userMock, _ := setupLoginTest(t)

	testUser := createTestUser(t, "alice@example.com", "Test123!!", false)
	userMock.users[testUser.Email] = testUser

	result, err := handler.Handle(context.Background(), LoginCommand{
		Email:    "alice@example.com",
		Password: "Test123!!",
	})

	assert.Nil(t, result)
	require.Error(t, err)

	var forbiddenErr *cqrs.ForbiddenError
	require.True(t, errors.As(err, &forbiddenErr), "expected ForbiddenError, got %T: %v", err, err)
	assert.Contains(t, forbiddenErr.Detail, "verify your email")
}

func TestLoginCommandHandler_TimingAntiEnumeration(t *testing.T) {
	handler, userMock, _ := setupLoginTest(t)

	// Create a real user with bcrypt-hashed password.
	testUser := createTestUser(t, "alice@example.com", "Test123!!", true)
	userMock.users[testUser.Email] = testUser

	// Measure time for wrong password (user exists).
	start1 := time.Now()
	_, _ = handler.Handle(context.Background(), LoginCommand{
		Email:    "alice@example.com",
		Password: "WrongPassword!!",
	})
	wrongPasswordDuration := time.Since(start1)

	// Measure time for non-existent user (dummy hash comparison).
	start2 := time.Now()
	_, _ = handler.Handle(context.Background(), LoginCommand{
		Email:    "nobody@example.com",
		Password: "SomePassword!!",
	})
	nonExistentDuration := time.Since(start2)

	// Both should take comparable time (within 200ms — bcrypt is slow).
	diff := wrongPasswordDuration - nonExistentDuration
	if diff < 0 {
		diff = -diff
	}
	assert.Less(t, diff, 200*time.Millisecond,
		"login timing for non-existent user (%v) and wrong password (%v) should be within 200ms",
		nonExistentDuration, wrongPasswordDuration)
}

func TestLoginCommandHandler_RefreshTokenStored(t *testing.T) {
	handler, userMock, authRepo := setupLoginTest(t)

	testUser := createTestUser(t, "alice@example.com", "Test123!!", true)
	userMock.users[testUser.Email] = testUser

	_, err := handler.Handle(context.Background(), LoginCommand{
		Email:    "alice@example.com",
		Password: "Test123!!",
	})
	require.NoError(t, err)

	require.Len(t, authRepo.tokens, 1, "expected exactly one refresh token to be stored")
	assert.Equal(t, "user-123", authRepo.tokens[0].UserID)
	assert.NotEmpty(t, authRepo.tokens[0].TokenHash, "expected non-empty token hash")
	assert.NotEmpty(t, authRepo.tokens[0].FamilyID, "expected non-empty family ID")
	assert.False(t, authRepo.tokens[0].ExpiresAt.IsZero(), "expected non-zero expiry")
}

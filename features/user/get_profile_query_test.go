package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"arsen/pkg/cqrs"
)

func TestGetProfileQueryHandler_ValidUser(t *testing.T) {
	repo := newMockRepository()

	// Seed a user into the mock repository.
	u := &User{
		Email:         "alice@example.com",
		PasswordHash:  "hashed",
		EmailVerified: true,
	}
	require.NoError(t, repo.Create(context.Background(), u))

	handler := NewGetProfileQueryHandler(repo)
	bus := cqrs.NewQueryBus[GetProfileQuery, *GetProfileResult](handler)

	result, err := bus.Ask(context.Background(), GetProfileQuery{UserID: u.ID})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.User)

	assert.Equal(t, u.ID, result.User.ID)
	assert.Equal(t, "alice@example.com", result.User.Email)
	assert.True(t, result.User.EmailVerified)
	assert.False(t, result.User.CreatedAt.IsZero())
}

func TestGetProfileQueryHandler_NonExistentUser(t *testing.T) {
	repo := newMockRepository()

	handler := NewGetProfileQueryHandler(repo)
	bus := cqrs.NewQueryBus[GetProfileQuery, *GetProfileResult](handler)

	result, err := bus.Ask(context.Background(), GetProfileQuery{UserID: "non-existent-id"})
	require.Error(t, err)
	assert.Nil(t, result)

	var notFoundErr *cqrs.NotFoundError
	assert.True(t, errors.As(err, &notFoundErr), "expected NotFoundError, got %T: %v", err, err)
}

func TestGetProfileQuery_Validate_EmptyUserID(t *testing.T) {
	q := GetProfileQuery{UserID: ""}
	err := q.Validate()
	require.Error(t, err)

	var validationErr *cqrs.ValidationError
	require.True(t, errors.As(err, &validationErr))
	assert.Len(t, validationErr.Fields, 1)
	assert.Equal(t, "userId", validationErr.Fields[0].Field)
}

func TestGetProfileQuery_Validate_ValidUserID(t *testing.T) {
	q := GetProfileQuery{UserID: "some-id"}
	assert.NoError(t, q.Validate())
}

// Ensure the test file uses the time import (compile guard).
var _ = time.Now

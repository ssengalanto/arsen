package user

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"arsen/pkg/cqrs"
)

func TestGetUserByEmailQueryHandler_ValidUser(t *testing.T) {
	repo := newMockRepository()

	u := &User{
		Email:         "alice@example.com",
		PasswordHash:  "hashed",
		EmailVerified: true,
	}
	require.NoError(t, repo.Create(context.Background(), u))

	handler := NewGetUserByEmailQueryHandler(repo)
	bus := cqrs.NewQueryBus[GetUserByEmailQuery, *GetUserByEmailResult](handler)

	result, err := bus.Ask(context.Background(), GetUserByEmailQuery{Email: "alice@example.com"})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.User)

	assert.Equal(t, u.ID, result.User.ID)
	assert.Equal(t, "alice@example.com", result.User.Email)
	assert.True(t, result.User.EmailVerified)
	assert.False(t, result.User.CreatedAt.IsZero())
}

func TestGetUserByEmailQueryHandler_NonExistentUser(t *testing.T) {
	repo := newMockRepository()

	handler := NewGetUserByEmailQueryHandler(repo)
	bus := cqrs.NewQueryBus[GetUserByEmailQuery, *GetUserByEmailResult](handler)

	result, err := bus.Ask(context.Background(), GetUserByEmailQuery{Email: "nobody@example.com"})
	require.Error(t, err)
	assert.Nil(t, result)

	var notFoundErr *cqrs.NotFoundError
	assert.True(t, errors.As(err, &notFoundErr), "expected NotFoundError, got %T: %v", err, err)
}

func TestGetUserByEmailQueryHandler_RepositoryError(t *testing.T) {
	repo := newMockRepository()
	repo.getByEmailErr = fmt.Errorf("connection refused")

	handler := NewGetUserByEmailQueryHandler(repo)
	bus := cqrs.NewQueryBus[GetUserByEmailQuery, *GetUserByEmailResult](handler)

	result, err := bus.Ask(context.Background(), GetUserByEmailQuery{Email: "alice@example.com"})
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestGetUserByEmailQuery_Validate_EmptyEmail(t *testing.T) {
	q := GetUserByEmailQuery{Email: ""}
	err := q.Validate()
	require.Error(t, err)

	var validationErr *cqrs.ValidationError
	require.True(t, errors.As(err, &validationErr))
	assert.Len(t, validationErr.Fields, 1)
	assert.Equal(t, "email", validationErr.Fields[0].Field)
}

func TestGetUserByEmailQuery_Validate_ValidEmail(t *testing.T) {
	q := GetUserByEmailQuery{Email: "alice@example.com"}
	assert.NoError(t, q.Validate())
}

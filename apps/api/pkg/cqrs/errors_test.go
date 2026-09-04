package cqrs_test

import (
	"errors"
	"fmt"
	"testing"

	"arsen/pkg/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotFoundError_Message(t *testing.T) {
	err := cqrs.NewNotFoundError("User", "abc-123")
	assert.Equal(t, `User with identifier "abc-123" not found`, err.Error())
}

func TestConflictError_Message(t *testing.T) {
	err := cqrs.NewConflictError("Order", "ord-1", "already shipped")
	assert.Equal(t, `conflict on Order with identifier "ord-1": already shipped`, err.Error())
}

func TestValidationError_MultipleFields(t *testing.T) {
	err := cqrs.NewValidationError([]cqrs.FieldError{
		{Field: "email", Detail: "required"},
		{Field: "name", Detail: "too short"},
		{Field: "age", Detail: "must be positive"},
	})
	assert.Equal(t, "validation failed on 3 fields", err.Error())
}

func TestValidationError_SingleField(t *testing.T) {
	err := cqrs.NewValidationError([]cqrs.FieldError{
		{Field: "email", Detail: "required"},
	})
	assert.Equal(t, "validation failed: email: required", err.Error())
}

func TestValidationError_NoFields(t *testing.T) {
	err := cqrs.NewValidationError(nil)
	assert.Equal(t, "validation failed", err.Error())
}

func TestUnauthorizedError_Message(t *testing.T) {
	err := cqrs.NewUnauthorizedError("token expired")
	assert.Equal(t, "unauthorized: token expired", err.Error())
}

func TestForbiddenError_Message(t *testing.T) {
	err := cqrs.NewForbiddenError("insufficient role")
	assert.Equal(t, "forbidden: insufficient role", err.Error())
}

func TestErrorsAs_NotFoundError(t *testing.T) {
	original := cqrs.NewNotFoundError("Account", "42")
	wrapped := fmt.Errorf("lookup: %w", original)

	var target *cqrs.NotFoundError
	require.True(t, errors.As(wrapped, &target))
	assert.Equal(t, "Account", target.ResourceType)
	assert.Equal(t, "42", target.Identifier)
}

func TestErrorsAs_ConflictError(t *testing.T) {
	original := cqrs.NewConflictError("Session", "s-1", "duplicate")
	wrapped := fmt.Errorf("save: %w", original)

	var target *cqrs.ConflictError
	require.True(t, errors.As(wrapped, &target))
	assert.Equal(t, "Session", target.ResourceType)
	assert.Equal(t, "duplicate", target.Detail)
}

func TestErrorsAs_ValidationError(t *testing.T) {
	original := cqrs.NewValidationError([]cqrs.FieldError{{Field: "x", Detail: "bad"}})
	wrapped := fmt.Errorf("input: %w", original)

	var target *cqrs.ValidationError
	require.True(t, errors.As(wrapped, &target))
	assert.Len(t, target.Fields, 1)
}

func TestErrorsAs_UnauthorizedError(t *testing.T) {
	original := cqrs.NewUnauthorizedError("no token")
	wrapped := fmt.Errorf("auth: %w", original)

	var target *cqrs.UnauthorizedError
	require.True(t, errors.As(wrapped, &target))
	assert.Equal(t, "no token", target.Detail)
}

func TestErrorsAs_ForbiddenError(t *testing.T) {
	original := cqrs.NewForbiddenError("admin only")
	wrapped := fmt.Errorf("access: %w", original)

	var target *cqrs.ForbiddenError
	require.True(t, errors.As(wrapped, &target))
	assert.Equal(t, "admin only", target.Detail)
}

package user

import (
	"context"

	"arsen/pkg/cqrs"
)

type GetUserByEmailQuery struct {
	Email string
}

type GetUserByEmailResult struct {
	User *User
}

func (q GetUserByEmailQuery) Validate() error {
	if q.Email == "" {
		return cqrs.NewValidationError([]cqrs.FieldError{
			{Field: "email", Detail: "is required"},
		})
	}
	return nil
}

type GetUserByEmailQueryHandler struct {
	repo Repository
}

func NewGetUserByEmailQueryHandler(repo Repository) *GetUserByEmailQueryHandler {
	return &GetUserByEmailQueryHandler{repo: repo}
}

func (h *GetUserByEmailQueryHandler) Handle(ctx context.Context, query GetUserByEmailQuery) (*GetUserByEmailResult, error) {
	u, err := h.repo.GetByEmail(ctx, query.Email)
	if err != nil {
		return nil, err
	}
	return &GetUserByEmailResult{User: u}, nil
}

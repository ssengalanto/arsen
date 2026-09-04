package user

import (
	"context"

	"arsen/pkg/cqrs"
)

type GetProfileQuery struct {
	UserID string
}

type GetProfileResult struct {
	User *User
}

func (q GetProfileQuery) Validate() error {
	if q.UserID == "" {
		return cqrs.NewValidationError([]cqrs.FieldError{
			{Field: "userId", Detail: "is required"},
		})
	}
	return nil
}

type GetProfileQueryHandler struct {
	repo Repository
}

func NewGetProfileQueryHandler(repo Repository) *GetProfileQueryHandler {
	return &GetProfileQueryHandler{repo: repo}
}

func (h *GetProfileQueryHandler) Handle(ctx context.Context, query GetProfileQuery) (*GetProfileResult, error) {
	user, err := h.repo.GetByID(ctx, query.UserID)
	if err != nil {
		return nil, err
	}
	return &GetProfileResult{User: user}, nil
}

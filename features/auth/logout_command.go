package auth

import (
	"context"
	"fmt"

	"arsen/pkg/cqrs"
)

type LogoutCommand struct {
	UserID string
}

func (c LogoutCommand) Validate() error {
	var fields []cqrs.FieldError
	if c.UserID == "" {
		fields = append(fields, cqrs.FieldError{Field: "userID", Detail: "is required"})
	}
	if len(fields) > 0 {
		return cqrs.NewValidationError(fields)
	}
	return nil
}

type LogoutCommandHandler struct {
	repo Repository
}

func NewLogoutCommandHandler(repo Repository) *LogoutCommandHandler {
	return &LogoutCommandHandler{repo: repo}
}

func (h *LogoutCommandHandler) Handle(ctx context.Context, cmd LogoutCommand) (cqrs.Unit, error) {
	if err := h.repo.RevokeAllUserRefreshTokens(ctx, cmd.UserID); err != nil {
		return cqrs.Unit{}, fmt.Errorf("logout: failed to revoke refresh tokens: %w", err)
	}
	return cqrs.Unit{}, nil
}

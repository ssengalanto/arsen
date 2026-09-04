package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"arsen/pkg/cqrs"
	"arsen/pkg/token"
)

type ResetPasswordCommand struct {
	Token    string
	Password string
}

func (c ResetPasswordCommand) Validate() error {
	var fields []cqrs.FieldError
	if c.Token == "" {
		fields = append(fields, cqrs.FieldError{Field: "token", Detail: "is required"})
	}
	if c.Password == "" {
		fields = append(fields, cqrs.FieldError{Field: "password", Detail: "is required"})
	} else if errs := validatePassword(c.Password); len(errs) > 0 {
		for _, e := range errs {
			fields = append(fields, cqrs.FieldError{Field: "password", Detail: e})
		}
	}
	if len(fields) > 0 {
		return cqrs.NewValidationError(fields)
	}
	return nil
}

type ResetPasswordCommandHandler struct {
	repo    Repository
	revoker RefreshTokenRevoker
}

func NewResetPasswordCommandHandler(repo Repository, revoker RefreshTokenRevoker) *ResetPasswordCommandHandler {
	return &ResetPasswordCommandHandler{repo: repo, revoker: revoker}
}

func (h *ResetPasswordCommandHandler) Handle(ctx context.Context, cmd ResetPasswordCommand) (cqrs.Unit, error) {
	// Hash the submitted token.
	hash, err := token.Hash(cmd.Token)
	if err != nil {
		return cqrs.Unit{}, cqrs.NewUnauthorizedError("invalid or expired reset token")
	}

	// Look up the reset token (unused only).
	prt, err := h.repo.GetPasswordResetTokenByHash(ctx, hash)
	if err != nil {
		var notFound *cqrs.NotFoundError
		if errors.As(err, &notFound) {
			return cqrs.Unit{}, cqrs.NewUnauthorizedError("invalid or expired reset token")
		}
		return cqrs.Unit{}, fmt.Errorf("reset password: failed to look up token: %w", err)
	}

	// Check expiry.
	if prt.ExpiresAt.Before(time.Now()) {
		return cqrs.Unit{}, cqrs.NewUnauthorizedError("invalid or expired reset token")
	}

	// Hash the new password.
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), 12)
	if err != nil {
		return cqrs.Unit{}, fmt.Errorf("reset password: failed to hash password: %w", err)
	}

	// Update user's password.
	if err := h.repo.UpdatePasswordHash(ctx, prt.UserID, string(passwordHash)); err != nil {
		return cqrs.Unit{}, fmt.Errorf("reset password: failed to update password: %w", err)
	}

	// Mark the token as used.
	if err := h.repo.InvalidateUserPasswordResetTokens(ctx, prt.UserID); err != nil {
		return cqrs.Unit{}, fmt.Errorf("reset password: failed to invalidate tokens: %w", err)
	}

	// Revoke all refresh token families for this user.
	if err := h.revoker.RevokeAllUserRefreshTokens(ctx, prt.UserID); err != nil {
		return cqrs.Unit{}, fmt.Errorf("reset password: failed to revoke refresh tokens: %w", err)
	}

	return cqrs.Unit{}, nil
}

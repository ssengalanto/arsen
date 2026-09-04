package user

import (
	"context"
	"fmt"
	"time"

	"arsen/pkg/cqrs"
	"arsen/pkg/token"
)

type VerifyEmailCommand struct {
	Token string
}

type VerifyEmailResult struct {
	User *User
}

func (c VerifyEmailCommand) Validate() error {
	if c.Token == "" {
		return cqrs.NewValidationError([]cqrs.FieldError{
			{Field: "token", Detail: "is required"},
		})
	}
	return nil
}

type VerifyEmailCommandHandler struct {
	repo     Repository
	eventBus *cqrs.EventBus[EmailVerifiedEvent]
}

func NewVerifyEmailCommandHandler(repo Repository, eventBus *cqrs.EventBus[EmailVerifiedEvent]) *VerifyEmailCommandHandler {
	return &VerifyEmailCommandHandler{
		repo:     repo,
		eventBus: eventBus,
	}
}

func (h *VerifyEmailCommandHandler) Handle(ctx context.Context, cmd VerifyEmailCommand) (*VerifyEmailResult, error) {
	// Hash the submitted token.
	hash, err := token.Hash(cmd.Token)
	if err != nil {
		return nil, cqrs.NewUnauthorizedError("invalid or expired verification token")
	}

	// Look up the verification token.
	vt, err := h.repo.GetVerificationTokenByHash(ctx, hash)
	if err != nil {
		return nil, cqrs.NewUnauthorizedError("invalid or expired verification token")
	}

	// Check expiration.
	if vt.ExpiresAt.Before(time.Now()) {
		return nil, cqrs.NewUnauthorizedError("invalid or expired verification token")
	}

	// Invalidate all verification tokens for the user.
	if err := h.repo.InvalidateUserVerificationTokens(ctx, vt.UserID); err != nil {
		return nil, fmt.Errorf("verify email: failed to invalidate tokens: %w", err)
	}

	// Mark the user's email as verified.
	if err := h.repo.UpdateEmailVerified(ctx, vt.UserID, true); err != nil {
		return nil, fmt.Errorf("verify email: failed to update email verified: %w", err)
	}

	// Fetch updated user.
	u, err := h.repo.GetByID(ctx, vt.UserID)
	if err != nil {
		return nil, fmt.Errorf("verify email: failed to get user: %w", err)
	}

	// Publish event.
	if err := h.eventBus.Publish(ctx, EmailVerifiedEvent{
		UserID: u.ID,
		Email:  u.Email,
	}); err != nil {
		return nil, fmt.Errorf("verify email: failed to publish event: %w", err)
	}

	return &VerifyEmailResult{User: u}, nil
}

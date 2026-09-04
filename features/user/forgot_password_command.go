package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"arsen/pkg/config"
	"arsen/pkg/cqrs"
	"arsen/pkg/email"
	"arsen/pkg/token"
)

type ForgotPasswordCommand struct {
	Email string
}

func (c ForgotPasswordCommand) Validate() error {
	if c.Email == "" {
		return cqrs.NewValidationError([]cqrs.FieldError{
			{Field: "email", Detail: "is required"},
		})
	}
	return nil
}

type ForgotPasswordCommandHandler struct {
	repo        Repository
	emailSender email.Sender
	cfg         *config.Config
}

func NewForgotPasswordCommandHandler(repo Repository, emailSender email.Sender, cfg *config.Config) *ForgotPasswordCommandHandler {
	return &ForgotPasswordCommandHandler{repo: repo, emailSender: emailSender, cfg: cfg}
}

func (h *ForgotPasswordCommandHandler) Handle(ctx context.Context, cmd ForgotPasswordCommand) (cqrs.Unit, error) {
	// Anti-enumeration: always return success regardless of outcome.
	u, err := h.repo.GetByEmail(ctx, cmd.Email)
	if err != nil {
		var notFound *cqrs.NotFoundError
		if errors.As(err, &notFound) {
			return cqrs.Unit{}, nil
		}
		return cqrs.Unit{}, fmt.Errorf("forgot password: failed to look up user: %w", err)
	}

	// Don't send reset email to unverified users.
	if !u.EmailVerified {
		return cqrs.Unit{}, nil
	}

	// Invalidate previous reset tokens.
	if err := h.repo.InvalidateUserPasswordResetTokens(ctx, u.ID); err != nil {
		return cqrs.Unit{}, fmt.Errorf("forgot password: failed to invalidate tokens: %w", err)
	}

	// Generate new reset token (256-bit, 1-hour expiry).
	rawToken, tokenHash, err := token.Generate()
	if err != nil {
		return cqrs.Unit{}, fmt.Errorf("forgot password: failed to generate token: %w", err)
	}

	prt := &PasswordResetToken{
		UserID:    u.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	if err := h.repo.CreatePasswordResetToken(ctx, prt); err != nil {
		return cqrs.Unit{}, fmt.Errorf("forgot password: failed to store token: %w", err)
	}

	// Send reset email.
	resetURL := h.cfg.AppBaseURL + "/api/users/reset-password?token=" + rawToken
	html := fmt.Sprintf(
		`<p>You requested a password reset. Click the link below to set a new password:</p><p><a href="%s">Reset Password</a></p><p>This link expires in 1 hour. If you didn't request this, you can safely ignore this email.</p>`,
		resetURL,
	)
	if err := h.emailSender.Send(ctx, email.SendParams{
		To:      u.Email,
		Subject: "Reset your password",
		HTML:    html,
	}); err != nil {
		return cqrs.Unit{}, fmt.Errorf("forgot password: failed to send email: %w", err)
	}

	return cqrs.Unit{}, nil
}

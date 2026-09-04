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

type ResendVerificationCommand struct {
	Email string
}

func (c ResendVerificationCommand) Validate() error {
	if c.Email == "" {
		return cqrs.NewValidationError([]cqrs.FieldError{
			{Field: "email", Detail: "is required"},
		})
	}
	return nil
}

type ResendVerificationCommandHandler struct {
	repo        Repository
	emailSender email.Sender
	cfg         *config.Config
}

func NewResendVerificationCommandHandler(repo Repository, emailSender email.Sender, cfg *config.Config) *ResendVerificationCommandHandler {
	return &ResendVerificationCommandHandler{
		repo:        repo,
		emailSender: emailSender,
		cfg:         cfg,
	}
}

func (h *ResendVerificationCommandHandler) Handle(ctx context.Context, cmd ResendVerificationCommand) (cqrs.Unit, error) {
	// Look up user; if not found or already verified, return success silently.
	u, err := h.repo.GetByEmail(ctx, cmd.Email)
	if err != nil {
		var notFound *cqrs.NotFoundError
		if errors.As(err, &notFound) {
			return cqrs.Unit{}, nil
		}
		return cqrs.Unit{}, fmt.Errorf("resend verification: failed to get user: %w", err)
	}
	if u.EmailVerified {
		return cqrs.Unit{}, nil
	}

	// Invalidate existing verification tokens.
	if err := h.repo.InvalidateUserVerificationTokens(ctx, u.ID); err != nil {
		return cqrs.Unit{}, fmt.Errorf("resend verification: failed to invalidate tokens: %w", err)
	}

	// Generate a new verification token.
	rawToken, tokenHash, err := token.Generate()
	if err != nil {
		return cqrs.Unit{}, fmt.Errorf("resend verification: failed to generate token: %w", err)
	}

	vt := &VerificationToken{
		UserID:    u.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := h.repo.CreateVerificationToken(ctx, vt); err != nil {
		return cqrs.Unit{}, fmt.Errorf("resend verification: failed to store token: %w", err)
	}

	// Send verification email.
	verifyURL := h.cfg.AppBaseURL + "/api/users/verify?token=" + rawToken
	html := fmt.Sprintf( //nolint:gocritic // HTML template uses quoted URL, not Go quoting
		`<p>Please verify your email by clicking the link below:</p><p><a href="%s">Verify Email</a></p><p>This link expires in 24 hours.</p>`,
		verifyURL,
	)
	if err := h.emailSender.Send(ctx, email.SendParams{
		To:      u.Email,
		Subject: "Verify your email",
		HTML:    html,
	}); err != nil {
		return cqrs.Unit{}, fmt.Errorf("resend verification: failed to send email: %w", err)
	}

	return cqrs.Unit{}, nil
}

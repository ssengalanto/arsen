package user

import (
	"context"
	"log/slog"

	"arsen/pkg/email"
)

type EmailVerifiedEvent struct {
	UserID string
	Email  string
}

type WelcomeEmailHandler struct {
	emailSender email.Sender
}

func NewWelcomeEmailHandler(emailSender email.Sender) *WelcomeEmailHandler {
	return &WelcomeEmailHandler{emailSender: emailSender}
}

func (h *WelcomeEmailHandler) Handle(ctx context.Context, event EmailVerifiedEvent) error {
	html := `<p>Welcome to Arsen! Your email has been verified.</p><p>You now have full access to your account.</p>`

	if err := h.emailSender.Send(ctx, email.SendParams{
		To:      event.Email,
		Subject: "Welcome!",
		HTML:    html,
	}); err != nil {
		slog.ErrorContext(ctx, "failed to send welcome email",
			slog.String("user_id", event.UserID),
			slog.String("email", event.Email),
			slog.String("error", err.Error()),
		)
		return nil // best-effort: don't propagate
	}

	return nil
}

package user

import (
	"context"
	"fmt"
	"regexp"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"arsen/pkg/config"
	"arsen/pkg/cqrs"
	"arsen/pkg/email"
	"arsen/pkg/token"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type RegisterCommand struct {
	Email    string
	Password string
}

type RegisterResult struct {
	User     *User
	RawToken string
}

func (c RegisterCommand) Validate() error {
	var fields []cqrs.FieldError

	if c.Email == "" {
		fields = append(fields, cqrs.FieldError{Field: "email", Detail: "is required"})
	} else if !emailRegex.MatchString(c.Email) {
		fields = append(fields, cqrs.FieldError{Field: "email", Detail: "must be a valid email address"})
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

func validatePassword(password string) []string {
	var errs []string

	if len(password) < 8 {
		errs = append(errs, "must be at least 8 characters")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	if !hasUpper {
		errs = append(errs, "must contain at least 1 uppercase letter")
	}
	if !hasLower {
		errs = append(errs, "must contain at least 1 lowercase letter")
	}
	if !hasDigit {
		errs = append(errs, "must contain at least 1 digit")
	}
	if !hasSpecial {
		errs = append(errs, "must contain at least 1 special character")
	}

	return errs
}

type RegisterCommandHandler struct {
	repo        Repository
	emailSender email.Sender
	cfg         *config.Config
}

func NewRegisterCommandHandler(repo Repository, emailSender email.Sender, cfg *config.Config) *RegisterCommandHandler {
	return &RegisterCommandHandler{
		repo:        repo,
		emailSender: emailSender,
		cfg:         cfg,
	}
}

func (h *RegisterCommandHandler) Handle(ctx context.Context, cmd RegisterCommand) (*RegisterResult, error) {
	// Check email uniqueness.
	if _, err := h.repo.GetByEmail(ctx, cmd.Email); err == nil {
		return nil, cqrs.NewConflictError("User", cmd.Email, "email already registered")
	}

	// Hash password with bcrypt cost 12.
	hash, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("register: failed to hash password: %w", err)
	}

	// Create user.
	u := &User{
		Email:        cmd.Email,
		PasswordHash: string(hash),
	}
	if err := h.repo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("register: failed to create user: %w", err)
	}

	// Generate verification token.
	rawToken, tokenHash, err := token.Generate()
	if err != nil {
		return nil, fmt.Errorf("register: failed to generate verification token: %w", err)
	}

	vt := &VerificationToken{
		UserID:    u.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := h.repo.CreateVerificationToken(ctx, vt); err != nil {
		return nil, fmt.Errorf("register: failed to store verification token: %w", err)
	}

	// Send verification email.
	verifyURL := h.cfg.AppBaseURL + "/api/users/verify?token=" + rawToken
	html := fmt.Sprintf( //nolint:gocritic // HTML template uses quoted URL, not Go quoting
		`<p>Welcome! Please verify your email by clicking the link below:</p><p><a href="%s">Verify Email</a></p><p>This link expires in 24 hours.</p>`,
		verifyURL,
	)
	if err := h.emailSender.Send(ctx, email.SendParams{
		To:      u.Email,
		Subject: "Verify your email",
		HTML:    html,
	}); err != nil {
		return nil, fmt.Errorf("register: failed to send verification email: %w", err)
	}

	return &RegisterResult{User: u, RawToken: rawToken}, nil
}

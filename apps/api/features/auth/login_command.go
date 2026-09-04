package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"arsen/features/user"
	"arsen/pkg/config"
	"arsen/pkg/cqrs"
	"arsen/pkg/jwt"
	"arsen/pkg/token"
)

// dummyHash is a pre-computed bcrypt hash used for constant-time comparison
// when the user is not found, preventing timing-based user enumeration.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), 12)

type LoginCommand struct {
	Email    string
	Password string
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int // seconds
}

func (c LoginCommand) Validate() error {
	var fields []cqrs.FieldError
	if c.Email == "" {
		fields = append(fields, cqrs.FieldError{Field: "email", Detail: "is required"})
	}
	if c.Password == "" {
		fields = append(fields, cqrs.FieldError{Field: "password", Detail: "is required"})
	}
	if len(fields) > 0 {
		return cqrs.NewValidationError(fields)
	}
	return nil
}

type LoginCommandHandler struct {
	userByEmail *cqrs.QueryBus[user.GetUserByEmailQuery, *user.GetUserByEmailResult]
	authRepo    Repository
	jwtService  *jwt.Service
	cfg         *config.Config
}

func NewLoginCommandHandler(
	userByEmail *cqrs.QueryBus[user.GetUserByEmailQuery, *user.GetUserByEmailResult],
	authRepo Repository,
	jwtService *jwt.Service,
	cfg *config.Config,
) *LoginCommandHandler {
	return &LoginCommandHandler{
		userByEmail: userByEmail,
		authRepo:    authRepo,
		jwtService:  jwtService,
		cfg:         cfg,
	}
}

func (h *LoginCommandHandler) Handle(ctx context.Context, cmd LoginCommand) (*LoginResult, error) {
	// Look up user by email via the user feature's query bus.
	result, err := h.userByEmail.Ask(ctx, user.GetUserByEmailQuery{Email: cmd.Email})

	var u *user.User
	if result != nil {
		u = result.User
	}

	var notFound *cqrs.NotFoundError
	if errors.As(err, &notFound) {
		// User not found: run bcrypt compare against dummy hash for constant-time behaviour.
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(cmd.Password))
		return nil, cqrs.NewUnauthorizedError("invalid email or password")
	}
	if err != nil {
		return nil, fmt.Errorf("login: failed to look up user: %w", err)
	}

	// Verify password.
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(cmd.Password)); err != nil {
		return nil, cqrs.NewUnauthorizedError("invalid email or password")
	}

	// Ensure email is verified.
	if !u.EmailVerified {
		return nil, cqrs.NewForbiddenError("you must verify your email address before logging in")
	}

	// Generate access token (JWT).
	accessToken, err := h.jwtService.CreateToken(u.ID, h.cfg.JWT.AccessTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("login: failed to create access token: %w", err)
	}

	// Generate refresh token (opaque).
	rawToken, hash, err := token.Generate()
	if err != nil {
		return nil, fmt.Errorf("login: failed to generate refresh token: %w", err)
	}

	// Generate family ID (random v4 UUID).
	familyID, err := generateUUID()
	if err != nil {
		return nil, fmt.Errorf("login: failed to generate family ID: %w", err)
	}

	// Persist refresh token.
	rt := &RefreshToken{
		UserID:    u.ID,
		TokenHash: hash,
		FamilyID:  familyID,
		ExpiresAt: time.Now().Add(h.cfg.JWT.RefreshTokenDuration),
	}
	if err := h.authRepo.CreateRefreshToken(ctx, rt); err != nil {
		return nil, fmt.Errorf("login: failed to store refresh token: %w", err)
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: rawToken,
		ExpiresIn:    int(h.cfg.JWT.AccessTokenDuration.Seconds()),
	}, nil
}

// generateUUID generates a random v4 UUID string from crypto/rand.
func generateUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// Set version 4 (bits 12-15 of time_hi_and_version).
	b[6] = (b[6] & 0x0f) | 0x40
	// Set variant bits (bits 6-7 of clock_seq_hi_and_reserved).
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

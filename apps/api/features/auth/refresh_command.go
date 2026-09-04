package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"arsen/pkg/config"
	"arsen/pkg/cqrs"
	"arsen/pkg/jwt"
	"arsen/pkg/token"
)

type RefreshTokenCommand struct {
	Token string // raw refresh token from client
}

type RefreshTokenResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

func (c RefreshTokenCommand) Validate() error {
	if c.Token == "" {
		return cqrs.NewValidationError([]cqrs.FieldError{
			{Field: "token", Detail: "is required"},
		})
	}
	return nil
}

type RefreshTokenCommandHandler struct {
	repo       Repository
	jwtService *jwt.Service
	cfg        *config.Config
}

func NewRefreshTokenCommandHandler(
	repo Repository,
	jwtService *jwt.Service,
	cfg *config.Config,
) *RefreshTokenCommandHandler {
	return &RefreshTokenCommandHandler{
		repo:       repo,
		jwtService: jwtService,
		cfg:        cfg,
	}
}

func (h *RefreshTokenCommandHandler) Handle(ctx context.Context, cmd RefreshTokenCommand) (*RefreshTokenResult, error) {
	// Hash the submitted token.
	hash, err := token.Hash(cmd.Token)
	if err != nil {
		return nil, cqrs.NewUnauthorizedError("invalid or expired refresh token")
	}

	// Look up the refresh token by hash (only non-revoked tokens).
	rt, err := h.repo.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		var notFound *cqrs.NotFoundError
		if errors.As(err, &notFound) {
			return nil, cqrs.NewUnauthorizedError("invalid or expired refresh token")
		}
		return nil, fmt.Errorf("refresh: failed to look up refresh token: %w", err)
	}

	// Check if the token has expired.
	if rt.ExpiresAt.Before(time.Now()) {
		// Revoke the entire family and return unauthorized.
		_ = h.repo.RevokeRefreshTokenFamily(ctx, rt.FamilyID)
		return nil, cqrs.NewUnauthorizedError("invalid or expired refresh token")
	}

	// Rotate: revoke the old token's entire family.
	if err := h.repo.RevokeRefreshTokenFamily(ctx, rt.FamilyID); err != nil {
		return nil, fmt.Errorf("refresh: failed to revoke token family: %w", err)
	}

	// Generate a new refresh token.
	rawNewToken, newHash, err := token.Generate()
	if err != nil {
		return nil, fmt.Errorf("refresh: failed to generate refresh token: %w", err)
	}

	// Store the new refresh token with the same family ID.
	newRT := &RefreshToken{
		UserID:    rt.UserID,
		TokenHash: newHash,
		FamilyID:  rt.FamilyID,
		ExpiresAt: time.Now().Add(h.cfg.JWT.RefreshTokenDuration),
	}
	if err := h.repo.CreateRefreshToken(ctx, newRT); err != nil {
		return nil, fmt.Errorf("refresh: failed to store refresh token: %w", err)
	}

	// Generate a new access token.
	accessToken, err := h.jwtService.CreateToken(rt.UserID, h.cfg.JWT.AccessTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("refresh: failed to create access token: %w", err)
	}

	return &RefreshTokenResult{
		AccessToken:  accessToken,
		RefreshToken: rawNewToken,
		ExpiresIn:    int(h.cfg.JWT.AccessTokenDuration.Seconds()),
	}, nil
}

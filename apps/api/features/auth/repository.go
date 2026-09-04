package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"arsen/pkg/cqrs"
)

type RefreshToken struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	TokenHash []byte    `db:"token_hash"`
	FamilyID  string    `db:"family_id"`
	ExpiresAt time.Time `db:"expires_at"`
	Revoked   bool      `db:"revoked"`
	CreatedAt time.Time `db:"created_at"`
}

type Repository interface {
	CreateRefreshToken(ctx context.Context, token *RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, hash []byte) (*RefreshToken, error)
	RevokeRefreshTokenFamily(ctx context.Context, familyID string) error
	RevokeAllUserRefreshTokens(ctx context.Context, userID string) error
}

type SQLRepository struct {
	db *sqlx.DB
}

func NewSQLRepository(db *sqlx.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	return r.db.QueryRowxContext(
		ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, family_id, expires_at)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		token.UserID, token.TokenHash, token.FamilyID, token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt)
}

func (r *SQLRepository) GetRefreshTokenByHash(ctx context.Context, hash []byte) (*RefreshToken, error) {
	var rt RefreshToken
	if err := r.db.GetContext(ctx, &rt, `SELECT * FROM refresh_tokens WHERE token_hash = $1 AND revoked = false`, hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, cqrs.NewNotFoundError("RefreshToken", "hash")
		}
		return nil, err
	}
	return &rt, nil
}

func (r *SQLRepository) RevokeRefreshTokenFamily(ctx context.Context, familyID string) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE refresh_tokens SET revoked = true WHERE family_id = $1 AND revoked = false`,
		familyID,
	)
	return err
}

func (r *SQLRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE refresh_tokens SET revoked = true WHERE user_id = $1 AND revoked = false`,
		userID,
	)
	return err
}

package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"arsen/pkg/cqrs"
)

type User struct {
	ID            string    `db:"id"`
	Email         string    `db:"email"`
	PasswordHash  string    `db:"password_hash"`
	EmailVerified bool      `db:"email_verified"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type VerificationToken struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	TokenHash []byte     `db:"token_hash"`
	ExpiresAt time.Time  `db:"expires_at"`
	UsedAt    *time.Time `db:"used_at"`
	CreatedAt time.Time  `db:"created_at"`
}

type PasswordResetToken struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	TokenHash []byte     `db:"token_hash"`
	ExpiresAt time.Time  `db:"expires_at"`
	UsedAt    *time.Time `db:"used_at"`
	CreatedAt time.Time  `db:"created_at"`
}

type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	UpdateEmailVerified(ctx context.Context, id string, verified bool) error
	UpdatePasswordHash(ctx context.Context, id string, hash string) error
	CreateVerificationToken(ctx context.Context, token *VerificationToken) error
	GetVerificationTokenByHash(ctx context.Context, hash []byte) (*VerificationToken, error)
	InvalidateUserVerificationTokens(ctx context.Context, userID string) error
	CreatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error
	GetPasswordResetTokenByHash(ctx context.Context, hash []byte) (*PasswordResetToken, error)
	InvalidateUserPasswordResetTokens(ctx context.Context, userID string) error
}

// RefreshTokenRevoker abstracts revoking refresh tokens without importing the auth package.
type RefreshTokenRevoker interface {
	RevokeAllUserRefreshTokens(ctx context.Context, userID string) error
}

type sqlRefreshTokenRevoker struct {
	db *sqlx.DB
}

// NewSQLRefreshTokenRevoker returns a RefreshTokenRevoker backed by sqlx.
func NewSQLRefreshTokenRevoker(db *sqlx.DB) RefreshTokenRevoker {
	return &sqlRefreshTokenRevoker{db: db}
}

func (r *sqlRefreshTokenRevoker) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1 AND revoked = false`, userID)
	return err
}

type SQLRepository struct {
	db *sqlx.DB
}

func NewSQLRepository(db *sqlx.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(ctx context.Context, user *User) error {
	return r.db.QueryRowxContext(
		ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, created_at, updated_at`,
		user.Email, user.PasswordHash,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *SQLRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	if err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE email = $1`, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, cqrs.NewNotFoundError("User", email)
		}
		return nil, err
	}
	return &u, nil
}

func (r *SQLRepository) GetByID(ctx context.Context, id string) (*User, error) {
	var u User
	if err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE id = $1`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, cqrs.NewNotFoundError("User", id)
		}
		return nil, err
	}
	return &u, nil
}

func (r *SQLRepository) UpdateEmailVerified(ctx context.Context, id string, verified bool) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE users SET email_verified = $2, updated_at = now() WHERE id = $1`,
		id, verified,
	)
	return err
}

func (r *SQLRepository) UpdatePasswordHash(ctx context.Context, id, hash string) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE users SET password_hash = $2, updated_at = now() WHERE id = $1`,
		id, hash,
	)
	return err
}

func (r *SQLRepository) CreateVerificationToken(ctx context.Context, token *VerificationToken) error {
	return r.db.QueryRowxContext(
		ctx,
		`INSERT INTO verification_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3) RETURNING id, created_at`,
		token.UserID, token.TokenHash, token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt)
}

func (r *SQLRepository) GetVerificationTokenByHash(ctx context.Context, hash []byte) (*VerificationToken, error) {
	var vt VerificationToken
	if err := r.db.GetContext(ctx, &vt, `SELECT * FROM verification_tokens WHERE token_hash = $1 AND used_at IS NULL`, hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, cqrs.NewNotFoundError("VerificationToken", "hash")
		}
		return nil, err
	}
	return &vt, nil
}

func (r *SQLRepository) InvalidateUserVerificationTokens(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE verification_tokens SET used_at = now() WHERE user_id = $1 AND used_at IS NULL`,
		userID,
	)
	return err
}

func (r *SQLRepository) CreatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error {
	return r.db.QueryRowxContext(
		ctx,
		`INSERT INTO password_reset_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3) RETURNING id, created_at`,
		token.UserID, token.TokenHash, token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt)
}

func (r *SQLRepository) GetPasswordResetTokenByHash(ctx context.Context, hash []byte) (*PasswordResetToken, error) {
	var prt PasswordResetToken
	if err := r.db.GetContext(ctx, &prt, `SELECT * FROM password_reset_tokens WHERE token_hash = $1 AND used_at IS NULL`, hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, cqrs.NewNotFoundError("PasswordResetToken", "hash")
		}
		return nil, err
	}
	return &prt, nil
}

func (r *SQLRepository) InvalidateUserPasswordResetTokens(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE password_reset_tokens SET used_at = now() WHERE user_id = $1 AND used_at IS NULL`,
		userID,
	)
	return err
}

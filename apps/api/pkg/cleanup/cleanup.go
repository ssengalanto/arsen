package cleanup

import (
	"context"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
)

type Service struct {
	db       *sqlx.DB
	interval time.Duration
}

func NewService(db *sqlx.DB, interval time.Duration) *Service {
	return &Service{db: db, interval: interval}
}

func (s *Service) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run once on startup.
	s.run(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("token cleanup stopped")
			return
		case <-ticker.C:
			s.run(ctx)
		}
	}
}

func (s *Service) run(ctx context.Context) {
	total := int64(0)

	n, err := s.deleteExpiredRefreshTokens(ctx)
	if err != nil {
		slog.Error("cleanup: refresh tokens", "error", err)
	} else {
		total += n
	}

	n, err = s.deleteExpiredVerificationTokens(ctx)
	if err != nil {
		slog.Error("cleanup: verification tokens", "error", err)
	} else {
		total += n
	}

	n, err = s.deleteExpiredPasswordResetTokens(ctx)
	if err != nil {
		slog.Error("cleanup: password reset tokens", "error", err)
	} else {
		total += n
	}

	if total > 0 {
		slog.Info("token cleanup completed", "deleted", total)
	}
}

func (s *Service) deleteExpiredRefreshTokens(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE revoked = true OR expires_at < now()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *Service) deleteExpiredVerificationTokens(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx, `DELETE FROM verification_tokens WHERE used_at IS NOT NULL OR expires_at < now()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *Service) deleteExpiredPasswordResetTokens(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx, `DELETE FROM password_reset_tokens WHERE used_at IS NOT NULL OR expires_at < now()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

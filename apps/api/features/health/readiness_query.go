package health

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type ReadinessQuery struct{}

func (q ReadinessQuery) Validate() error { return nil }

type ReadinessResult struct {
	Ready    bool
	Database string // "reachable" or "unreachable"
}

type ReadinessQueryHandler struct {
	db *sqlx.DB
}

func NewReadinessQueryHandler(db *sqlx.DB) *ReadinessQueryHandler {
	return &ReadinessQueryHandler{db: db}
}

func (h *ReadinessQueryHandler) Handle(ctx context.Context, _ ReadinessQuery) (*ReadinessResult, error) {
	if err := h.db.PingContext(ctx); err != nil {
		return &ReadinessResult{Ready: false, Database: "unreachable"}, nil //nolint:nilerr // degraded readiness, not a handler error
	}
	return &ReadinessResult{Ready: true, Database: "reachable"}, nil
}

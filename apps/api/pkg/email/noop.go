package email

import (
	"context"
	"log/slog"
)

type NoopSender struct {
	logger *slog.Logger
}

func NewNoopSender(logger *slog.Logger) *NoopSender {
	return &NoopSender{logger: logger}
}

func (s *NoopSender) Send(_ context.Context, params SendParams) error {
	s.logger.Info("noop email send", "to", params.To, "subject", params.Subject)
	return nil
}

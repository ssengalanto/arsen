package email

import (
	"log/slog"
	"strings"

	"arsen/pkg/config"

	"go.uber.org/fx"
)

var Module = fx.Module("email",
	fx.Provide(func(cfg *config.Config, logger *slog.Logger) Sender {
		if isProduction(cfg.Env) {
			return NewResendSender(cfg.Email.ResendAPIKey, cfg.Email.FromAddress, cfg.Email.FromName)
		}
		return NewNoopSender(logger)
	}),
)

func isProduction(env string) bool {
	e := strings.ToLower(env)
	return e == "prod" || e == "production"
}

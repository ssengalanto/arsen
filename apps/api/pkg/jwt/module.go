package jwt

import (
	"arsen/pkg/config"

	"go.uber.org/fx"
)

// Module provides the JWT service to the fx dependency graph.
var Module = fx.Module("jwt",
	fx.Provide(func(cfg *config.Config) *Service {
		return NewService(cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.Audience)
	}),
)

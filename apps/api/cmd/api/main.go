package main

import (
	"log/slog"

	"go.uber.org/fx"

	"arsen/features/auth"
	"arsen/features/health"
	"arsen/features/user"
	"arsen/pkg/cleanup"
	"arsen/pkg/config"
	"arsen/pkg/database"
	"arsen/pkg/email"
	"arsen/pkg/jwt"
	"arsen/pkg/redis"
	"arsen/pkg/server"

	_ "arsen/docs/swagger" // swagger docs
)

// @title Arsen API
// @version 1.0
// @description Go backend boilerplate with JWT authentication, CQRS, and vertical slice architecture.

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer {token}" (with quotes removed)

// @tag.name auth
// @tag.description Authentication endpoints (login, token refresh, logout)
// @tag.name users
// @tag.description User management endpoints (register, verify, profile, password reset)
// @tag.name health
// @tag.description Health and readiness check endpoints
func main() {
	fx.New(
		fx.Provide(slog.Default),
		config.Module,
		database.Module,
		redis.Module,
		email.Module,
		jwt.Module,
		server.Module,
		user.Module,
		auth.Module,
		health.Module,
		cleanup.Module,
	).Run()
}

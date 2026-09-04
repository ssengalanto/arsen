package main

import (
	"go.uber.org/fx"

	"arsen/features/auth"
	"arsen/features/user"
	"arsen/pkg/config"
	"arsen/pkg/database"
	"arsen/pkg/email"
	"arsen/pkg/jwt"
	"arsen/pkg/redis"
	"arsen/pkg/server"
)

func main() {
	fx.New(
		config.Module,
		database.Module,
		redis.Module,
		email.Module,
		jwt.Module,
		server.Module,
		user.Module,
		auth.Module,
	).Run()
}

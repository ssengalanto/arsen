package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Env        string
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	JWT        JWTConfig
	Email      EmailConfig
	RateLimit  RateLimitConfig
	Cleanup    CleanupConfig
	AppBaseURL string
}

type ServerConfig struct {
	Address string
}

type DatabaseConfig struct {
	URL string
}

type RedisConfig struct {
	URL string
}

type JWTConfig struct {
	Secret               string
	Issuer               string
	Audience             string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

type EmailConfig struct {
	ResendAPIKey string
	FromAddress  string
	FromName     string
}

type RateLimitConfig struct {
	Requests int
	Burst    int
}

type CleanupConfig struct {
	Interval time.Duration
}

func New() (*Config, error) {
	viper.AutomaticEnv()

	viper.SetDefault("ENV", "dev")
	viper.SetDefault("SERVER_ADDRESS", ":8080")
	viper.SetDefault("ACCESS_TOKEN_DURATION", "15m")
	viper.SetDefault("REFRESH_TOKEN_DURATION", "168h")
	viper.SetDefault("RATE_LIMIT_REQUESTS", 10)
	viper.SetDefault("RATE_LIMIT_BURST", 20)
	viper.SetDefault("CLEANUP_INTERVAL", "1h")
	viper.SetDefault("LOG_LEVEL", "debug")

	env := viper.GetString("ENV")

	cfg := &Config{
		Env: env,
		Server: ServerConfig{
			Address: viper.GetString("SERVER_ADDRESS"),
		},
		Database: DatabaseConfig{
			URL: viper.GetString("DATABASE_URL"),
		},
		Redis: RedisConfig{
			URL: viper.GetString("REDIS_URL"),
		},
		JWT: JWTConfig{
			Secret:               viper.GetString("JWT_SECRET"),
			Issuer:               viper.GetString("JWT_ISSUER"),
			Audience:             viper.GetString("JWT_AUDIENCE"),
			AccessTokenDuration:  viper.GetDuration("ACCESS_TOKEN_DURATION"),
			RefreshTokenDuration: viper.GetDuration("REFRESH_TOKEN_DURATION"),
		},
		Email: EmailConfig{
			ResendAPIKey: viper.GetString("RESEND_API_KEY"),
			FromAddress:  viper.GetString("EMAIL_FROM_ADDRESS"),
			FromName:     viper.GetString("EMAIL_FROM_NAME"),
		},
		RateLimit: RateLimitConfig{
			Requests: viper.GetInt("RATE_LIMIT_REQUESTS"),
			Burst:    viper.GetInt("RATE_LIMIT_BURST"),
		},
		Cleanup: CleanupConfig{
			Interval: viper.GetDuration("CLEANUP_INTERVAL"),
		},
		AppBaseURL: viper.GetString("APP_BASE_URL"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.Redis.URL == "" {
		return fmt.Errorf("REDIS_URL is required")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.Env == "prod" && c.Email.ResendAPIKey == "" {
		return fmt.Errorf("RESEND_API_KEY is required in production")
	}
	return nil
}

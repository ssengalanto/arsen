package database

import (
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"arsen/pkg/config"
)

func New(cfg *config.Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("opening database connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return db, nil
}

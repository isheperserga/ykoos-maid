package database

import (
	"fmt"

	"yk-dc-bot/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Database struct {
	*sqlx.DB
}

func NewPostgresDB(cfg *config.Config) (*Database, error) {
	dbURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Name,
	)

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	if err := RunMigrations(db); err != nil {
		return nil, fmt.Errorf("error running migrations: %w", err)
	}

	return &Database{db}, nil
}

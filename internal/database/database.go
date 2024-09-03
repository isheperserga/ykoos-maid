package database

import (
	"fmt"

	"yk-dc-bot/internal/apperrors"
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
		return nil, apperrors.Wrap(err, "DB_CONNECTION_ERROR", "error connecting to database")
	}

	if err := RunMigrations(db); err != nil {
		return nil, apperrors.Wrap(err, "DB_MIGRATION_ERROR", "error running migrations")
	}

	return &Database{db}, nil
}

func (db *Database) QueryIter(query string, args ...interface{}) func(yield func(map[string]interface{}, error) bool) {
	return func(yield func(map[string]interface{}, error) bool) {
		rows, err := db.Queryx(query, args...)
		if err != nil {
			yield(nil, err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			result := make(map[string]interface{})
			err := rows.MapScan(result)
			if !yield(result, err) {
				return
			}
		}
	}
}

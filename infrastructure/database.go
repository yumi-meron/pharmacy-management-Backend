package infrastructure

import (
	"database/sql"
	"pharmacy-management-backend/config"

	_ "github.com/lib/pq" // Postgres driver
	"github.com/rs/zerolog"
)

// NewDatabase initializes a Postgres database connection
func NewDatabase(cfg *config.Config, logger zerolog.Logger) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to database")
		return nil, err
	}

	if err = db.Ping(); err != nil {
		logger.Error().Err(err).Msg("Database ping failed")
		return nil, err
	}

	logger.Info().Msg("Database connected successfully")
	return db, nil
}

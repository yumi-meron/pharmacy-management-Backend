package infrastructure

import (
	"database/sql"

	"pharmacy-management-backend/config"

	"github.com/rs/zerolog"
)

// NewDatabase initializes a new database connection
func NewDatabase(cfg *config.Config, logger zerolog.Logger) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to database")
		return nil, err
	}

	if err := db.Ping(); err != nil {
		logger.Error().Err(err).Msg("Failed to ping database")
		return nil, err
	}

	logger.Info().Msg("Database connection established")
	return db, nil
}

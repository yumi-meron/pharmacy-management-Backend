package infrastructure

import (
	"database/sql"
	"time"

	"pharmacist-backend/config"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

// NewDatabase initializes a PostgreSQL database connection
func NewDatabase(cfg *config.Config, logger zerolog.Logger) (*sql.DB, error) {
	// Open database connection
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to open database connection")
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		logger.Error().Err(err).Msg("Failed to ping database")
		db.Close()
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	logger.Info().Msg("Database connection established")
	return db, nil
}

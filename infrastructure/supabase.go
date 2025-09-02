package infrastructure

import (
	"pharmacy-management-backend/config"

	"github.com/rs/zerolog"
	"github.com/supabase-community/supabase-go"
)

// NewSupabase initializes a new Supabase client
func NewSupabase(cfg *config.Config, logger zerolog.Logger) (*supabase.Client, error) {
	client, err := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseKey, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to initialize Supabase client")
		return nil, err
	}
	logger.Info().Msg("Supabase client initialized successfully")
	return client, nil
}

package config

import (
    "errors"
    "os"

    "github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
    DatabaseURL          string
    JWTSecret           string
    RefreshTokenSecret  string
    TwilioAccountSID    string
    TwilioAuthToken     string
    TwilioPhoneNumber   string
    TwilioMockMode      bool
    Port                string
}

// Load loads environment variables into a Config struct
func Load() (*Config, error) {
    // Load .env file
    if err := godotenv.Load(); err != nil {
        return nil, err
    }

    cfg := &Config{
        DatabaseURL:         os.Getenv("DATABASE_URL"),
        JWTSecret:          os.Getenv("JWT_SECRET"),
        RefreshTokenSecret: os.Getenv("REFRESH_TOKEN_SECRET"),
        TwilioAccountSID:   os.Getenv("TWILIO_ACCOUNT_SID"),
        TwilioAuthToken:    os.Getenv("TWILIO_AUTH_TOKEN"),
        TwilioPhoneNumber:  os.Getenv("TWILIO_PHONE_NUMBER"),
        TwilioMockMode:     os.Getenv("TWILIO_MOCK_MODE") == "true",
        Port:               os.Getenv("PORT"),
    }

    // Set default port if not specified
    if cfg.Port == "" {
        cfg.Port = "8080"
    }

    // Validate required fields
    if cfg.DatabaseURL == "" {
        return nil, errors.New("DATABASE_URL is required")
    }
    if cfg.JWTSecret == "" {
        return nil, errors.New("JWT_SECRET is required")
    }
    if cfg.RefreshTokenSecret == "" {
        return nil, errors.New("REFRESH_TOKEN_SECRET is required")
    }
    if !cfg.TwilioMockMode {
        if cfg.TwilioAccountSID == "" || cfg.TwilioAuthToken == "" || cfg.TwilioPhoneNumber == "" {
            return nil, errors.New("Twilio configuration is required when TWILIO_MOCK_MODE is false")
        }
    }

    return cfg, nil
}
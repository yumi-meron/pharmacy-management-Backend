package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	TwilioSID   string
	TwilioToken string
	TwilioFrom  string
	MockTwilio  bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	_ = godotenv.Load() // Load .env file if present
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/pharmacist?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "your_jwt_secret"),
		TwilioSID:   getEnv("TWILIO_SID", ""),
		TwilioToken: getEnv("TWILIO_TOKEN", ""),
		TwilioFrom:  getEnv("TWILIO_FROM", ""),
		MockTwilio:  getEnvBool("TWILIO_MOCK", false),
	}
	return cfg, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvBool retrieves an environment variable as a boolean
func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		return value == "true"
	}
	return defaultValue
}

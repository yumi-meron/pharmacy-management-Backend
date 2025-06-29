package config

import (
	"github.com/joho/godotenv"
)

// LoadConfig loads environment variables
func LoadConfig() error {
	return godotenv.Load()
}

package infrastructure

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"
)

// NewDB initializes a new database connection
func NewDB() (*sql.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	return sql.Open("postgres", dbURL)
}

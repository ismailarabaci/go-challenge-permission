package server

import (
	"os"
	"time"
)

// Config holds the configuration for the server and database connection
type Config struct {
	// DatabaseDSN is the data source name for the database connection
	// Example: "user:password@tcp(host:port)/dbname"
	DatabaseDSN string

	// MaxOpenConns sets the maximum number of open connections to the database
	MaxOpenConns int

	// MaxIdleConns sets the maximum number of connections in the idle connection pool
	MaxIdleConns int

	// ConnMaxLifetime sets the maximum amount of time a connection may be reused
	ConnMaxLifetime time.Duration
}

// DefaultConfig returns a Config with sensible defaults
// Reads DatabaseDSN from MYSQL_DSN environment variable if set
func DefaultConfig() Config {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "blp:password@tcp(localhost:3306)/blp-coding-challenge"
	}

	return Config{
		DatabaseDSN:     dsn,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}
}

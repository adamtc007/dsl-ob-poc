package config

import (
	"os"

	"dsl-ob-poc/internal/datastore"
)

// GetDataStoreConfig returns the data store configuration based on environment variables and flags
func GetDataStoreConfig() datastore.Config {
	// Always use PostgreSQL - mock mode removed from production code
	config := datastore.Config{
		Type:             datastore.PostgreSQLStore,
		ConnectionString: getConnectionString(),
	}

	return config
}

// getConnectionString returns the database connection string
func getConnectionString() string {
	connStr := os.Getenv("DB_CONN_STRING")
	if connStr == "" {
		// Default connection string for local development
		return "postgres://localhost:5432/postgres?sslmode=disable"
	}
	return connStr
}

// IsMockMode returns false - mock mode has been removed from production code
// This function is kept for backward compatibility but always returns false
func IsMockMode() bool {
	return false
}

package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	// Read APP_ENV; default to "dev" if not set.
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "dev"
	}

	// Retrieve common database configuration.
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	var connStr string
	// Use different connection string formats depending on the environment.
	if appEnv == "dev" {
		// In development, you might disable SSL.
		connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=verify-full sslrootcert=rds-ca.pem",
			host, port, user, password, dbname)
	} else {
		// In production/UAT, enforce SSL (adjust as necessary).
		connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=verify-full sslrootcert=rds-ca.pem",
			host, port, user, password, dbname)
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}

	fmt.Println("Connected to Postgres")
}

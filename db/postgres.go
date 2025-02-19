package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq" // Postgres driver
)

// DB is a global variable that holds the database connection pool.
var DB *sql.DB

// Connect reads environment variables, creates a connection string,
// and connects to the Postgres database.
func Connect() {
	// Retrieve database credentials from environment variables.
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	// Using DB_INSTANCE_IDENTIFIER as the database name.
	dbName := os.Getenv("DB_INSTANCE_IDENTIFIER")
	engine := os.Getenv("DB_ENGINE")

	// Check if we are using the Postgres engine.
	if engine != "postgres" {
		log.Fatalf("Unsupported database engine: %s", engine)
	}

	// Build the connection string with SSL enabled.
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		host, port, username, password, dbName)

	// Open the connection.
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// Verify the connection is successful.
	if err = DB.Ping(); err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	log.Println("Successfully connected to the Postgres database")
}

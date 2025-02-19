package main

import (
	"auth-service/db"
	"auth-service/routes"
	"auth-service/secretmanager" // Ensure this is available in production.
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

// loadProdSecrets fetches secrets from the secret manager in production,
// parses the JSON, and sets the corresponding environment variables.
func loadProdSecrets() {
	// --- Load JWT Secret ---
	jwtSecretJSON, err := secretmanager.GetSecret("prod/jwt")
	if err != nil {
		log.Fatalf("Error retrieving JWT secret: %v", err)
	}
	var jwtSecrets map[string]string
	if err := json.Unmarshal([]byte(jwtSecretJSON), &jwtSecrets); err != nil {
		log.Fatalf("Error parsing JWT secret JSON: %v", err)
	}
	// Set each key/value from the JWT secret.
	for key, value := range jwtSecrets {
		os.Setenv(key, value)
	}

	// --- Load Postgres Credentials ---
	pgSecretJSON, err := secretmanager.GetSecret("prod/postgres")
	if err != nil {
		log.Fatalf("Error retrieving Postgres secret: %v", err)
	}
	var pgSecrets map[string]interface{}
	if err := json.Unmarshal([]byte(pgSecretJSON), &pgSecrets); err != nil {
		log.Fatalf("Error parsing Postgres secret JSON: %v", err)
	}
	// Map the secret keys to environment variables for the database connection.
	os.Setenv("DB_USERNAME", pgSecrets["username"].(string))
	os.Setenv("DB_PASSWORD", pgSecrets["password"].(string))
	os.Setenv("DB_ENGINE", pgSecrets["engine"].(string))
	os.Setenv("DB_HOST", pgSecrets["host"].(string))
	// Convert port (which might be a number) to a string.
	portStr := fmt.Sprintf("%v", pgSecrets["port"])
	os.Setenv("DB_PORT", portStr)
	os.Setenv("DB_INSTANCE_IDENTIFIER", pgSecrets["dbInstanceIdentifier"].(string))
	log.Printf("DB_USERNAME_TEST_SECRET")
	log.Printf("%s", pgSecrets["username"].(string))
}

func main() {
	// Always attempt to load the .env file.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; using system environment variables")
	}
	log.Println(os.Getenv("DB_USERNAME"))

	// Read the APP_ENV environment variable; default to "dev" if not set.
	appEnv := os.Getenv("APP_ENV")

	log.Println("Environment:", appEnv)

	// In production, retrieve secrets from the secret manager.
	if appEnv == "prod" {
		loadProdSecrets()
	}

	// Connect to the database.
	db.Connect()

	// Setup routes.
	router := routes.SetupRoutes()

	// Set the server port (default to 8080 if not provided).
	port := os.Getenv("APP_PORT")

	log.Printf("Starting server on port %s in %s environment", port, appEnv)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

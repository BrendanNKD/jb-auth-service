package main

import (
	"auth-service/db"
	"auth-service/routes"
	"auth-service/secretmanager" // Your JWT package that now uses environment variables as needed.
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Always attempt to load the .env file.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; using system environment variables")
	}

	// Read the APP_ENV environment variable; default to "dev" if not set.
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "dev"
	}
	log.Println("Environment:", appEnv)

	// In production/UAT, load secrets from AWS Secrets Manager.
	// For example, assume you have two secrets:
	// - One for JWT config (with key/value pairs like JWT_SECRET, JWT_EXPIRE_HOURS).
	// - One for PostgreSQL config (with key/value pairs like POSTGRES_USER, POSTGRES_PASSWORD, etc).
	region := os.Getenv("AWS_REGION")
	if appEnv == "uat" || appEnv == "prod" {
		jwtSecretName := os.Getenv("AWS_JWT_SECRET_NAME")
		if jwtSecretName == "" {
			log.Fatal("AWS_JWT_SECRET_NAME environment variable is not set")
		}
		postgresSecretName := os.Getenv("AWS_POSTGRES_SECRET_NAME")
		if postgresSecretName == "" {
			log.Fatal("AWS_POSTGRES_SECRET_NAME environment variable is not set")
		}
		if region == "" {
			log.Fatal("AWS_REGION environment variable is not set")
		}

		// Load and cache the JWT secret key/value pairs into environment variables.
		if err := secretmanager.LoadSecretToEnv(jwtSecretName, region); err != nil {
			log.Fatalf("Failed to load JWT secret: %v", err)
		}
		log.Println("Loaded JWT secret from AWS Secrets Manager.")

		// Load and cache the PostgreSQL secret key/value pairs into environment variables.
		if err := secretmanager.LoadSecretToEnv(postgresSecretName, region); err != nil {
			log.Fatalf("Failed to load PostgreSQL secret: %v", err)
		}
		log.Println("Loaded PostgreSQL secret from AWS Secrets Manager.")
	} else if appEnv == "dev" {
		log.Println("Running in development mode; using local environment variables.")
	} else {
		log.Fatalf("Unknown environment: %s", appEnv)
	}

	// Connect to the database.
	db.Connect()

	// Setup routes.
	router := routes.SetupRoutes()

	// Set the server port (default to 8080 if not provided).
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s in %s environment", port, appEnv)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

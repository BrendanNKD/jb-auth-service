package main

import (
	"auth-service/db"
	"auth-service/routes"

	//"auth-service/secretmanager" // Your JWT package that now uses environment variables as needed.
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

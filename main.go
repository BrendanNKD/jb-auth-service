package main

import (
	"auth-service/db"
	"auth-service/routes"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Attempt to load environment variables from .env file.
	// If the file doesn't exist, log a warning and continue.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; using environment variables")
	}

	// Connect to the database.
	db.Connect()

	// Setup and start the server.
	router := routes.SetupRoutes()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

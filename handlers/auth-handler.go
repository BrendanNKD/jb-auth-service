// handlers/auth.go
package handlers

import (
	"auth-service/db"
	"auth-service/models"
	jwt "auth-service/utils"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type JSONResponse map[string]interface{}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate input for username, password, and role
	if user.Username == "" || user.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}
	if user.Role == "" {
		http.Error(w, "Role is required", http.StatusBadRequest)
		return
	}
	if user.Role != "employer" && user.Role != "job_seeker" {
		http.Error(w, "Invalid role. Must be either 'employer' or 'job_seeker'", http.StatusBadRequest)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Insert the new user into the database.
	_, err = db.DB.Exec("INSERT INTO users (username, password, role) VALUES ($1, $2, $3)",
		user.Username, string(hashedPassword), user.Role)
	if err != nil {
		log.Printf("Error inserting user into database: %v", err)
		http.Error(w, "User already exists or database error", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(JSONResponse{"message": "User registered successfully"})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate input
	if user.Username == "" || user.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Retrieve the user's password and role from the database
	var storedPassword, role string
	err := db.DB.QueryRow("SELECT password, role FROM users WHERE username = $1", user.Username).Scan(&storedPassword, &role)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		} else {
			log.Printf("Database error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Verify the password hash
	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password)); err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Generate JWT token using the revised GenerateToken function (with username and role)
	token, err := jwt.GenerateToken(user.Username, role)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	// Return the token
	json.NewEncoder(w).Encode(JSONResponse{"token": token})
}

func AuthenticateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":   false,
			"message": "No token provided",
		})
		return
	}

	// Unconditionally trim the "Bearer " prefix.
	authHeader = strings.TrimPrefix(authHeader, "Bearer ")

	claims, isExpired, err := jwt.ValidateToken(authHeader)
	if err != nil {
		if isExpired {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"valid":   false,
				"message": "Token has expired",
			})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":   false,
			"message": "Invalid token",
		})
		return
	}

	// Return valid response
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":    true,
		"message":  "Token is valid",
		"username": claims.Username,
		"role":     claims.Role,
	})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(JSONResponse{"message": "Logged out successfully"})
}

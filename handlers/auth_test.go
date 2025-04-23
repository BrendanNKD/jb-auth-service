package handlers_test

import (
	"auth-service/db"
	"auth-service/handlers"
	"auth-service/models"
	jwt "auth-service/utils"
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// setupMockDB creates a sqlmock DB and assigns it to the global DB.
func setupMockDB() (sqlmock.Sqlmock, func()) {
	mockDB, mock, _ := sqlmock.New()
	db.DB = mockDB
	return mock, func() { mockDB.Close() }
}

// --------------------
// RegisterHandler Tests
// --------------------

// Successful registration test.
func TestRegisterHandler(t *testing.T) {
	mock, cleanup := setupMockDB()
	defer cleanup()

	// Use a valid role "job_seeker" (or "employer")
	mock.ExpectExec("INSERT INTO users").
		WithArgs("testuser", sqlmock.AnyArg(), "job_seeker").
		WillReturnResult(sqlmock.NewResult(1, 1))

	user := models.Users{Username: "testuser", Password: "password", Role: "job_seeker"}
	body, err := json.Marshal(user)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.RegisterHandler(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

// Invalid JSON payload.
func TestRegisterHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/register", strings.NewReader("{invalid-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.RegisterHandler(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// Missing username.
func TestRegisterHandler_MissingUsername(t *testing.T) {
	user := models.Users{Password: "password"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.RegisterHandler(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// Missing password.
func TestRegisterHandler_MissingPassword(t *testing.T) {
	user := models.Users{Username: "testuser"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.RegisterHandler(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// Missing role.
// func TestRegisterHandler_MissingRole(t *testing.T) {
// 	user := models.Users{Username: "testuser", Password: "password", Role: ""}
// 	body, _ := json.Marshal(user)
// 	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
// 	req.Header.Set("Content-Type", "application/json")
// 	rec := httptest.NewRecorder()

// 	handlers.RegisterHandler(rec, req)
// 	assert.Equal(t, http.StatusBadRequest, rec.Code)
// }

// // Invalid role.
// func TestRegisterHandler_InvalidRole(t *testing.T) {
// 	user := models.Users{Username: "testuser", Password: "password", Role: "invalid"}
// 	body, _ := json.Marshal(user)
// 	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
// 	req.Header.Set("Content-Type", "application/json")
// 	rec := httptest.NewRecorder()

// 	handlers.RegisterHandler(rec, req)
// 	assert.Equal(t, http.StatusBadRequest, rec.Code)
// }

// Database error during registration.
func TestRegisterHandler_DBError(t *testing.T) {
	mock, cleanup := setupMockDB()
	defer cleanup()

	mock.ExpectExec("INSERT INTO users").
		WithArgs("testuser", sqlmock.AnyArg(), "job_seeker").
		WillReturnError(sql.ErrConnDone) // simulate connection error

	user := models.Users{Username: "testuser", Password: "password"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.RegisterHandler(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

// --------------------
// LoginHandler Tests
// --------------------

// Successful login.
func TestLoginHandler(t *testing.T) {
	mock, cleanup := setupMockDB()
	defer cleanup()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	mock.ExpectQuery(`SELECT password, role FROM users WHERE username = \$1`).
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"password", "role"}).
			AddRow(string(hashedPassword), "job_seeker"))

	user := models.Users{Username: "testuser", Password: "password"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.LoginHandler(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// Invalid JSON payload for login.
func TestLoginHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/login", strings.NewReader("{invalid-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.LoginHandler(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// Missing username in login.
func TestLoginHandler_MissingUsername(t *testing.T) {
	user := models.Users{Password: "password"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.LoginHandler(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// Missing password in login.
func TestLoginHandler_MissingPassword(t *testing.T) {
	user := models.Users{Username: "testuser"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.LoginHandler(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// User not found.
func TestLoginHandler_UserNotFound(t *testing.T) {
	mock, cleanup := setupMockDB()
	defer cleanup()

	mock.ExpectQuery(`SELECT password, role FROM users WHERE username = \$1`).
		WithArgs("testuser").
		WillReturnError(sql.ErrNoRows)

	user := models.Users{Username: "testuser", Password: "password"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.LoginHandler(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// Database error during login.
func TestLoginHandler_DBError(t *testing.T) {
	mock, cleanup := setupMockDB()
	defer cleanup()

	mock.ExpectQuery(`SELECT password, role FROM users WHERE username = \$1`).
		WithArgs("testuser").
		WillReturnError(sql.ErrConnDone)

	user := models.Users{Username: "testuser", Password: "password"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.LoginHandler(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// Wrong password.
func TestLoginHandler_WrongPassword(t *testing.T) {
	mock, cleanup := setupMockDB()
	defer cleanup()

	// Create a hash for a different password so that the comparison fails.
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("different_password"), bcrypt.DefaultCost)

	mock.ExpectQuery(`SELECT password, role FROM users WHERE username = \$1`).
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"password", "role"}).
			AddRow(string(hashedPassword), "job_seeker"))

	user := models.Users{Username: "testuser", Password: "password"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.LoginHandler(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// --------------------
// AuthenticateHandler Tests
// --------------------

// No token provided.
func TestAuthenticateHandler_NoToken(t *testing.T) {
	req := httptest.NewRequest("GET", "/authenticate", nil)
	rec := httptest.NewRecorder()

	handlers.AuthenticateHandler(rec, req)
	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, false, response["valid"])
	assert.Equal(t, "No token provided", response["message"])
}

// Invalid token provided.
func TestAuthenticateHandler_InvalidToken(t *testing.T) {
	req := httptest.NewRequest("GET", "/authenticate", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec := httptest.NewRecorder()

	handlers.AuthenticateHandler(rec, req)
	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, false, response["valid"])
	assert.Equal(t, "Invalid token", response["message"])
}

// Expired token.
func TestAuthenticateHandler_ExpiredToken(t *testing.T) {
	// Force expiration by setting JWT_EXPIRE_HOURS to -1.
	os.Setenv("JWT_SECRET", "supersecret")
	os.Setenv("JWT_EXPIRE_HOURS", "-1")
	os.Setenv("JWT_ISSUER", "test-issuer")

	token, err := jwt.GenerateToken("testuser", "employer")
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/authenticate", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handlers.AuthenticateHandler(rec, req)
	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, false, response["valid"])
	assert.Equal(t, "Token has expired", response["message"])
}

// Valid token.
func TestAuthenticateHandler_ValidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "supersecret")
	os.Setenv("JWT_EXPIRE_HOURS", "72")
	os.Setenv("JWT_ISSUER", "test-issuer")

	token, err := jwt.GenerateToken("testuser", "employer")
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/authenticate", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handlers.AuthenticateHandler(rec, req)
	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, true, response["valid"])
	assert.Equal(t, "Token is valid", response["message"])
	assert.Equal(t, "testuser", response["username"])
	assert.Equal(t, "employer", response["role"])
}

// --------------------
// LogoutHandler Test
// --------------------

func TestLogoutHandler(t *testing.T) {
	req := httptest.NewRequest("POST", "/logout", nil)
	rec := httptest.NewRecorder()

	handlers.LogoutHandler(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Logged out successfully", response["message"])
}

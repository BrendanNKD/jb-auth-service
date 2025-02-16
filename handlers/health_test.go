package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"auth-service/handlers"

	"github.com/stretchr/testify/assert"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handlers.HealthHandler(rec, req)

	// Assert status code
	assert.Equal(t, http.StatusOK, rec.Code)

	// Assert content type header
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	// Decode the JSON response
	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

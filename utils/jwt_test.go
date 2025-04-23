package jwt_test

import (
	"os"
	"testing"
	"time"

	jwt "auth-service/utils"

	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	// Set required environment variables.
	os.Setenv("JWT_SECRET", "supersecret")
	os.Setenv("JWT_EXPIRE_HOURS", "72")
	os.Setenv("JWT_ISSUER", "test-issuer")

	token, err := jwt.GenerateToken("testuser", "employer")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "supersecret")
	os.Setenv("JWT_EXPIRE_HOURS", "72")
	os.Setenv("JWT_ISSUER", "test-issuer")

	token, err := jwt.GenerateToken("testuser", "employer")
	assert.NoError(t, err)

	claims, expired, err := jwt.ValidateToken(token)
	assert.NoError(t, err)
	assert.False(t, expired)
	assert.Equal(t, "testuser", claims.Username)
}

func TestExpiredToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "supersecret")
	// Set expire hours to -1 so that the token is already expired.
	os.Setenv("JWT_EXPIRE_HOURS", "-1")
	os.Setenv("JWT_ISSUER", "test-issuer")

	token, err := jwt.GenerateToken("testuser", "employer")
	assert.NoError(t, err)

	claims, expired, err := jwt.ValidateToken(token)
	// An expired token should return an error, and the expired flag should be true.
	assert.Error(t, err)
	assert.True(t, expired)
	assert.Contains(t, err.Error(), "expired")
	// If claims are returned, verify the username.
	if claims != nil {
		assert.Equal(t, "testuser", claims.Username)
	}
}

func TestInvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "supersecret")
	// Use a string that is not a valid JWT.
	invalidToken := "not.a.valid.token"
	claims, expired, err := jwt.ValidateToken(invalidToken)
	assert.Error(t, err)
	assert.False(t, expired)
	assert.Nil(t, claims)
}

func TestInvalidJwtExpireHours(t *testing.T) {
	os.Setenv("JWT_SECRET", "supersecret")
	// Set JWT_EXPIRE_HOURS to an invalid value so that the code falls back to 72.
	os.Setenv("JWT_EXPIRE_HOURS", "notanumber")
	os.Setenv("JWT_ISSUER", "test-issuer")

	token, err := jwt.GenerateToken("testuser", "employer")
	assert.NoError(t, err)

	claims, expired, err := jwt.ValidateToken(token)
	assert.NoError(t, err)
	assert.False(t, expired)
	assert.Equal(t, "testuser", claims.Username)

	// Check that the token expiration is roughly 72 hours from now.
	expectedDuration := time.Hour * 72
	now := time.Now()
	diff := claims.ExpiresAt.Time.Sub(now)
	// Allow a small margin for processing delay (e.g., 5 seconds).
	assert.InDelta(t, expectedDuration.Seconds(), diff.Seconds(), 5, "Token expiration should be close to 72 hours")
}

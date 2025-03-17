package jwt

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Claims defines the custom JWT claims, including a Role field.
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// getJwtSecret retrieves the JWT secret directly from the environment.
func getJwtSecret() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is not set")
	}
	return []byte(secret), nil
}

// GenerateToken creates a JWT for the given username and role.
func GenerateToken(username, role string) (string, error) {
	jwtSecret, err := getJwtSecret()
	if err != nil {
		return "", err
	}

	now := time.Now()
	expirationTime := now.Add(time.Hour * time.Duration(getJwtExpireHours()))
	issuer := os.Getenv("JWT_ISSUER") // Optionally set via an environment variable

	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   username,
			Issuer:    issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken validates a token and returns its claims, a boolean indicating expiration, and an error if any.
func ValidateToken(tokenStr string) (*Claims, bool, error) {
	jwtSecret, err := getJwtSecret()
	if err != nil {
		return nil, false, err
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		// Check if the error is due to expiration.
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return claims, true, fmt.Errorf("token is expired")
			}
		}
		return nil, false, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, false, fmt.Errorf("token is not valid")
	}

	return claims, false, nil
}

func getJwtExpireHours() int {
	expHoursStr := os.Getenv("JWT_EXPIRE_HOURS")
	if expHoursStr == "" {
		return 72 // default expiration hours
	}
	expHours, err := strconv.Atoi(expHoursStr)
	if err != nil {
		return 72
	}
	return expHours
}

package db_test

import (
	"auth-service/db"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	assert.NoError(t, err)

	db.DB = mockDB // Mock the global DB variable

	err = db.DB.Ping()
	assert.NoError(t, err)
}

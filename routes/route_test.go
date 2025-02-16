package routes_test

import (
	"auth-service/routes"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestSetupRoutes(t *testing.T) {
	router := routes.SetupRoutes()
	assert.IsType(t, &mux.Router{}, router)

	tests := []struct {
		method string
		path   string
	}{
		{"POST", "/register"},
		{"POST", "/login"},
		{"POST", "/logout"},
		{"GET", "/authenticate"},
		{"GET", "/health"},
	}

	for _, tt := range tests {
		req, _ := http.NewRequest(tt.method, tt.path, nil)
		match := &mux.RouteMatch{}
		assert.True(t, router.Match(req, match), "Route %s %s not registered", tt.method, tt.path)
	}
}

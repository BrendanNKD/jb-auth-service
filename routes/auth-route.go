// routes/routes.go
package routes

import (
	"auth-service/handlers"

	"github.com/gorilla/mux"
)

func SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/register", handlers.RegisterHandler).Methods("POST")
	router.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
	router.HandleFunc("/logout", handlers.LogoutHandler).Methods("POST")
	router.Handle("/authenticate", middleware.AuthMiddleware(
			http.HandlerFunc(handlers.AuthenticateHandler)))
	router.HandleFunc("/health", handlers.HealthHandler).Methods("GET")
	return router
}

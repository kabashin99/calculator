// @title Example API
// @version 1.0
// @description This is an example API.
// @contact.name Your Name
// @contact.url http://example.com
// @contact.email your.email@example.com
// @BasePath /api/v1
package main

import (
	"calculator_app/internal/handler"
	"calculator_app/internal/middleware"
	_ "calculator_app/swagger"
	"github.com/swaggo/http-swagger" // httpSwagger
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/calculate", handler.CalculateHandler)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	loggedMux := middleware.Logging(mux)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", loggedMux); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
	middleware.CloseLogFile()
}

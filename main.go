// @title Calculator API
// @version 1.0
// @description HTTP-сервер, который обрабатывает входящие арифметические выражения и возвращает результаты вычислений
// @host localhost:8080
// @BasePath /api/v1

package main

import (
	_ "calculator_app/docs"
	"calculator_app/internal/handler"
	"calculator_app/internal/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/calculate", handler.CalculateHandler)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	//http.Handle("/swagger/", http.StripPrefix("/swagger", httpSwagger.WrapHandler))

	loggedMux := middleware.Logging(mux)

	log.Println("Сервер запущен на :8080")
	if err := http.ListenAndServe(":8080", loggedMux); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %s\n", err)
	}
	middleware.CloseLogFile()
}
